package leagues

import "testing"

func score(value int) *int { return &value }

func TestCalculateStandingsUsesHeadToHeadFirstWithTwoLegs(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 2,
		Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Matches: []Match{
			{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(1), AwayScore: score(0)},
			{HomeTeamID: "b", AwayTeamID: "a", State: "completed", HomeScore: score(0), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: score(0), AwayScore: score(3)},
			{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: score(2), AwayScore: score(0)},
		},
	}

	standings := calculateStandings(league)

	if standings[0].TeamID != "a" || standings[0].Position != 1 {
		t.Fatalf("first standing = %#v, want team a in position 1", standings[0])
	}
	if standings[1].TeamID != "b" || standings[1].Position != 2 {
		t.Fatalf("second standing = %#v, want team b in position 2", standings[1])
	}
}

func TestCalculateStandingsUsesGeneralGoalDifferenceFirstWithOneLeg(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 1,
		Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Matches: []Match{
			{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(1), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: score(0), AwayScore: score(3)},
			{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: score(2), AwayScore: score(0)},
		},
	}

	standings := calculateStandings(league)

	if standings[0].TeamID != "c" || standings[0].Position != 1 {
		t.Fatalf("first standing = %#v, want team c in position 1", standings[0])
	}
	if standings[1].TeamID != "b" || standings[1].Position != 2 {
		t.Fatalf("second standing = %#v, want team b in position 2", standings[1])
	}
}

func TestCalculateStandingsSharesPositionWhenEveryCriterionIsEqual(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 1,
		Teams:   []Team{{ID: "a"}, {ID: "b"}},
		Matches: []Match{{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(1), AwayScore: score(1)}},
	}

	standings := calculateStandings(league)

	if standings[0].Position != 1 || standings[1].Position != 1 {
		t.Fatalf("positions = %d, %d, want shared first position", standings[0].Position, standings[1].Position)
	}
}

func TestCalculateStandingsRanksThreeWayTieByCompleteHeadToHeadMiniTable(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 2,
		Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Matches: []Match{
			{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(1), AwayScore: score(0)},
			{HomeTeamID: "b", AwayTeamID: "a", State: "completed", HomeScore: score(0), AwayScore: score(1)},
			{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: score(2), AwayScore: score(0)},
			{HomeTeamID: "c", AwayTeamID: "b", State: "completed", HomeScore: score(0), AwayScore: score(2)},
			{HomeTeamID: "c", AwayTeamID: "a", State: "completed", HomeScore: score(3), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: score(0), AwayScore: score(3)},
		},
	}

	standings := calculateStandings(league)

	assertStandings(t, standings, []string{"c", "b", "a"}, []int{1, 2, 3})
}

func TestCalculateStandingsRanksFourWayTieByCompleteHeadToHeadMiniTable(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 2,
		Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Matches: []Match{
			{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(5), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: score(0), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "d", State: "completed", HomeScore: score(0), AwayScore: score(1)},
			{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: score(3), AwayScore: score(0)},
			{HomeTeamID: "b", AwayTeamID: "d", State: "completed", HomeScore: score(0), AwayScore: score(0)},
			{HomeTeamID: "c", AwayTeamID: "d", State: "completed", HomeScore: score(2), AwayScore: score(0)},
		},
	}

	standings := calculateStandings(league)

	assertStandings(t, standings, []string{"a", "c", "d", "b"}, []int{1, 2, 3, 4})
}

func TestCalculateStandingsSharesPositionForFourWayTieWhenEveryMetricMatches(t *testing.T) {
	league := League{
		State: "in_progress", RoundRobinLegs: 2,
		Teams: []Team{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Matches: []Match{
			{HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: score(1), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: score(0), AwayScore: score(0)},
			{HomeTeamID: "a", AwayTeamID: "d", State: "completed", HomeScore: score(0), AwayScore: score(1)},
			{HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: score(1), AwayScore: score(0)},
			{HomeTeamID: "b", AwayTeamID: "d", State: "completed", HomeScore: score(0), AwayScore: score(0)},
			{HomeTeamID: "c", AwayTeamID: "d", State: "completed", HomeScore: score(1), AwayScore: score(0)},
		},
	}

	standings := calculateStandings(league)

	assertStandings(t, standings, []string{"a", "b", "c", "d"}, []int{1, 1, 1, 1})
}

func assertStandings(t *testing.T, standings []Standing, teamIDs []string, positions []int) {
	t.Helper()
	if len(standings) != len(teamIDs) {
		t.Fatalf("standings length = %d, want %d", len(standings), len(teamIDs))
	}
	for i := range teamIDs {
		if standings[i].TeamID != teamIDs[i] || standings[i].Position != positions[i] {
			t.Errorf("standing %d = %#v, want team %q in position %d", i, standings[i], teamIDs[i], positions[i])
		}
	}
}
