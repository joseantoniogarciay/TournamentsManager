package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationRuntimePoolEnforcesStatementAndLockBudgets(t *testing.T) {
	fixture := integrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	id := createVerifiedLocalAccount(ctx, t, fixture, "budget@example.test", "budgetuser", "correct password")
	pool, err := NewPool(ctx, os.Getenv("TM_INTEGRATION_DATABASE_URL"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var statement, lock string
	if err := pool.QueryRow(ctx, `SELECT current_setting('statement_timeout'),current_setting('lock_timeout')`).Scan(&statement, &lock); err != nil || statement != "5s" || lock != "2s" {
		t.Fatalf("budgets: %s %s %v", statement, lock, err)
	}
	_, err = pool.Exec(ctx, `SELECT pg_sleep(30)`)
	var failure *pgconn.PgError
	if !errors.As(err, &failure) || failure.Code != "57014" {
		t.Fatalf("statement not interrupted: %v", err)
	}
	tx, err := fixture.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`, id)
	if !errors.As(err, &failure) || failure.Code != "55P03" {
		t.Fatalf("lock not interrupted: %v", err)
	}
}

func TestIntegrationBrowsingBudgetWithLargeFamilyAndMaximumLeague(t *testing.T) {
	fixture := integrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	owner := createVerifiedLocalAccount(ctx, t, fixture, "volume@example.test", "volumeowner", "correct password")
	pool, err := NewPool(ctx, os.Getenv("TM_INTEGRATION_DATABASE_URL"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewAccountTournamentRepository(pool)
	create := tournaments.NewCreationService(repo)
	teams := make([]tournaments.TeamInput, 64)
	for i := range teams {
		teams[i].Name = fmt.Sprintf("Team %02d", i)
	}
	root, err := create.Create(ctx, owner, tournaments.CreateInput{Name: "Volume", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	root, err = create.Start(ctx, owner, root.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tournaments (organizer_account_id,name,sport,state,published_at,repetition_source_id,repetition_account_id,repetition_request_id) SELECT $1,'Volume edition','football','published',now(),$2,$1,uuidv7() FROM generate_series(1,5000)`, owner, root.ID); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	editions, err := repo.Editions(ctx, root.ID, "", 51)
	if err != nil || len(editions) != 51 {
		t.Fatalf("family: %d %v", len(editions), err)
	}
	t.Logf("5001-member family, 51 rows: %s", time.Since(started))
	started = time.Now()
	matches, err := repo.PersonalMatches(ctx, owner, "pending", "", 51)
	if err != nil || len(matches.Items) != 51 {
		t.Fatalf("matches: %d %v", len(matches.Items), err)
	}
	t.Logf("64-team league / %d matches, 51 personal rows: %s", len(root.Matches), time.Since(started))
}
