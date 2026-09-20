package tournaments

import (
	"errors"
	"sort"
)

const (
	// FormatLeagueThenSingleElimination composes a qualifying league and bracket.
	FormatLeagueThenSingleElimination = "league_then_single_elimination"
	// LeagueStructureSingleTable uses one table containing every team.
	LeagueStructureSingleTable = "single_table"
	// LeagueStructureGroups divides the field into balanced groups.
	LeagueStructureGroups = "groups"
)

var (
	// ErrInvalidMixedConfiguration rejects a roster that cannot satisfy its chosen format.
	ErrInvalidMixedConfiguration = errors.New("invalid mixed tournament configuration")
	// ErrTournamentStageTransitionConflict rejects an unavailable or incomplete next stage.
	ErrTournamentStageTransitionConflict = errors.New("tournament stage cannot advance")
)

func validStartInput(input StartInput) bool {
	switch input.Format {
	case "league":
		return (input.RoundRobinLegs == 1 || input.RoundRobinLegs == 2) && noMixedFields(input)
	case "single_elimination":
		return input.RoundRobinLegs == 0 && noMixedFields(input)
	case FormatLeagueThenSingleElimination:
		if input.RoundRobinLegs != 1 && input.RoundRobinLegs != 2 {
			return false
		}
		switch input.LeagueStructure {
		case LeagueStructureSingleTable:
			return input.GroupCount == 0 && input.QualifiersPerGroup == 0 && validFullBracketSize(input.QualifierCount)
		case LeagueStructureGroups:
			return input.QualifierCount == 0 && input.GroupCount >= 2 && input.QualifiersPerGroup >= 1 && validFullBracketSize(input.GroupCount*input.QualifiersPerGroup)
		default:
			return false
		}
	default:
		return false
	}
}

func noMixedFields(input StartInput) bool {
	return input.LeagueStructure == "" && input.QualifierCount == 0 && input.GroupCount == 0 && input.QualifiersPerGroup == 0
}

func validFullBracketSize(size int) bool {
	return size >= minimumEntrants && size <= maximumEntrants && size&(size-1) == 0
}

// ValidateMixedConfiguration checks roster-dependent rules before any match is created.
func ValidateMixedConfiguration(teamCount int, input StartInput) error {
	if !validStartInput(input) || input.Format != FormatLeagueThenSingleElimination || teamCount < minimumEntrants || teamCount > maximumEntrants {
		return ErrInvalidMixedConfiguration
	}
	if input.LeagueStructure == LeagueStructureSingleTable {
		if input.QualifierCount > teamCount {
			return ErrInvalidMixedConfiguration
		}
		return nil
	}
	if teamCount%input.GroupCount != 0 {
		return ErrInvalidMixedConfiguration
	}
	groupSize := teamCount / input.GroupCount
	if input.QualifiersPerGroup >= groupSize {
		return ErrInvalidMixedConfiguration
	}
	return nil
}

// AssignGroups applies the accepted serpentine distribution to seed-ordered teams.
func AssignGroups(teamIDs []string, groupCount int) ([]StageTeam, error) {
	if groupCount < 2 || len(teamIDs) == 0 || len(teamIDs)%groupCount != 0 {
		return nil, ErrInvalidMixedConfiguration
	}
	assignments := make([]StageTeam, len(teamIDs))
	seen := make(map[string]bool, len(teamIDs))
	for index, teamID := range teamIDs {
		if teamID == "" || seen[teamID] {
			return nil, ErrInvalidMixedConfiguration
		}
		seen[teamID] = true
		row, offset := index/groupCount, index%groupCount
		group := offset + 1
		if row%2 == 1 {
			group = groupCount - offset
		}
		assignments[index] = StageTeam{TeamID: teamID, SeedPosition: index + 1, GroupNumber: group}
	}
	return assignments, nil
}

// CalculateTournamentStandings preserves every league-stage table, including groups.
func CalculateTournamentStandings(tournament Tournament) []Standing {
	result := []Standing{}
	for _, stage := range tournament.Stages {
		if stage.Type != "league" || stage.State == "pending" || stage.State == "cancelled" {
			continue
		}
		if stage.LeagueStructure == LeagueStructureGroups {
			for group := 1; group <= stage.GroupCount; group++ {
				result = append(result, calculateStageStandings(tournament, stage, group)...)
			}
			continue
		}
		result = append(result, calculateStageStandings(tournament, stage, 0)...)
	}
	if len(tournament.Stages) == 0 && tournament.Format != "single_elimination" {
		return calculateStandings(tournament)
	}
	return result
}

