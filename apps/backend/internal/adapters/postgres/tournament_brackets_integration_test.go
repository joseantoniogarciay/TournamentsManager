package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationBracketLifecycle(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "bracket@example.test", "bracket_owner", "correct horse battery staple")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Bracket", Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}}})
	if err != nil {
		t.Fatal(err)
	}
	tournament, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tournament.Stages) != 1 || tournament.Stages[0].Type != "single_elimination" || len(tournament.Matches) != 7 || len(tournament.Standings) != 0 {
		t.Fatalf("invalid bracket: %#v", tournament)
	}
	first := tournament.Matches[0]
	if _, err = service.RecordResult(ctx, owner, tournament.ID, first.ID, tournaments.MatchResultInput{}); !errors.Is(err, tournaments.ErrInvalidBracketResult) {
		t.Fatalf("tie: %v", err)
	}
	if _, err = service.RecordResult(ctx, owner, tournament.ID, tournament.Matches[6].ID, tournaments.MatchResultInput{HomeScore: 1}); !errors.Is(err, tournaments.ErrBracketMatchNotReady) {
		t.Fatalf("future match: %v", err)
	}
	if _, err = service.WithdrawTeam(ctx, owner, tournament.ID, tournament.Teams[0].ID); !errors.Is(err, tournaments.ErrTournamentWithdrawalConflict) {
		t.Fatalf("league withdrawal applied to bracket: %v", err)
	}
	hp, ap := 4, 5
	tournament, err = service.RecordResult(ctx, owner, tournament.ID, first.ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 1, HomePenalties: &hp, AwayPenalties: &ap})
	if err != nil {
		t.Fatal(err)
	}
	if tournament.Matches[4].HomeTeamID != first.AwayTeamID {
		t.Fatal("winner not persisted into semifinal")
	}
	tournament, err = service.RecordResult(ctx, owner, tournament.ID, first.ID, tournaments.MatchResultInput{HomeScore: 2})
	if err != nil {
		t.Fatal(err)
	}
	if tournament.Matches[4].HomeTeamID != first.HomeTeamID {
		t.Fatal("correction not propagated")
	}
	for _, match := range tournament.Matches {
		if match.State != "pending" {
			continue
		}
		tournament, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, tournaments.MatchResultInput{HomeScore: 1})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = service.RecordResult(ctx, owner, tournament.ID, first.ID, tournaments.MatchResultInput{HomeScore: 3}); !errors.Is(err, tournaments.ErrBracketResultDependency) {
		t.Fatalf("dependent correction: %v", err)
	}
	var changes int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1`, first.ID).Scan(&changes); err != nil {
		t.Fatal(err)
	}
	if changes != 2 {
		t.Fatalf("history changed after rejected requests: %d", changes)
	}
	tournament, err = service.Complete(ctx, owner, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tournament.State != "completed" || tournament.Stages[0].State != "completed" || len(tournament.ChampionTeamIDs) != 1 || tournament.ChampionTeamIDs[0] != tournament.Matches[6].WinnerTeamID {
		t.Fatal("incorrect champion or lifecycle")
	}
}
