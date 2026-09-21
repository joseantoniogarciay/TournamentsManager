package tournaments

import (
	"errors"
	"sort"
)

const (
	// FormatLeagueThenSingleElimination composes a qualifying league and bracket.
	FormatLeagueThenSingleElimination = "league_then_single_elimination"
	// StageTypeQualificationTieBreak resolves an exact qualification cutoff tie.
	StageTypeQualificationTieBreak = "qualification_tiebreak"
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
	// ErrQualificationTieBreakRequired prevents a hidden seed-order qualification.
	ErrQualificationTieBreakRequired = errors.New("qualification tiebreak required")
)

// QualificationTieBreakPlan describes one tied block crossing a qualification cutoff.
type QualificationTieBreakPlan struct {
	PoolNumber        int
	SourceGroupNumber int
	QualifierCount    int
	Standings         []Standing
}

// QualificationPlan separates already resolved qualifiers from tied cutoff blocks.
type QualificationPlan struct {
	Direct []Standing
	Pools  []QualificationTieBreakPlan
}

// QualificationTieBreakResolution describes the latest durable state of one pool.
type QualificationTieBreakResolution struct {
	QualifiedTeamIDs        []string
	PendingTeamIDs          []string
	RemainingQualifierCount int
	Complete                bool
	NeedsNextCycle          bool
}

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

// PlanQualification finds every exact tie that crosses a qualification cutoff.
func PlanQualification(tournament Tournament, stage Stage) (QualificationPlan, error) {
	if stage.Type != "league" || (stage.State != "in_progress" && stage.State != "completed") {
		return QualificationPlan{}, ErrTournamentStageTransitionConflict
	}
	for _, match := range tournament.Matches {
		if match.StageID == stage.ID && match.State != "completed" {
			return QualificationPlan{}, ErrTournamentStageTransitionConflict
		}
	}
	plan := QualificationPlan{Direct: []Standing{}, Pools: []QualificationTieBreakPlan{}}
	appendTable := func(standings []Standing, qualifierCount, sourceGroup int) error {
		if len(standings) < qualifierCount {
			return ErrTournamentStageTransitionConflict
		}
		direct, tied, tiedQualifierCount := qualificationCut(standings, qualifierCount)
		plan.Direct = append(plan.Direct, direct...)
		if len(tied) > 0 {
			plan.Pools = append(plan.Pools, QualificationTieBreakPlan{
				PoolNumber:        len(plan.Pools) + 1,
				SourceGroupNumber: sourceGroup,
				QualifierCount:    tiedQualifierCount,
				Standings:         tied,
			})
		}
		return nil
	}
	if stage.LeagueStructure == LeagueStructureSingleTable {
		standings := eligibleStandings(tournament, calculateStageStandings(tournament, stage, 0))
		if !validFullBracketSize(stage.QualifierCount) {
			return QualificationPlan{}, ErrTournamentStageTransitionConflict
		}
		if err := appendTable(standings, stage.QualifierCount, 0); err != nil {
			return QualificationPlan{}, err
		}
		return plan, nil
	}
	if stage.LeagueStructure != LeagueStructureGroups || stage.GroupCount < 2 || stage.QualifiersPerGroup < 1 {
		return QualificationPlan{}, ErrTournamentStageTransitionConflict
	}
	for group := 1; group <= stage.GroupCount; group++ {
		standings := eligibleStandings(tournament, calculateStageStandings(tournament, stage, group))
		if err := appendTable(standings, stage.QualifiersPerGroup, group); err != nil {
			return QualificationPlan{}, err
		}
	}
	if !validFullBracketSize(len(plan.Direct) + plannedTieBreakQualifierCount(plan.Pools)) {
		return QualificationPlan{}, ErrTournamentStageTransitionConflict
	}
	return plan, nil
}

func qualificationCut(standings []Standing, qualifierCount int) ([]Standing, []Standing, int) {
	if qualifierCount <= 0 || qualifierCount > len(standings) {
		return nil, nil, 0
	}
	boundaryPosition := standings[qualifierCount-1].Position
	end := qualifierCount
	for end < len(standings) && standings[end].Position == boundaryPosition {
		end++
	}
	if end == qualifierCount {
		return standings[:qualifierCount], nil, 0
	}
	start := qualifierCount - 1
	for start > 0 && standings[start-1].Position == boundaryPosition {
		start--
	}
	return standings[:start], standings[start:end], qualifierCount - start
}

func plannedTieBreakQualifierCount(pools []QualificationTieBreakPlan) int {
	total := 0
	for _, pool := range pools {
		total += pool.QualifierCount
	}
	return total
}

