package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// Wait until the join has an older snapshot, then revoke or rotate its capability.
func TestIntegrationInvitationRevalidatedAfterWaitingForTournamentLock(t *testing.T) {
	for _, action := range []string{"revoke", "rotate", "start"} {
		t.Run(action, func(t *testing.T) { testInvitationAfterLock(t, action) })
	}
}
func testInvitationAfterLock(t *testing.T, action string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool := integrationPool(t)
	owner := createVerifiedLocalAccount(ctx, t, pool, "review-owner@example.test", "reviewowner", "correct password")
	participant := createVerifiedLocalAccount(ctx, t, pool, "review-player@example.test", "reviewplayer", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	tournament, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Audit", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Owner"}}})
	if err != nil {
		t.Fatal(err)
	}
	token, err := service.CreateTeamInvitation(ctx, owner, tournament.ID)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, `SELECT id FROM tournaments WHERE id=$1 FOR UPDATE`, tournament.ID); err != nil {
		t.Fatal(err)
	}
	outcome := make(chan error, 1)
	go func() {
		_, err := service.JoinTeamInvitation(ctx, participant, token, tournaments.TeamInput{Name: "Player"})
		outcome <- err
	}()
	for {
		var blocked bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE '%FOR UPDATE OF tournaments%')`).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			break
		}
		select {
		case err := <-outcome:
			t.Fatalf("join did not wait: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	query := `DELETE FROM tournament_team_invitations WHERE tournament_id=$1`
	if action == "rotate" {
		query = `UPDATE tournament_team_invitations SET token_hash=decode(repeat('ab',32),'hex') WHERE tournament_id=$1`
	}
	if action == "start" {
		query = `UPDATE tournaments SET state='in_progress' WHERE id=$1`
	}
	if _, err = tx.Exec(ctx, query, tournament.ID); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-outcome:
		if !errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
			t.Fatalf("stale capability accepted: %v", err)
		}
		var links int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournament_team_accounts WHERE tournament_id=$1 AND account_id=$2`, tournament.ID, participant).Scan(&links); err != nil || links != 0 {
			t.Fatalf("partial enrollment: %d %v", links, err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
