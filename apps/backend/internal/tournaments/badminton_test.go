package tournaments

import (
	"errors"
	"testing"
)

func TestBadmintonPointProfiles(t *testing.T) {
	for _, c := range []struct {
		points, home, away int
		valid              bool
	}{
		{21, 21, 0, true}, {21, 21, 19, true}, {21, 22, 20, true}, {21, 29, 27, true},
		{21, 30, 28, true}, {21, 30, 29, true}, {21, 21, 20, false}, {21, 22, 19, false},
		{21, 31, 29, false}, {21, 30, 27, false}, {21, 20, 18, false}, {21, 21, 21, false},
		{15, 15, 0, true}, {15, 15, 13, true}, {15, 16, 14, true}, {15, 20, 18, true},
		{15, 21, 19, true}, {15, 21, 20, true}, {15, 15, 14, false}, {15, 16, 13, false},
		{15, 22, 20, false}, {15, 21, 18, false}, {15, 14, 12, false}, {15, 15, 15, false},
		{21, 21, -1, false}, {15, -1, 15, false},
	} {
		for _, set := range []SetScore{{HomeScore: c.home, AwayScore: c.away}, {HomeScore: c.away, AwayScore: c.home}} {
			if got := validSportSet(SportBadminton, c.points, 0, set); got != c.valid {
				t.Errorf("points=%d set=%+v valid=%v want=%v", c.points, set, got, c.valid)
			}
		}
	}
	for _, points := range []int{15, 21} {
		sets := []SetScore{{HomeScore: points, AwayScore: 0}, {HomeScore: 0, AwayScore: points}, {HomeScore: points, AwayScore: points - 2}}
		input, err := NormalizeSetResult(SportBadminton, 3, points, MatchResultInput{HomeScore: 999, AwayScore: 999, Sets: sets})
		if err != nil || input.HomeScore != 2 || input.AwayScore != 1 {
			t.Fatalf("aggregate=%+v err=%v", input, err)
		}
		for _, invalid := range [][]SetScore{sets[:1], append(append([]SetScore{}, sets...), sets[0]), {sets[0], sets[0], sets[0]}} {
			if _, err := NormalizeSetResult(SportBadminton, 3, points, MatchResultInput{Sets: invalid}); !errors.Is(err, ErrInvalidBracketResult) {
				t.Errorf("accepted incomplete or extra games: %+v", invalid)
			}
		}
		bracket, err := GenerateSingleElimination([]string{"A", "B", "C", "D"})
		if err != nil {
			t.Fatal(err)
		}
		bracket.Sport, bracket.BestOfSets, bracket.PointsPerGame = SportBadminton, 3, points
		next, err := bracket.RecordResult(1, 1, BracketResult(input))
		if err != nil || next.PointsPerGame != points {
			t.Fatalf("profile lost: %+v %v", next, err)
		}
		resolved, err := next.Resolve()
		if err != nil || resolved[2].HomeTeamID != "A" {
			t.Fatalf("winner not propagated: %+v %v", resolved, err)
		}
	}
}

func TestBadmintonCreationRequiresClosedConfiguration(t *testing.T) {
	for _, points := range []int{0, 11, 15, 21, 30} {
		for _, best := range []int{0, 3, 5, 7} {
			input := CreateInput{Name: "Badminton", Sport: SportBadminton, BestOfSets: best, PointsPerGame: points, Teams: []TeamInput{{Name: "A"}}}
			if validCreateInput(input) != (best == 3 && (points == 15 || points == 21)) {
				t.Errorf("best=%d points=%d", best, points)
			}
		}
	}
	if validCreateInput(CreateInput{Name: "Tennis", Sport: SportTennis, BestOfSets: 3, PointsPerGame: 21, Teams: []TeamInput{{Name: "A"}}}) {
		t.Fatal("badminton configuration accepted for tennis")
	}
	for _, points := range []int{0, 11, 30} {
		if _, err := NormalizeSetResult(SportBadminton, 3, points, MatchResultInput{Sets: []SetScore{{HomeScore: 21, AwayScore: 0}, {HomeScore: 21, AwayScore: 0}}}); !errors.Is(err, ErrInvalidBracketResult) {
			t.Fatalf("points=%d: %v", points, err)
		}
	}
}
