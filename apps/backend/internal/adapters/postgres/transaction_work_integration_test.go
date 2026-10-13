package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationBracketWritesOnlyChangedMatchesAndKeepsHistory(t *testing.T) {
	fixturePool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(ctx, t, fixturePool, "changed@example.test", "changedowner", "correct password")
	measurement := &transactionMeasurement{}
	pool, err := NewPool(ctx, os.Getenv("TM_INTEGRATION_DATABASE_URL"), measurement)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	teams := make([]tournaments.TeamInput, 64)
	for i := range teams {
		teams[i].Name = fmt.Sprintf("Team %02d", i)
	}
	value, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Changed rows", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	value, err = service.Start(ctx, owner, value.ID, tournaments.StartInput{Format: "single_elimination"})
	if err != nil {
		t.Fatal(err)
	}
	matchID := value.Matches[0].ID
	for _, expected := range []int{2, 0} {
		measurement.updates = 0
		if _, err := service.RecordResult(ctx, owner, value.ID, matchID, tournaments.MatchResultInput{HomeScore: 1}); err != nil {
			t.Fatal(err)
		}
		if measurement.updates != expected {
			t.Fatalf("match updates=%d want=%d", measurement.updates, expected)
		}
	}
	var history int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM match_result_changes WHERE match_id=$1`, matchID).Scan(&history); err != nil || history != 2 {
		t.Fatalf("history=%d err=%v", history, err)
	}
}

func TestIntegrationFixtureLaterBatchFailureRollsBackEarlierBatches(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(ctx, t, pool, "batch@example.test", "batchowner", "correct password")
	repo := NewAccountTournamentRepository(pool)
	teams := make([]tournaments.TeamInput, 24)
	for i := range teams {
		teams[i].Name = fmt.Sprintf("Team %02d", i)
	}
	created, err := repo.Create(ctx, owner, tournaments.CreateInput{Name: "Rollback", Sport: tournaments.SportFootball, Teams: teams})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var stageID string
	if err := tx.QueryRow(ctx, `INSERT INTO tournament_stages(tournament_id,position,type,state,round_robin_legs) VALUES ($1,1,'league','in_progress',1) RETURNING id::text`, created.ID).Scan(&stageID); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, len(created.Teams))
	for i, team := range created.Teams {
		ids[i] = team.ID
	}
	values := fixtures(ids, 1) // 276 matches: first batch succeeds, second fails.
	if len(values) <= fixtureInsertBatchSize {
		t.Fatal("fixture does not cross a batch boundary")
	}
	values[len(values)-1].home = "00000000-0000-0000-0000-000000000000"
	if err := insertFixtures(ctx, tx, created.ID, stageID, values, 0); err == nil {
		t.Fatal("foreign key failure expected")
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var matches, stages int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM matches WHERE tournament_id=$1),(SELECT count(*) FROM tournament_stages WHERE tournament_id=$1)`, created.ID).Scan(&matches, &stages); err != nil {
		t.Fatal(err)
	}
	if matches != 0 || stages != 0 {
		t.Fatalf("partial transaction survived: matches=%d stages=%d", matches, stages)
	}
}

func TestIntegrationPurgeBoundsEveryExpiredCategory(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	// Events and unlinked legal evidence are not removed by integrationPool's account truncation.
	for _, table := range []string{"google_risc_events", "apple_login_challenges", "legal_account_acceptances"} {
		if _, err := pool.Exec(ctx, "TRUNCATE "+table); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{
		`INSERT INTO google_risc_events(id,received_at,expires_at) SELECT 'batch-'||i,now()-interval '2 days',CASE WHEN i=4 THEN now()+interval '1 day' ELSE now()-interval '1 day' END FROM generate_series(1,4) i`,
		`INSERT INTO apple_login_challenges(state_hash,nonce_hash,proof_hash,platform,created_at,expires_at) SELECT decode(md5(i::text)||md5(i::text),'hex'),decode(repeat('00',32),'hex'),decode(repeat('00',32),'hex'),'web',now()-interval '2 days',CASE WHEN i=4 THEN now()+interval '1 day' ELSE now()-interval '1 day' END FROM generate_series(1,4) i`,
		`INSERT INTO legal_account_acceptances(email_hash,terms_version,terms_content_hash,source,retention_until) SELECT decode(repeat('00',32),'hex'),'test',decode(repeat('00',32),'hex'),'password_registration',CASE WHEN i=4 THEN now()+interval '1 day' ELSE now()-interval '1 day' END FROM generate_series(1,4) i`,
	} {
		if _, err := pool.Exec(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	for pass, expected := range []int{2, 1, 1} {
		if deleted, err := NewAccountTournamentRepository(pool).PurgeExpired(ctx, 2); err != nil || deleted != 0 {
			t.Fatalf("purge %d: %d %v", pass, deleted, err)
		}
		for _, table := range []string{"google_risc_events", "apple_login_challenges", "legal_account_acceptances"} {
			var count int
			if err := pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expected {
				t.Fatalf("pass=%d table=%s count=%d want=%d err=%v", pass, table, count, expected, err)
			}
		}
	}
}
