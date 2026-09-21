package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationMixedTournamentFreezesQualifiersAndStartsSeededBracket(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-owner@example.com", "mixed_owner", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := make([]tournaments.TeamInput, 8)
	for index := range teams {
		teams[index] = tournaments.TeamInput{Name: fmt.Sprintf("Team %d", index+1)}
	}
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Mixed", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	tournament, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{
		Format:          tournaments.FormatLeagueThenSingleElimination,
		RoundRobinLegs:  1,
		LeagueStructure: tournaments.LeagueStructureSingleTable,
		QualifierCount:  4,
	})
	if err != nil {
		t.Fatal(err)
	}
	leagueStage := mixedStageByType(t, tournament, "league")
	tieBreakStage := mixedStageByType(t, tournament, tournaments.StageTypeQualificationTieBreak)
	eliminationStage := mixedStageByType(t, tournament, "single_elimination")
	if len(tournament.Stages) != 3 || leagueStage.State != "in_progress" || tieBreakStage.State != "pending" || eliminationStage.State != "pending" {
		t.Fatalf("stages after start = %#v", tournament.Stages)
	}
	leagueMatchIDs := make([]string, 0, len(tournament.Matches))
	for _, match := range tournament.Matches {
		if match.StageID != leagueStage.ID {
			t.Fatalf("unexpected match before transition: %#v", match)
		}
		leagueMatchIDs = append(leagueMatchIDs, match.ID)
		if _, err = recordHigherSeedWin(ctx, service, owner, tournament, match); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = service.Complete(ctx, owner, tournament.ID); !errors.Is(err, tournaments.ErrTournamentCompletionConflict) {
		t.Fatalf("completion before elimination error = %v", err)
	}
	tournament, err = service.StartElimination(ctx, owner, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	leagueStage = mixedStageByType(t, tournament, "league")
	tieBreakStage = mixedStageByType(t, tournament, tournaments.StageTypeQualificationTieBreak)
	eliminationStage = mixedStageByType(t, tournament, "single_elimination")
	if leagueStage.State != "completed" || tieBreakStage.State != "completed" || eliminationStage.State != "in_progress" {
		t.Fatalf("stages after transition = %#v", tournament.Stages)
	}
	seeded := []string{}
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == eliminationStage.ID {
			seeded = append(seeded, assignment.TeamID)
		}
	}
	if len(seeded) != 4 {
		t.Fatalf("knockout seeds = %#v", seeded)
	}
	firstRound := []tournaments.Match{}
	for _, match := range tournament.Matches {
		if match.StageID == eliminationStage.ID && match.RoundNumber == 1 {
			firstRound = append(firstRound, match)
			if match.State == "bye" {
				t.Fatalf("mixed bracket contains bye: %#v", match)
			}
		}
	}
	if len(firstRound) != 2 || firstRound[0].HomeTeamID != seeded[0] || firstRound[0].AwayTeamID != seeded[3] || firstRound[1].HomeTeamID != seeded[1] || firstRound[1].AwayTeamID != seeded[2] {
		t.Fatalf("first round = %#v; seeds = %#v", firstRound, seeded)
	}
	if _, err = service.RecordResult(ctx, owner, tournament.ID, leagueMatchIDs[0], tournaments.MatchResultInput{HomeScore: 2}); !errors.Is(err, tournaments.ErrMatchResultConflict) {
		t.Fatalf("frozen league correction error = %v", err)
	}
}

