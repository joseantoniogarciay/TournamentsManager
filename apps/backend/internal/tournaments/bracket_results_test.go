package tournaments

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBracketCompletesEverySupportedFieldSize(t *testing.T) {
	for count := 2; count <= 64; count++ {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			teams := make([]string, count)
			for i := range teams {
				teams[i] = fmt.Sprintf("team-%d", i)
			}
			bracket, err := GenerateSingleElimination(teams)
			if err != nil {
				t.Fatal(err)
			}
			again, err := GenerateSingleElimination(teams)
			if err != nil || !reflect.DeepEqual(bracket, again) {
				t.Fatal("generation is not deterministic")
			}
			played, byes := 0, 0
			for _, match := range bracket.Matches {
				if match.AutoWinnerTeamID != "" {
					byes++
					continue
				}
				bracket, err = bracket.RecordResult(match.Round, match.Sequence, BracketResult{HomeScore: 2, AwayScore: 1})
				if err != nil {
					t.Fatal(err)
				}
				played++
			}
			resolved, err := bracket.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			if played != count-1 || byes != bracket.Size-count {
				t.Fatalf("played=%d byes=%d", played, byes)
			}
			if resolved[len(resolved)-1].WinnerTeamID == "" {
				t.Fatal("missing champion")
			}
		})
	}
}

