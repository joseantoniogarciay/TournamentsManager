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
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Bracket", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}}})
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

func TestIntegrationTennisSetsAreValidatedPersistedAndHistorized(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(t, ctx, pool, "tennis@example.test", "tennis_owner", "correct horse battery staple")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{
		Name: "Tenis", Sport: tournaments.SportTennis, BestOfSets: 3,
		Teams: []tournaments.TeamInput{{Name: "Ana"}, {Name: "Bea"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1}); !errors.Is(err, tournaments.ErrInvalidTournamentInput) {
		t.Fatalf("tennis league started: %v", err)
	}
	tournament, err = service.Start(ctx, owner, tournament.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	match := tournament.Matches[0]
	if _, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 6, AwayScore: 5}, {HomeScore: 6, AwayScore: 0}}}); !errors.Is(err, tournaments.ErrInvalidBracketResult) {
		t.Fatalf("impossible set accepted: %v", err)
	}
	first := tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 6, AwayScore: 4}, {HomeScore: 3, AwayScore: 6}, {HomeScore: 7, AwayScore: 6}}}
	tournament, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, first)
	if err != nil {
		t.Fatal(err)
	}
	if tournament.Matches[0].HomeScore == nil || *tournament.Matches[0].HomeScore != 2 || len(tournament.Matches[0].Sets) != 3 {
		t.Fatalf("persisted tennis result = %#v", tournament.Matches[0])
	}
	second := tournaments.MatchResultInput{Sets: []tournaments.SetScore{{HomeScore: 4, AwayScore: 6}, {HomeScore: 5, AwayScore: 7}}}
	tournament, err = service.RecordResult(ctx, owner, tournament.ID, match.ID, second)
	if err != nil {
		t.Fatal(err)
	}
	if tournament.Matches[0].WinnerTeamID != match.AwayTeamID || len(tournament.Matches[0].Sets) != 2 {
		t.Fatalf("corrected tennis result = %#v", tournament.Matches[0])
	}
	var changes, snapshots int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1`, match.ID).Scan(&changes); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM match_result_change_sets s JOIN match_result_changes c ON c.id=s.change_id WHERE c.match_id=$1`, match.ID).Scan(&snapshots); err != nil {
		t.Fatal(err)
	}
	if changes != 2 || snapshots != 5 {
		t.Fatalf("history changes=%d set snapshots=%d, want 2 and 5", changes, snapshots)
	}
}