// QualifiedTeamIDs returns a field only when no cutoff tiebreak is necessary.
func QualifiedTeamIDs(tournament Tournament, stage Stage) ([]string, error) {
	plan, err := PlanQualification(tournament, stage)
	if err != nil {
		return nil, err
	}
	if len(plan.Pools) > 0 {
		return nil, ErrQualificationTieBreakRequired
	}
	return orderQualifiedTeamIDs(tournament, plan.Direct, nil), nil
}

// ResolvedQualifiedTeamIDs combines the frozen league with every completed tiebreak pool.
func ResolvedQualifiedTeamIDs(tournament Tournament, leagueStage, tieBreakStage Stage) ([]string, error) {
	plan, err := PlanQualification(tournament, leagueStage)
	if err != nil {
		return nil, err
	}
	if len(plan.Pools) == 0 {
		return orderQualifiedTeamIDs(tournament, plan.Direct, nil), nil
	}
	poolByNumber := map[int]QualificationTieBreakPool{}
	for _, pool := range tournament.TieBreakPools {
		if pool.StageID == tieBreakStage.ID {
			poolByNumber[pool.PoolNumber] = pool
		}
	}
	poolRanks := map[string][2]int{}
	candidates := append([]Standing{}, plan.Direct...)
	for _, planned := range plan.Pools {
		pool, exists := poolByNumber[planned.PoolNumber]
		if !exists || pool.State != "completed" || pool.SourceGroupNumber != planned.SourceGroupNumber || pool.QualifierCount != planned.QualifierCount {
			return nil, ErrTournamentStageTransitionConflict
		}
		resolution, err := ResolveQualificationTieBreak(tournament, pool)
		if err != nil || !resolution.Complete || len(resolution.QualifiedTeamIDs) != planned.QualifierCount {
			return nil, ErrTournamentStageTransitionConflict
		}
		standingByTeam := map[string]Standing{}
		for _, standing := range planned.Standings {
			standingByTeam[standing.TeamID] = standing
		}
		for rank, teamID := range resolution.QualifiedTeamIDs {
			standing, exists := standingByTeam[teamID]
			if !exists {
				return nil, ErrTournamentStageTransitionConflict
			}
			candidates = append(candidates, standing)
			poolRanks[teamID] = [2]int{planned.PoolNumber, rank + 1}
		}
	}
	return orderQualifiedTeamIDs(tournament, candidates, poolRanks), nil
}

func orderQualifiedTeamIDs(tournament Tournament, candidates []Standing, poolRanks map[string][2]int) []string {
	seedByTeam := map[string]int{}
	for _, team := range tournament.Teams {
		seedByTeam[team.ID] = team.Position
	}
	ordered := append([]Standing{}, candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		if left.Position != right.Position {
			return left.Position < right.Position
		}
		if comparison := compareStanding(left, right); comparison != 0 {
			return comparison > 0
		}
		leftPool, leftHasPool := poolRanks[left.TeamID]
		rightPool, rightHasPool := poolRanks[right.TeamID]
		if leftHasPool && rightHasPool && leftPool[0] == rightPool[0] {
			return leftPool[1] < rightPool[1]
		}
		return seedByTeam[left.TeamID] < seedByTeam[right.TeamID]
	})
	return standingTeamIDs(ordered)
}

// ResolveQualificationTieBreak replays every cycle and returns either a final
// qualifier order or the exact subgroup that needs another round.
func ResolveQualificationTieBreak(tournament Tournament, pool QualificationTieBreakPool) (QualificationTieBreakResolution, error) {
	if pool.StageID == "" || pool.PoolNumber < 1 || pool.QualifierCount < 1 || pool.CurrentCycle < 1 {
		return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
	}
	expectedTeams := tieBreakStageTeamIDs(tournament, pool)
	if len(expectedTeams) <= pool.QualifierCount {
		return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
	}
	remaining := pool.QualifierCount
	qualified := []string{}
	for cycle := 1; cycle <= pool.CurrentCycle; cycle++ {
		standings, complete, err := calculateTieBreakStandings(tournament, pool, cycle)
		if err != nil || !sameTeamSet(expectedTeams, standingTeamIDs(standings)) {
			return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
		}
		if !complete {
			if cycle != pool.CurrentCycle || pool.State == "completed" {
				return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
			}
			return QualificationTieBreakResolution{
				QualifiedTeamIDs:        qualified,
				PendingTeamIDs:          standingTeamIDs(standings),
				RemainingQualifierCount: remaining,
			}, nil
		}
		selected, tied, tiedQualifierCount := qualificationCut(standings, remaining)
		qualified = append(qualified, standingTeamIDs(selected)...)
		if len(tied) == 0 {
			if cycle != pool.CurrentCycle {
				return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
			}
			return QualificationTieBreakResolution{QualifiedTeamIDs: qualified, Complete: true}, nil
		}
		expectedTeams = standingTeamIDs(tied)
		remaining = tiedQualifierCount
		if cycle == pool.CurrentCycle {
			if pool.State == "completed" {
				return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
			}
			return QualificationTieBreakResolution{
				QualifiedTeamIDs:        qualified,
				PendingTeamIDs:          expectedTeams,
				RemainingQualifierCount: remaining,
				NeedsNextCycle:          true,
			}, nil
		}
	}
	return QualificationTieBreakResolution{}, ErrTournamentStageTransitionConflict
}

