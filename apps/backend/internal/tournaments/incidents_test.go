package tournaments

import (
	"errors"
	"testing"
)

func TestIncidentResultsAcrossSports(t *testing.T) {
	for _, sport := range []Sport{SportFootball, SportBasketball, SportHandball, SportVolleyball, SportTennis, SportPadel, SportTableTennis, SportBadminton} {
		best, points, winning := 0, 0, 3
		switch sport {
		case SportBasketball:
			winning = 20
		case SportHandball:
			winning = 10
		case SportVolleyball:
			best = 5
		case SportTennis, SportPadel, SportTableTennis:
			best = 3
			winning = 2
		case SportBadminton:
			best = 3
			points = 15
			winning = 2
		}
		for _, kind := range []ResultType{ResultNoShow, ResultRetirement} {
			for _, side := range []string{"home", "away"} {
				i := &MatchIncident{Type: kind, Side: side}
				if kind == ResultRetirement {
					if SetSport(sport) {
						i.PartialSets = []SetScore{{HomeScore: 2, AwayScore: 1}}
					} else {
						i.PartialHomeScore = integer(52)
						i.PartialAwayScore = integer(48)
					}
				}
				input, err := NormalizeIncidentResult(sport, best, points, MatchResultInput{Incident: i})
				if err != nil {
					t.Fatalf("%s %s %s: %v", sport, kind, side, err)
				}
				if side == "home" && (input.HomeScore != 0 || input.AwayScore != winning) || side == "away" && (input.HomeScore != winning || input.AwayScore != 0) {
					t.Fatalf("%s %+v", sport, input)
				}
				winner, err := DecisiveWinnerTeamID(sport, best, points, "a", "b", input)
				expected := "a"
				if side == "home" {
					expected = "b"
				}
				if err != nil || winner != expected {
					t.Fatalf("%s winner=%s err=%v", sport, winner, err)
				}
				if sport == SportVolleyball && len(input.Sets) != 3 {
					t.Fatal("administrative volleyball sets missing")
				}
				if RacketSport(sport) && len(input.Sets) != 0 {
					t.Fatal("fictional played sets")
				}
				if i.PartialHomeScore != nil {
					*i.PartialHomeScore = 999
					if *input.Incident.PartialHomeScore == 999 {
						t.Fatal("mutable partial snapshot")
					}
				}
			}
		}
	}
}

func TestRetirementPartialBoundaries(t *testing.T) {
	cases := []struct {
		sport        Sport
		best, points int
		sets         []SetScore
		valid        bool
	}{
		{SportPadel, 3, 0, []SetScore{{6, 4}, {2, 1}}, true},
		{SportPadel, 3, 0, []SetScore{{6, 6}}, true},
		{SportPadel, 3, 0, []SetScore{{5, 2}, {2, 1}}, false},
		{SportPadel, 3, 0, []SetScore{{7, 4}}, false},
		{SportPadel, 3, 0, []SetScore{{6, 4}, {6, 0}}, false},
		{SportTableTennis, 7, 0, []SetScore{{11, 9}, {10, 10}}, true},
		{SportTableTennis, 3, 0, []SetScore{{12, 0}}, false},
		{SportBadminton, 3, 15, []SetScore{{20, 20}}, true},
		{SportBadminton, 3, 15, []SetScore{{22, 20}}, false},
		{SportBadminton, 3, 21, []SetScore{{29, 29}}, true},
		{SportBadminton, 3, 21, []SetScore{{31, 29}}, false},
		{SportVolleyball, 5, 0, []SetScore{{25, 20}, {20, 25}, {25, 20}, {20, 25}, {14, 14}}, true},
		{SportVolleyball, 5, 0, []SetScore{{26, 0}}, false},
		{SportTennis, 5, 0, []SetScore{{-1, 2}}, false},
	}
	for _, c := range cases {
		_, err := NormalizeIncidentResult(c.sport, c.best, c.points, MatchResultInput{Incident: &MatchIncident{Type: ResultRetirement, Side: "home", PartialSets: c.sets}})
		if (err == nil) != c.valid {
			t.Errorf("%s %+v err=%v expected=%v", c.sport, c.sets, err, c.valid)
		}
	}
	for _, i := range []*MatchIncident{
		{Type: ResultNoShow, Side: "both"}, {Type: "unknown", Side: "home"}, {Type: ResultNoShow, Side: "home", PartialHomeScore: integer(0), PartialAwayScore: integer(0)},
		{Type: ResultRetirement, Side: "home", PartialHomeScore: integer(0)},
		{Type: ResultRetirement, Side: "home", PartialHomeScore: integer(-1), PartialAwayScore: integer(0)},
		{Type: ResultRetirement, Side: "home", PartialSets: []SetScore{{1, 0}}},
	} {
		if _, err := NormalizeIncidentResult(SportFootball, 0, 0, MatchResultInput{Incident: i}); err == nil {
			t.Errorf("invalid incident accepted: %+v", i)
		}
	}
}

func TestIncidentBracketCorrectionsAndSnapshot(t *testing.T) {
	b, err := GenerateSingleElimination([]string{"a", "b", "c", "d"})
	if err != nil {
		t.Fatal(err)
	}
	b.Sport = SportPadel
	b.BestOfSets = 3
	result, err := NormalizeIncidentResult(b.Sport, 3, 0, MatchResultInput{Incident: &MatchIncident{Type: ResultRetirement, Side: "home", PartialSets: []SetScore{{6, 4}, {2, 1}}}})
	if err != nil {
		t.Fatal(err)
	}
	b, err = b.RecordResult(1, 1, BracketResult(result))
	if err != nil {
		t.Fatal(err)
	}
	result.Incident.PartialSets[0].HomeScore = 99
	if b.Matches[0].Result.Incident.PartialSets[0].HomeScore != 6 {
		t.Fatal("mutated history")
	}
	played := BracketResult{HomeScore: 2, AwayScore: 0, Sets: []SetScore{{6, 0}, {6, 1}}}
	corrected, err := b.RecordResult(1, 1, played)
	if err != nil {
		t.Fatal(err)
	}
	if b.Matches[0].Result.Incident == nil || corrected.Matches[0].Result.Incident != nil {
		t.Fatal("incident transition erased previous snapshot")
	}
	corrected, err = corrected.RecordResult(1, 2, played)
	if err != nil {
		t.Fatal(err)
	}
	corrected, err = corrected.RecordResult(2, 1, played)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = corrected.RecordResult(1, 1, *b.Matches[0].Result); !errors.Is(err, ErrBracketResultDependency) {
		t.Fatalf("dependency: %v", err)
	}
}
