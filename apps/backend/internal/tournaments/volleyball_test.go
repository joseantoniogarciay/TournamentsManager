package tournaments

import (
	"reflect"
	"testing"
)

func TestVolleyballSetSequenceAndConfiguration(t *testing.T) {
	cases := []struct {
		name       string
		sets       []SetScore
		valid      bool
		home, away int
	}{
		{"straight win", []SetScore{{25, 0}, {25, 23}, {26, 24}}, true, 3, 0},
		{"four sets", []SetScore{{25, 23}, {23, 25}, {28, 26}, {25, 10}}, true, 3, 1},
		{"deciding fifth", []SetScore{{25, 0}, {0, 25}, {25, 0}, {0, 25}, {16, 14}}, true, 3, 2},
		{"away win", []SetScore{{0, 25}, {0, 25}, {0, 25}}, true, 0, 3},
		{"unfinished", []SetScore{{25, 0}, {25, 0}}, false, 0, 0},
		{"additional set", []SetScore{{25, 0}, {25, 0}, {25, 0}, {0, 25}}, false, 0, 0},
		{"one point lead", []SetScore{{25, 24}, {25, 0}, {25, 0}}, false, 0, 0},
		{"skipped termination", []SetScore{{27, 23}, {25, 0}, {25, 0}}, false, 0, 0},
		{"short early set", []SetScore{{15, 0}, {25, 0}, {25, 0}}, false, 0, 0},
		{"long fifth", []SetScore{{25, 0}, {0, 25}, {25, 0}, {0, 25}, {25, 0}}, false, 0, 0},
		{"storage boundary", []SetScore{{32767, 32765}, {25, 0}, {25, 0}}, true, 3, 0},
		{"overflow", []SetScore{{32768, 32766}, {25, 0}, {25, 0}}, false, 0, 0},
		{"negative", []SetScore{{25, -1}, {25, 0}, {25, 0}}, false, 0, 0},
		{"aggregate only", nil, false, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			input, err := NormalizeSetResult(SportVolleyball, 5, 0, MatchResultInput{Sets: c.sets})
			if (err == nil) != c.valid || c.valid && (input.HomeScore != c.home || input.AwayScore != c.away) {
				t.Fatalf("result=%#v err=%v", input, err)
			}
		})
	}
	for _, best := range []int{0, 3, 7} {
		if ValidSetFormat(SportVolleyball, best) {
			t.Fatalf("accepted best=%d", best)
		}
	}
	penalty := 1
	if _, err := NormalizeSetResult(SportVolleyball, 5, 0, MatchResultInput{Sets: cases[0].sets, HomePenalties: &penalty}); err == nil {
		t.Fatal("shootout accepted")
	}
}

func TestVolleyballStandingPrioritiesAndExactRatios(t *testing.T) {
	cases := []struct {
		name        string
		left, right Standing
		want        int
	}{
		{"wins before points", Standing{Won: 2, Points: 4}, Standing{Won: 1, Points: 5}, 1},
		{"points before ratio", Standing{Won: 1, Points: 3, ScoreFor: 3, ScoreAgainst: 2}, Standing{Won: 1, Points: 2, ScoreFor: 3}, 1},
		{"sets before rallies", Standing{ScoreFor: 3, ScoreAgainst: 1, RallyPointsFor: 75, RallyPointsAgainst: 70}, Standing{ScoreFor: 3, ScoreAgainst: 2, RallyPointsFor: 100, RallyPointsAgainst: 1}, 1},
		{"rallies exact", Standing{ScoreFor: 3, ScoreAgainst: 2, RallyPointsFor: 10001, RallyPointsAgainst: 10000}, Standing{ScoreFor: 6, ScoreAgainst: 4, RallyPointsFor: 10000, RallyPointsAgainst: 9999}, -1},
		{"scaled exact equality", Standing{ScoreFor: 3, ScoreAgainst: 2, RallyPointsFor: 100, RallyPointsAgainst: 80}, Standing{ScoreFor: 6, ScoreAgainst: 4, RallyPointsFor: 200, RallyPointsAgainst: 160}, 0},
		{"infinite sets", Standing{ScoreFor: 3}, Standing{ScoreFor: 300, ScoreAgainst: 1}, 1},
		{"infinite rallies", Standing{ScoreFor: 3, RallyPointsFor: 75}, Standing{ScoreFor: 3, RallyPointsFor: 75, RallyPointsAgainst: 1}, 1},
		{"zero over zero", Standing{}, Standing{}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := compareVolleyballStanding(c.left, c.right)
			if got < 0 {
				got = -1
			}
			if got > 0 {
				got = 1
			}
			if got != c.want {
				t.Fatalf("comparison=%d want=%d", got, c.want)
			}
		})
	}
}

func TestVolleyballLeaguePointsTotalsAndSharedPosition(t *testing.T) {
	matches := []Match{
		{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: integer(3), AwayScore: integer(2), Sets: []SetScore{{25, 0}, {0, 25}, {25, 0}, {0, 25}, {15, 0}}},
		{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: integer(3), AwayScore: integer(1), Sets: []SetScore{{25, 0}, {0, 25}, {25, 0}, {25, 0}}},
	}
	rows := CalculateStandings(Tournament{Sport: SportVolleyball, Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Matches: matches})
	if rows[0].TeamID != "b" || rows[0].Points != 4 || rows[0].Played != 2 || rows[0].Won != 1 || rows[0].Lost != 1 || rows[0].ScoreFor != 5 || rows[0].ScoreAgainst != 4 || rows[0].RallyPointsFor != 125 || rows[0].RallyPointsAgainst != 90 || rows[1].TeamID != "a" || rows[1].Points != 2 {
		t.Fatalf("standings=%#v", rows)
	}
	empty := CalculateStandings(Tournament{Sport: SportVolleyball, Teams: []Team{{ID: "a"}, {ID: "b"}}})
	if empty[0].Position != 1 || empty[1].Position != 1 {
		t.Fatal("zero records must share position")
	}
}

func TestVolleyballGroupSeedingUsesWinsBeforePoints(t *testing.T) {
	value := Tournament{Sport: SportVolleyball, Teams: []Team{{ID: "a", Position: 1}, {ID: "b", Position: 2}}}
	rows := []Standing{{TeamID: "a", Position: 1, Won: 1, Points: 5}, {TeamID: "b", Position: 1, Won: 2, Points: 4}}
	got := orderQualifiedTeamIDs(value, rows, nil)
	if !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Fatalf("order=%#v", got)
	}
}

func TestVolleyballBracketCorrectionPropagatesWinner(t *testing.T) {
	bracket, _ := GenerateSingleElimination([]string{"a", "b", "c", "d"})
	bracket.Sport, bracket.BestOfSets = SportVolleyball, 5
	sets := []SetScore{{25, 0}, {0, 25}, {25, 0}, {0, 25}, {15, 0}}
	next, err := bracket.RecordResult(1, 1, BracketResult{HomeScore: 3, AwayScore: 2, Sets: sets})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := next.Resolve()
	if err != nil || resolved[2].HomeTeamID != "a" {
		t.Fatal("five-set winner not propagated")
	}
	sets = []SetScore{{0, 25}, {0, 25}, {0, 25}}
	next, err = next.RecordResult(1, 1, BracketResult{HomeScore: 0, AwayScore: 3, Sets: sets})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err = next.Resolve()
	if err != nil || resolved[2].HomeTeamID != "c" {
		t.Fatal("corrected winner not propagated")
	}
}