func TestBracketShootoutAndCorrection(t *testing.T) {
	b, _ := GenerateSingleElimination([]string{"a", "b", "c", "d"})
	if _, err := b.RecordResult(2, 1, BracketResult{HomeScore: 1}); err != ErrBracketMatchNotReady {
		t.Fatalf("future final: %v", err)
	}
	if _, err := b.RecordResult(1, 1, BracketResult{}); err != ErrInvalidBracketResult {
		t.Fatalf("draw: %v", err)
	}
	home, away := 4, 5
	next, err := b.RecordResult(1, 1, BracketResult{HomeScore: 1, AwayScore: 1, HomePenalties: &home, AwayPenalties: &away})
	if err != nil {
		t.Fatal(err)
	}
	away = 0 // The recorded result must not alias the caller's mutable input.
	resolved, err := next.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if resolved[2].HomeTeamID != "c" || b.Matches[0].Result != nil {
		t.Fatal("shootout propagation or history changed")
	}
	next, err = next.RecordResult(1, 1, BracketResult{HomeScore: 2, AwayScore: 1})
	if err != nil {
		t.Fatal(err)
	}
	resolved, _ = next.Resolve()
	if resolved[2].HomeTeamID != "a" {
		t.Fatal("corrected winner not propagated")
	}
	next, err = next.RecordResult(1, 2, BracketResult{HomeScore: 1})
	if err != nil {
		t.Fatal(err)
	}
	next, err = next.RecordResult(2, 1, BracketResult{AwayScore: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := next.RecordResult(1, 1, BracketResult{HomeScore: 3}); err != ErrBracketResultDependency {
		t.Fatalf("descendant guard: %v", err)
	}
}

func TestBasketballBracketRejectsTiesAndShootouts(t *testing.T) {
	b, _ := GenerateSingleElimination([]string{"a", "b"})
	b.Sport = SportBasketball
	home, away := 5, 4
	if _, err := b.RecordResult(1, 1, BracketResult{HomeScore: 80, AwayScore: 80}); err != ErrInvalidBracketResult {
		t.Fatalf("tied basketball result = %v", err)
	}
	if _, err := b.RecordResult(1, 1, BracketResult{HomeScore: 80, AwayScore: 80, HomePenalties: &home, AwayPenalties: &away}); err != ErrInvalidBracketResult {
		t.Fatalf("basketball shootout = %v", err)
	}
	if _, err := b.RecordResult(1, 1, BracketResult{HomeScore: 81, AwayScore: 80}); err != nil {
		t.Fatalf("decisive basketball result = %v", err)
	}
}

func TestHandballBracketUsesSevenMeterShootoutAfterTiedFinalScore(t *testing.T) {
	b, _ := GenerateSingleElimination([]string{"a", "b"})
	b.Sport = SportHandball
	home, away := 5, 4

	updated, err := b.RecordResult(1, 1, BracketResult{HomeScore: 28, AwayScore: 28, HomePenalties: &home, AwayPenalties: &away})
	if err != nil {
		t.Fatalf("record handball shootout: %v", err)
	}
	resolved, err := updated.Resolve()
	if err != nil {
		t.Fatalf("resolve handball shootout: %v", err)
	}
	if len(resolved) != 1 || resolved[0].WinnerTeamID != "a" {
		t.Fatalf("resolved handball bracket = %#v, want team a", resolved)
	}
	if _, err := b.RecordResult(1, 1, BracketResult{HomeScore: 28, AwayScore: 28}); err != ErrInvalidBracketResult {
		t.Fatalf("tied handball result = %v, want invalid", err)
	}
}

func TestTennisBracketDerivesWinnerFromSets(t *testing.T) {
	b, _ := GenerateSingleElimination([]string{"a", "b"})
	b.Sport, b.BestOfSets = SportTennis, 5
	sets := []SetScore{{6, 4}, {4, 6}, {7, 5}, {6, 7}, {6, 2}}
	updated, err := b.RecordResult(1, 1, BracketResult{HomeScore: 3, AwayScore: 2, Sets: sets})
	if err != nil {
		t.Fatalf("record tennis result: %v", err)
	}
	sets[0].HomeScore = 0
	resolved, err := updated.Resolve()
	if err != nil || resolved[0].WinnerTeamID != "a" || resolved[0].Result.Sets[0].HomeScore != 6 {
		t.Fatalf("resolved tennis bracket = %#v, %v", resolved, err)
	}
	if _, err := b.RecordResult(1, 1, BracketResult{HomeScore: 2, AwayScore: 0, Sets: []SetScore{{6, 0}, {6, 0}}}); err != ErrInvalidBracketResult {
		t.Fatalf("unfinished best-of-five accepted: %v", err)
	}
}

func TestBracketAllowsCorrectionInIndependentBranch(t *testing.T) {
	b, _ := GenerateSingleElimination([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	var err error
	for _, position := range [][2]int{{1, 1}, {1, 2}, {2, 1}, {1, 3}} {
		b, err = b.RecordResult(position[0], position[1], BracketResult{HomeScore: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := b.RecordResult(1, 3, BracketResult{AwayScore: 1}); err != nil {
		t.Fatalf("independent branch: %v", err)
	}
}

func TestBracketRejectsMalformedSourcesAndResults(t *testing.T) {
	zero, one, negative := 0, 1, -1
	for _, result := range []BracketResult{
		{HomeScore: -1}, {}, {HomePenalties: &one},
		{HomePenalties: &one, AwayPenalties: &one},
		{HomePenalties: &negative, AwayPenalties: &zero},
		{HomeScore: 2, HomePenalties: &one, AwayPenalties: &zero},
	} {
		b, _ := GenerateSingleElimination([]string{"a", "b"})
		if _, err := b.RecordResult(1, 1, result); err != ErrInvalidBracketResult {
			t.Fatalf("invalid result accepted: %#v %v", result, err)
		}
	}
	b, _ := GenerateSingleElimination([]string{"a", "b", "c"})
	if _, err := b.RecordResult(1, 2, BracketResult{HomeScore: 1}); err != ErrBracketMatchNotReady {
		t.Fatalf("bye cannot have score: %v", err)
	}
	b.Matches[2].Home.SourceSequence = 2
	if _, err := b.Resolve(); err != ErrInvalidBracketStructure {
		t.Fatalf("wrong source: %v", err)
	}
}

func TestTableTennisPointGames(t *testing.T) {
	for _, best := range []int{3, 5, 7} {
		sets := make([]SetScore, best/2+1)
		for i := range sets {
			sets[i] = SetScore{HomeScore: 11, AwayScore: 9}
		}
		normalized, err := NormalizeSetResult(SportTableTennis, best, 0, MatchResultInput{Sets: sets})
		if err != nil || normalized.HomeScore != best/2+1 || normalized.AwayScore != 0 {
			t.Fatalf("best of %d: %#v %v", best, normalized, err)
		}
		if _, err := NormalizeSetResult(SportTableTennis, best, 0, MatchResultInput{Sets: append(sets, SetScore{HomeScore: 11, AwayScore: 0})}); err == nil {
			t.Fatal("extra game accepted")
		}
	}
	for _, test := range []struct {
		home, away int
		valid      bool
	}{{11, 0, true}, {11, 9, true}, {12, 10, true}, {15, 13, true}, {0, 11, true}, {10, 8, false}, {11, 10, false}, {12, 9, false}, {11, 11, false}, {-1, 11, false}, {32768, 32766, false}} {
		got := validSportSet(SportTableTennis, 0, 0, SetScore{HomeScore: test.home, AwayScore: test.away})
		if got != test.valid {
			t.Errorf("%d-%d valid=%v want %v", test.home, test.away, got, test.valid)
		}
	}
	if _, err := NormalizeSetResult(SportTableTennis, 3, 0, MatchResultInput{Sets: []SetScore{{HomeScore: 11, AwayScore: 0}}}); err == nil {
		t.Fatal("unfinished match accepted")
	}
	for _, best := range []int{0, 1, 2, 4, 6, 9} {
		if ValidSetFormat(SportTableTennis, best) {
			t.Errorf("invalid best of %d accepted", best)
		}
	}
}
