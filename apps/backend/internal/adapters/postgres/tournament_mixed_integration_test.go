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
	if len(tournament.Stages) != 2 || tournament.Stages[0].State != "in_progress" || tournament.Stages[1].State != "pending" {
		t.Fatalf("stages after start = %#v", tournament.Stages)
	}
	leagueMatchIDs := make([]string, 0, len(tournament.Matches))
	for _, match := range tournament.Matches {
		if match.StageID != tournament.Stages[0].ID {
			t.Fatalf("unexpected match before transition: %#v", match)
		}
		leagueMatchIDs = append(leagueMatchIDs, match.ID)
		if _, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
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
	if tournament.Stages[0].State != "completed" || tournament.Stages[1].State != "in_progress" {
		t.Fatalf("stages after transition = %#v", tournament.Stages)
	}
	seeded := []string{}
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == tournament.Stages[1].ID {
			seeded = append(seeded, assignment.TeamID)
		}
	}
	if len(seeded) != 4 {
		t.Fatalf("knockout seeds = %#v", seeded)
	}
	firstRound := []tournaments.Match{}
	for _, match := range tournament.Matches {
		if match.StageID == tournament.Stages[1].ID && match.RoundNumber == 1 {
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
		if assignment.StageID == tournament.Stages[0].ID {
			groupTeams[assignment.GroupNumber]++
		}
	}
	if groupTeams[1] != 4 || groupTeams[2] != 4 {
		t.Fatalf("group composition = %#v", groupTeams)
	}
	groupMatches := map[int]int{}
	for _, match := range tournament.Matches {
		groupMatches[match.GroupNumber]++
		if _, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
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
	qualified := 0
	for _, assignment := range tournament.StageTeams {
		if assignment.StageID == tournament.Stages[1].ID {
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
		if _, err = service.RecordResult(ctx, owner, created.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
			t.Fatal(err)
		}
	}
	transitioned, err := service.StartElimination(ctx, owner, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	qualified := []string{}
	for _, assignment := range transitioned.StageTeams {
		if assignment.StageID == transitioned.Stages[1].ID {
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
	for _, match := range transitioned.Matches {
		if match.StageID == transitioned.Stages[1].ID {
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