func calculateStageStandings(tournament Tournament, stage Stage, groupNumber int) []Standing {
	teamIDs := map[string]bool{}
	if groupNumber == 0 {
		for _, team := range tournament.Teams {
			teamIDs[team.ID] = true
		}
	} else {
		for _, assignment := range tournament.StageTeams {
			if assignment.StageID == stage.ID && assignment.GroupNumber == groupNumber {
				teamIDs[assignment.TeamID] = true
			}
		}
	}
	projection := Tournament{Sport: tournament.Sport, Teams: []Team{}, Matches: []Match{}}
	if stage.RoundRobinLegs != nil {
		projection.RoundRobinLegs = *stage.RoundRobinLegs
	}
	for _, team := range tournament.Teams {
		if teamIDs[team.ID] {
			projection.Teams = append(projection.Teams, team)
		}
	}
	for _, match := range tournament.Matches {
		if match.StageID == stage.ID && (groupNumber == 0 || match.GroupNumber == groupNumber) {
			projection.Matches = append(projection.Matches, match)
		}
	}
	standings := calculateStandings(projection)
	for index := range standings {
		standings[index].StageID = stage.ID
		standings[index].GroupNumber = groupNumber
	}
	return standings
}

// QualifiedTeamIDs returns the complete, deterministically seeded knockout field.
func QualifiedTeamIDs(tournament Tournament, stage Stage) ([]string, error) {
	if stage.Type != "league" || stage.State != "in_progress" {
		return nil, ErrTournamentStageTransitionConflict
	}
	for _, match := range tournament.Matches {
		if match.StageID == stage.ID && match.State != "completed" {
			return nil, ErrTournamentStageTransitionConflict
		}
	}
	seedByTeam := map[string]int{}
	for _, team := range tournament.Teams {
		seedByTeam[team.ID] = team.Position
	}
	if stage.LeagueStructure == LeagueStructureSingleTable {
		standings := eligibleStandings(tournament, calculateStageStandings(tournament, stage, 0))
		if len(standings) < stage.QualifierCount || !validFullBracketSize(stage.QualifierCount) {
			return nil, ErrTournamentStageTransitionConflict
		}
		return standingTeamIDs(standings[:stage.QualifierCount]), nil
	}
	if stage.LeagueStructure != LeagueStructureGroups || stage.GroupCount < 2 || stage.QualifiersPerGroup < 1 {
		return nil, ErrTournamentStageTransitionConflict
	}
	qualified := []Standing{}
	for group := 1; group <= stage.GroupCount; group++ {
		standings := eligibleStandings(tournament, calculateStageStandings(tournament, stage, group))
		if len(standings) < stage.QualifiersPerGroup {
			return nil, ErrTournamentStageTransitionConflict
		}
		qualified = append(qualified, standings[:stage.QualifiersPerGroup]...)
	}
	if !validFullBracketSize(len(qualified)) {
		return nil, ErrTournamentStageTransitionConflict
	}
	sort.SliceStable(qualified, func(i, j int) bool {
		left, right := qualified[i], qualified[j]
		if left.Position != right.Position {
			return left.Position < right.Position
		}
		if comparison := compareStanding(left, right); comparison != 0 {
			return comparison > 0
		}
		return seedByTeam[left.TeamID] < seedByTeam[right.TeamID]
	})
	return standingTeamIDs(qualified), nil
}

func eligibleStandings(tournament Tournament, standings []Standing) []Standing {
	withdrawn := make(map[string]bool, len(tournament.Teams))
	for _, team := range tournament.Teams {
		withdrawn[team.ID] = team.Withdrawn
	}
	eligible := make([]Standing, 0, len(standings))
	for _, standing := range standings {
		if !withdrawn[standing.TeamID] {
			eligible = append(eligible, standing)
		}
	}
	return eligible
}

func standingTeamIDs(standings []Standing) []string {
	ids := make([]string, len(standings))
	for index, standing := range standings {
		ids[index] = standing.TeamID
	}
	return ids
}