func TestIntegrationMixedTournamentRepeatsAnUnresolvedQualificationTieBreak(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-tiebreak@example.com", "mixed_tiebreak", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, owner, tournaments.CreateInput{
		Name: "Mixed repeated tiebreak", Sport: tournaments.SportFootball,
		Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	value, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{
		Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1,
		LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	leagueStage := mixedStageByType(t, value, "league")
	leagueMatchID := ""
	for _, match := range value.Matches {
		if match.StageID != leagueStage.ID {
			continue
		}
		leagueMatchID = match.ID
		value, err = service.RecordResult(ctx, owner, value.ID, match.ID, tournaments.MatchResultInput{})
		if err != nil {
			t.Fatal(err)
		}
	}
	value, err = service.StartElimination(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	tieBreakStage := mixedStageByType(t, value, tournaments.StageTypeQualificationTieBreak)
	if tieBreakStage.State != "in_progress" || len(value.TieBreakPools) != 1 || value.TieBreakPools[0].QualifierCount != 2 {
		t.Fatalf("initial tiebreak = stage %#v, pools %#v", tieBreakStage, value.TieBreakPools)
	}
	if _, err = service.RecordResult(ctx, owner, value.ID, leagueMatchID, tournaments.MatchResultInput{HomeScore: 1}); !errors.Is(err, tournaments.ErrMatchResultConflict) {
		t.Fatalf("frozen league correction error = %v", err)
	}

	teamIDs := map[string]string{}
	for _, team := range value.Teams {
		teamIDs[team.Name] = team.ID
	}
	cycleOneWinners := map[[2]string]string{
		orderedTeamPair(teamIDs["A"], teamIDs["B"]): teamIDs["A"],
		orderedTeamPair(teamIDs["A"], teamIDs["C"]): teamIDs["C"],
		orderedTeamPair(teamIDs["B"], teamIDs["C"]): teamIDs["B"],
	}
	cycleOne := tiebreakMatchesForCycle(value, tieBreakStage.ID, 1)
	if len(cycleOne) != 3 {
		t.Fatalf("cycle one matches = %#v", cycleOne)
	}
	if _, err = service.RecordResult(ctx, owner, value.ID, cycleOne[0].ID, tournaments.MatchResultInput{}); !errors.Is(err, tournaments.ErrInvalidTournamentInput) {
		t.Fatalf("draw without shootout error = %v", err)
	}
	for _, match := range cycleOne {
		winner := cycleOneWinners[orderedTeamPair(match.HomeTeamID, match.AwayTeamID)]
		input := tournaments.MatchResultInput{HomePenalties: intPointer(4), AwayPenalties: intPointer(3)}
		if winner == match.AwayTeamID {
			input.HomePenalties, input.AwayPenalties = intPointer(3), intPointer(4)
		}
		value, err = service.RecordResult(ctx, owner, value.ID, match.ID, input)
		if err != nil {
			t.Fatal(err)
		}
	}
	if value.TieBreakPools[0].CurrentCycle != 2 || mixedStageByType(t, value, tournaments.StageTypeQualificationTieBreak).State != "in_progress" {
		t.Fatalf("tiebreak after unresolved cycle = %#v", value.TieBreakPools)
	}

	cycleTwoWinners := map[[2]string]string{
		orderedTeamPair(teamIDs["A"], teamIDs["B"]): teamIDs["A"],
		orderedTeamPair(teamIDs["A"], teamIDs["C"]): teamIDs["A"],
		orderedTeamPair(teamIDs["B"], teamIDs["C"]): teamIDs["B"],
	}
	cycleTwo := tiebreakMatchesForCycle(value, tieBreakStage.ID, 2)
	if len(cycleTwo) != 3 {
		t.Fatalf("cycle two matches = %#v", cycleTwo)
	}
	for _, match := range cycleTwo {
		winner := cycleTwoWinners[orderedTeamPair(match.HomeTeamID, match.AwayTeamID)]
		input := tournaments.MatchResultInput{HomeScore: 1}
		if winner == match.AwayTeamID {
			input = tournaments.MatchResultInput{AwayScore: 1}
		}
		value, err = service.RecordResult(ctx, owner, value.ID, match.ID, input)
		if err != nil {
			t.Fatal(err)
		}
	}
	if mixedStageByType(t, value, tournaments.StageTypeQualificationTieBreak).State != "completed" || value.TieBreakPools[0].State != "completed" {
		t.Fatalf("completed tiebreak = stages %#v, pools %#v", value.Stages, value.TieBreakPools)
	}
	value, err = service.StartElimination(ctx, owner, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	eliminationStage := mixedStageByType(t, value, "single_elimination")
	qualified := []string{}
	for _, assignment := range value.StageTeams {
		if assignment.StageID == eliminationStage.ID {
			qualified = append(qualified, assignment.TeamID)
		}
	}
	if len(qualified) != 2 || qualified[0] != teamIDs["A"] || qualified[1] != teamIDs["B"] {
		t.Fatalf("qualified teams = %#v", qualified)
	}
}

func TestIntegrationMixedTournamentRejectsAnUnbalancedRoster(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-invalid@example.com", "mixed_invalid", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := make([]tournaments.TeamInput, 10)
	for index := range teams {
		teams[index] = tournaments.TeamInput{Name: fmt.Sprintf("Team %d", index+1)}
	}
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Unbalanced", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{
		Format:             tournaments.FormatLeagueThenSingleElimination,
		RoundRobinLegs:     1,
		LeagueStructure:    tournaments.LeagueStructureGroups,
		GroupCount:         4,
		QualifiersPerGroup: 2,
	})
	if !errors.Is(err, tournaments.ErrInvalidMixedConfiguration) {
		t.Fatalf("error = %v", err)
	}
}

func TestIntegrationMixedTournamentPersistsBalancedGroupsAndQualifiers(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-groups@example.com", "mixed_groups", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := make([]tournaments.TeamInput, 8)
	for index := range teams {
		teams[index] = tournaments.TeamInput{Name: fmt.Sprintf("Group team %d", index+1)}
	}
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Mixed groups", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	tournament, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{
		Format:             tournaments.FormatLeagueThenSingleElimination,
		RoundRobinLegs:     1,
		LeagueStructure:    tournaments.LeagueStructureGroups,
		GroupCount:         2,
		QualifiersPerGroup: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	groupTeams := map[int]int{}
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == mixedStageByType(t, tournament, "league").ID {
			groupTeams[assignment.GroupNumber]++
		}
	}
	if groupTeams[1] != 4 || groupTeams[2] != 4 {
		t.Fatalf("group composition = %#v", groupTeams)
	}
	groupMatches := map[int]int{}
	for _, match := range tournament.Matches {
		groupMatches[match.GroupNumber]++
		if _, err = recordHigherSeedWin(ctx, service, owner, tournament, match); err != nil {
			t.Fatal(err)
		}
	}
	if groupMatches[1] != 6 || groupMatches[2] != 6 {
		t.Fatalf("group fixtures = %#v", groupMatches)
	}
	tournament, err = service.StartElimination(ctx, owner, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	eliminationStage := mixedStageByType(t, tournament, "single_elimination")
	qualified := 0
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == eliminationStage.ID {
			qualified++
		}
	}
	if qualified != 4 {
		t.Fatalf("qualified teams = %d, want 4", qualified)
	}
}

