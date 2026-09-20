package tournaments

import "testing"

func TestValidateMixedConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		teams     int
		input     StartInput
		wantError bool
	}{
		{"single table", 10, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 8}, false},
		{"too many qualifiers", 6, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 8}, true},
		{"balanced groups", 16, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 2, LeagueStructure: LeagueStructureGroups, GroupCount: 4, QualifiersPerGroup: 2}, false},
		{"uneven groups", 14, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureGroups, GroupCount: 4, QualifiersPerGroup: 2}, true},
		{"everyone qualifies", 8, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureGroups, GroupCount: 4, QualifiersPerGroup: 2}, true},
		{"incomplete bracket", 12, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureGroups, GroupCount: 3, QualifiersPerGroup: 2}, true},
		{"minimum field", 2, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 2}, false},
		{"maximum field", 64, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 2, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 64}, false},
		{"group count above roster", 8, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureGroups, GroupCount: 16, QualifiersPerGroup: 1}, true},
		{"qualifiers above group", 8, StartInput{Format: FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: LeagueStructureGroups, GroupCount: 2, QualifiersPerGroup: 8}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateMixedConfiguration(test.teams, test.input)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestQualifiedTeamIDsSkipsWithdrawnTeamsAndUsesTheNextPosition(t *testing.T) {
	legs := 1
	stage := Stage{ID: "stage", Type: "league", State: "in_progress", RoundRobinLegs: &legs, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 2}
	tournament := Tournament{
		Sport:  SportFootball,
		Teams:  []Team{{ID: "first", Position: 1, Withdrawn: true}, {ID: "second", Position: 2}, {ID: "third", Position: 3}},
		Stages: []Stage{stage},
		Matches: []Match{
			{StageID: "stage", HomeTeamID: "first", AwayTeamID: "second", State: "completed", HomeScore: integer(3), AwayScore: integer(0)},
			{StageID: "stage", HomeTeamID: "first", AwayTeamID: "third", State: "completed", HomeScore: integer(3), AwayScore: integer(0)},
			{StageID: "stage", HomeTeamID: "second", AwayTeamID: "third", State: "completed", HomeScore: integer(1), AwayScore: integer(0)},
		},
	}
	ids, err := QualifiedTeamIDs(tournament, stage)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "second" || ids[1] != "third" {
		t.Fatalf("qualified = %#v, want active second and third", ids)
	}
}

func TestQualifiedTeamIDsRejectsTooFewEligibleTeams(t *testing.T) {
	legs := 1
	stage := Stage{ID: "stage", Type: "league", State: "in_progress", RoundRobinLegs: &legs, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 2}
	tournament := Tournament{
		Sport:   SportFootball,
		Teams:   []Team{{ID: "withdrawn", Position: 1, Withdrawn: true}, {ID: "active", Position: 2}},
		Stages:  []Stage{stage},
		Matches: []Match{{StageID: "stage", HomeTeamID: "withdrawn", AwayTeamID: "active", State: "completed", HomeScore: integer(0), AwayScore: integer(3)}},
	}
	if _, err := QualifiedTeamIDs(tournament, stage); err != ErrTournamentStageTransitionConflict {
		t.Fatalf("error = %v, want transition conflict", err)
	}
}

func TestQualifiedTeamIDsRejectsAGroupWithoutEnoughEligibleTeams(t *testing.T) {
	legs := 1
	stage := Stage{ID: "stage", Type: "league", State: "in_progress", RoundRobinLegs: &legs, LeagueStructure: LeagueStructureGroups, GroupCount: 2, QualifiersPerGroup: 1}
	tournament := Tournament{
		Sport:  SportFootball,
		Teams:  []Team{{ID: "a", Position: 1, Withdrawn: true}, {ID: "b", Position: 2, Withdrawn: true}, {ID: "c", Position: 3}, {ID: "d", Position: 4}},
		Stages: []Stage{stage},
		StageTeams: []StageTeam{
			{StageID: "stage", TeamID: "a", GroupNumber: 1},
			{StageID: "stage", TeamID: "b", GroupNumber: 1},
			{StageID: "stage", TeamID: "c", GroupNumber: 2},
			{StageID: "stage", TeamID: "d", GroupNumber: 2},
		},
		Matches: []Match{
			{StageID: "stage", GroupNumber: 1, HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: integer(0), AwayScore: integer(3)},
			{StageID: "stage", GroupNumber: 2, HomeTeamID: "c", AwayTeamID: "d", State: "completed", HomeScore: integer(1), AwayScore: integer(0)},
		},
	}
	if _, err := QualifiedTeamIDs(tournament, stage); err != ErrTournamentStageTransitionConflict {
		t.Fatalf("error = %v, want transition conflict", err)
	}
}

func TestQualifiedTeamIDsUsesPersistedSeedAtAnExactCutoffTie(t *testing.T) {
	legs := 1
	stage := Stage{ID: "stage", Type: "league", State: "in_progress", RoundRobinLegs: &legs, LeagueStructure: LeagueStructureSingleTable, QualifierCount: 2}
	tournament := Tournament{
		Sport:  SportFootball,
		Teams:  []Team{{ID: "seed-1", Position: 1}, {ID: "seed-2", Position: 2}, {ID: "seed-3", Position: 3}},
		Stages: []Stage{stage},
		Matches: []Match{
			{StageID: "stage", HomeTeamID: "seed-1", AwayTeamID: "seed-2", State: "completed", HomeScore: integer(0), AwayScore: integer(0)},
			{StageID: "stage", HomeTeamID: "seed-1", AwayTeamID: "seed-3", State: "completed", HomeScore: integer(0), AwayScore: integer(0)},
			{StageID: "stage", HomeTeamID: "seed-2", AwayTeamID: "seed-3", State: "completed", HomeScore: integer(0), AwayScore: integer(0)},
		},
	}
	ids, err := QualifiedTeamIDs(tournament, stage)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "seed-1" || ids[1] != "seed-2" {
		t.Fatalf("qualified = %#v, want first persisted seeds", ids)
	}
}

func TestAssignGroupsUsesSerpentineSeeds(t *testing.T) {
	assignments, err := AssignGroups([]string{"1", "2", "3", "4", "5", "6", "7", "8"}, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{1, 2, 3, 4, 4, 3, 2, 1}
	for index, group := range want {
		if assignments[index].GroupNumber != group || assignments[index].SeedPosition != index+1 {
			t.Fatalf("assignment %d = %#v", index, assignments[index])
		}
	}
}

func TestQualifiedTeamIDsOrdersGroupPositionsBeforePerformance(t *testing.T) {
	legs := 1
	stage := Stage{ID: "stage", Type: "league", State: "in_progress", RoundRobinLegs: &legs, LeagueStructure: LeagueStructureGroups, GroupCount: 2, QualifiersPerGroup: 2}
	tournament := Tournament{
		Sport:      SportFootball,
		Teams:      []Team{{ID: "a", Position: 1}, {ID: "b", Position: 2}, {ID: "c", Position: 3}, {ID: "d", Position: 4}, {ID: "e", Position: 5}, {ID: "f", Position: 6}},
		Stages:     []Stage{stage},
		StageTeams: []StageTeam{{StageID: "stage", TeamID: "a", GroupNumber: 1}, {StageID: "stage", TeamID: "b", GroupNumber: 1}, {StageID: "stage", TeamID: "c", GroupNumber: 1}, {StageID: "stage", TeamID: "d", GroupNumber: 2}, {StageID: "stage", TeamID: "e", GroupNumber: 2}, {StageID: "stage", TeamID: "f", GroupNumber: 2}},
		Matches: []Match{
			{StageID: "stage", GroupNumber: 1, HomeTeamID: "a", AwayTeamID: "b", State: "completed", HomeScore: integer(2), AwayScore: integer(0)},
			{StageID: "stage", GroupNumber: 1, HomeTeamID: "a", AwayTeamID: "c", State: "completed", HomeScore: integer(1), AwayScore: integer(0)},
			{StageID: "stage", GroupNumber: 1, HomeTeamID: "b", AwayTeamID: "c", State: "completed", HomeScore: integer(1), AwayScore: integer(0)},
			{StageID: "stage", GroupNumber: 2, HomeTeamID: "d", AwayTeamID: "e", State: "completed", HomeScore: integer(5), AwayScore: integer(0)},
			{StageID: "stage", GroupNumber: 2, HomeTeamID: "d", AwayTeamID: "f", State: "completed", HomeScore: integer(5), AwayScore: integer(0)},
			{StageID: "stage", GroupNumber: 2, HomeTeamID: "e", AwayTeamID: "f", State: "completed", HomeScore: integer(4), AwayScore: integer(0)},
		},
	}
	ids, err := QualifiedTeamIDs(tournament, stage)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"d", "a", "e", "b"}
	for index := range want {
		if ids[index] != want[index] {
			t.Fatalf("qualified = %#v", ids)
		}
	}
}

func integer(value int) *int { return &value }