func calculateTieBreakStandings(tournament Tournament, pool QualificationTieBreakPool, cycle int) ([]Standing, bool, error) {
	matches := []Match{}
	participants := map[string]bool{}
	pairs := map[[2]string]bool{}
	complete := true
	for _, match := range tournament.Matches {
		if match.StageID != pool.StageID || match.GroupNumber != pool.PoolNumber || match.RoundNumber != cycle {
			continue
		}
		if match.HomeTeamID == "" || match.AwayTeamID == "" || match.HomeTeamID == match.AwayTeamID {
			return nil, false, ErrTournamentStageTransitionConflict
		}
		pair := [2]string{match.HomeTeamID, match.AwayTeamID}
		if pair[0] > pair[1] {
			pair[0], pair[1] = pair[1], pair[0]
		}
		if pairs[pair] {
			return nil, false, ErrTournamentStageTransitionConflict
		}
		pairs[pair] = true
		participants[match.HomeTeamID] = true
		participants[match.AwayTeamID] = true
		complete = complete && match.State == "completed"
		matches = append(matches, match)
	}
	if len(participants) < 2 || len(matches) != len(participants)*(len(participants)-1)/2 {
		return nil, false, ErrTournamentStageTransitionConflict
	}
	order := tieBreakStageTeamIDs(tournament, pool)
	standingsByTeam := map[string]*Standing{}
	for _, teamID := range order {
		if participants[teamID] {
			standingsByTeam[teamID] = &Standing{StageID: pool.StageID, GroupNumber: pool.PoolNumber, TeamID: teamID}
		}
	}
	if len(standingsByTeam) != len(participants) {
		return nil, false, ErrTournamentStageTransitionConflict
	}
	for _, match := range matches {
		if match.State != "completed" {
			continue
		}
		if match.HomeScore == nil || match.AwayScore == nil || (match.WinnerTeamID != match.HomeTeamID && match.WinnerTeamID != match.AwayTeamID) {
			return nil, false, ErrTournamentStageTransitionConflict
		}
		home, away := standingsByTeam[match.HomeTeamID], standingsByTeam[match.AwayTeamID]
		home.Played, away.Played = home.Played+1, away.Played+1
		home.ScoreFor, home.ScoreAgainst = home.ScoreFor+*match.HomeScore, home.ScoreAgainst+*match.AwayScore
		away.ScoreFor, away.ScoreAgainst = away.ScoreFor+*match.AwayScore, away.ScoreAgainst+*match.HomeScore
		if match.WinnerTeamID == match.HomeTeamID {
			home.Won, away.Lost = home.Won+1, away.Lost+1
		} else {
			away.Won, home.Lost = away.Won+1, home.Lost+1
		}
	}
	standings := make([]Standing, 0, len(standingsByTeam))
	for _, teamID := range order {
		standing, exists := standingsByTeam[teamID]
		if !exists {
			continue
		}
		standing.ScoreDifference = standing.ScoreFor - standing.ScoreAgainst
		standing.Points = standing.Won
		standings = append(standings, *standing)
	}
	sort.SliceStable(standings, func(i, j int) bool {
		left, right := standings[i], standings[j]
		if left.Won != right.Won {
			return left.Won > right.Won
		}
		if left.ScoreDifference != right.ScoreDifference {
			return left.ScoreDifference > right.ScoreDifference
		}
		return left.ScoreFor > right.ScoreFor
	})
	position := 1
	for index := range standings {
		if index > 0 && (standings[index-1].Won != standings[index].Won || standings[index-1].ScoreDifference != standings[index].ScoreDifference || standings[index-1].ScoreFor != standings[index].ScoreFor) {
			position = index + 1
		}
		standings[index].Position = position
	}
	return standings, complete, nil
}

func tieBreakStageTeamIDs(tournament Tournament, pool QualificationTieBreakPool) []string {
	assignments := []StageTeam{}
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == pool.StageID && assignment.GroupNumber == pool.PoolNumber {
			assignments = append(assignments, assignment)
		}
	}
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].SeedPosition < assignments[j].SeedPosition })
	ids := make([]string, len(assignments))
	for index, assignment := range assignments {
		ids[index] = assignment.TeamID
	}
	return ids
}

func sameTeamSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]bool{}
	for _, teamID := range left {
		seen[teamID] = true
	}
	for _, teamID := range right {
		if !seen[teamID] {
			return false
		}
	}
	return true
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