func TestIntegrationMixedTournamentSupportsOddGroupsAndTwoLegs(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-odd-groups@example.com", "mixed_odd_groups", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := make([]tournaments.TeamInput, 6)
	for index := range teams {
		teams[index] = tournaments.TeamInput{Name: fmt.Sprintf("Odd group team %d", index+1)}
	}
	created, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Odd mixed groups", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{
		Format:             tournaments.FormatLeagueThenSingleElimination,
		RoundRobinLegs:     2,
		LeagueStructure:    tournaments.LeagueStructureGroups,
		GroupCount:         2,
		QualifiersPerGroup: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	groupMatches := map[int]int{}
	for _, match := range started.Matches {
		groupMatches[match.GroupNumber]++
	}
	if groupMatches[1] != 6 || groupMatches[2] != 6 || len(started.Matches) != 12 {
		t.Fatalf("two-leg odd-group fixtures = %#v (%d total)", groupMatches, len(started.Matches))
	}
}

func TestIntegrationMixedTournamentReplacesAWithdrawnQualifier(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-withdrawal@example.com", "mixed_withdrawal", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, owner, tournaments.CreateInput{
		Name: "Mixed withdrawal", Sport: tournaments.SportFootball,
		Teams: []tournaments.TeamInput{{Name: "First"}, {Name: "Second"}, {Name: "Third"}, {Name: "Fourth"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{
		Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1,
		LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	withdrawnID := started.Teams[0].ID
	started, err = service.WithdrawTeam(ctx, owner, created.ID, withdrawnID)
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range started.Matches {
		if match.State == "completed" {
			continue
		}
		if _, err = recordHigherSeedWin(ctx, service, owner, started, match); err != nil {
			t.Fatal(err)
		}
	}
	transitioned, err := service.StartElimination(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	eliminationStage := mixedStageByType(t, transitioned, "single_elimination")
	qualified := []string{}
	for _, assignment := range transitioned.StageTeams {
		if assignment.StageID == eliminationStage.ID {
			qualified = append(qualified, assignment.TeamID)
		}
	}
	if len(qualified) != 2 || qualified[0] == withdrawnID || qualified[1] == withdrawnID {
		t.Fatalf("qualified after withdrawal = %#v; withdrawn = %s", qualified, withdrawnID)
	}
}

func TestIntegrationMixedTournamentTransitionIsAuthorizedAndConcurrentSafe(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-concurrent@example.com", "mixed_concurrent", "password123")
	outsider := createVerifiedLocalAccount(t, ctx, pool, "mixed-outsider@example.com", "mixed_outsider", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Mixed concurrent", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "One"}, {Name: "Two"}}})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.RecordResult(ctx, owner, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.StartElimination(ctx, outsider, created.ID); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("outsider transition error = %v", err)
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Go(func() {
			<-start
			_, transitionErr := service.StartElimination(ctx, owner, created.ID)
			errs <- transitionErr
		})
	}
	close(start)
	group.Wait()
	close(errs)
	successes, conflicts := 0, 0
	for transitionErr := range errs {
		switch {
		case transitionErr == nil:
			successes++
		case errors.Is(transitionErr, tournaments.ErrTournamentStageTransitionConflict):
			conflicts++
		default:
			t.Fatalf("concurrent transition error = %v", transitionErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent transitions successes/conflicts = %d/%d", successes, conflicts)
	}
}

func TestIntegrationMixedBasketballCompletesThroughTheBracket(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-basketball@example.com", "mixed_basketball", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Mixed basketball", Sport: tournaments.SportBasketball, Teams: []tournaments.TeamInput{{Name: "One"}, {Name: "Two"}}})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.RecordResult(ctx, owner, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 10, AwayScore: 8}); err != nil {
		t.Fatal(err)
	}
	transitioned, err := service.StartElimination(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var final tournaments.Match
	eliminationStage := mixedStageByType(t, transitioned, "single_elimination")
	for _, match := range transitioned.Matches {
		if match.StageID == eliminationStage.ID {
			final = match
		}
	}
	if final.ID == "" {
		t.Fatal("missing basketball final")
	}
	if _, err = service.RecordResult(ctx, owner, created.ID, final.ID, tournaments.MatchResultInput{HomeScore: 20, AwayScore: 18}); err != nil {
		t.Fatal(err)
	}
	completed, err := service.Complete(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != "completed" || len(completed.ChampionTeamIDs) != 1 {
		t.Fatalf("completed basketball tournament = %#v", completed)
	}
}

func TestIntegrationMixedHandballCompletesWithSevenMetreShootout(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "mixed-handball@example.com", "mixed_handball", "password123")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Mixed handball", Sport: tournaments.SportHandball, Teams: []tournaments.TeamInput{{Name: "One"}, {Name: "Two"}}})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.Start(ctx, owner, created.ID, tournaments.StartInput{Format: tournaments.FormatLeagueThenSingleElimination, RoundRobinLegs: 1, LeagueStructure: tournaments.LeagueStructureSingleTable, QualifierCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.RecordResult(ctx, owner, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 28, AwayScore: 26}); err != nil {
		t.Fatal(err)
	}
	transitioned, err := service.StartElimination(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var final tournaments.Match
	eliminationStage := mixedStageByType(t, transitioned, "single_elimination")
	for _, match := range transitioned.Matches {
		if match.StageID == eliminationStage.ID {
			final = match
		}
	}
	if final.ID == "" {
		t.Fatal("missing handball final")
	}
	if _, err = service.RecordResult(ctx, owner, created.ID, final.ID, tournaments.MatchResultInput{
		HomeScore: 30, AwayScore: 30, HomePenalties: intPointer(5), AwayPenalties: intPointer(4),
	}); err != nil {
		t.Fatal(err)
	}
	completed, err := service.Complete(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != "completed" || len(completed.ChampionTeamIDs) != 1 {
		t.Fatalf("completed handball tournament = %#v", completed)
	}
}

func mixedStageByType(t *testing.T, tournament tournaments.Tournament, stageType string) tournaments.Stage {
	t.Helper()
	for _, stage := range tournament.Stages {
		if stage.Type == stageType {
			return stage
		}
	}
	t.Fatalf("missing %s stage in %#v", stageType, tournament.Stages)
	return tournaments.Stage{}
}

func recordHigherSeedWin(ctx context.Context, service tournaments.CreationService, owner string, tournament tournaments.Tournament, match tournaments.Match) (tournaments.Tournament, error) {
	position := map[string]int{}
	for _, team := range tournament.Teams {
		position[team.ID] = team.Position
	}
	input := tournaments.MatchResultInput{HomeScore: 1}
	if position[match.AwayTeamID] < position[match.HomeTeamID] {
		input = tournaments.MatchResultInput{AwayScore: 1}
	}
	return service.RecordResult(ctx, owner, tournament.ID, match.ID, input)
}

func orderedTeamPair(first, second string) [2]string {
	if first > second {
		first, second = second, first
	}
	return [2]string{first, second}
}

func tiebreakMatchesForCycle(tournament tournaments.Tournament, stageID string, cycle int) []tournaments.Match {
	matches := []tournaments.Match{}
	for _, match := range tournament.Matches {
		if match.StageID == stageID && match.RoundNumber == cycle {
			matches = append(matches, match)
		}
	}
	return matches
}

func intPointer(value int) *int { return &value }
