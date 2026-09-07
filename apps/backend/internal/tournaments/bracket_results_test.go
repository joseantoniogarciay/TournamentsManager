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
