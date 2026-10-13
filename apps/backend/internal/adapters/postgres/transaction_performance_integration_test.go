package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

type transactionQueryTiming struct {
	started time.Time
	kind    string
}
type transactionTimingKey struct{}
type transactionMeasurement struct {
	insertTimes []time.Duration
	updates     int
}

func (m *transactionMeasurement) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	kind := ""
	if strings.HasPrefix(data.SQL, "INSERT INTO matches") {
		kind = "insert"
	} else if strings.HasPrefix(data.SQL, "UPDATE matches SET") {
		kind = "update"
	}
	return context.WithValue(ctx, transactionTimingKey{}, transactionQueryTiming{time.Now(), kind})
}
func (m *transactionMeasurement) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	value, _ := ctx.Value(transactionTimingKey{}).(transactionQueryTiming)
	switch value.kind {
	case "insert":
		m.insertTimes = append(m.insertTimes, time.Since(value.started))
	case "update":
		m.updates++
	}
}

// This probe uses only the public repository API so it can run unchanged against
// the pre-change Git snapshot. Run separately from regression tests and record
// both versions on the same isolated database; timing is evidence, not an SLO.
func TestIntegrationTransactionPerformance(t *testing.T) {
	if os.Getenv("TM_MEASURE_TRANSACTIONS") != "1" {
		t.Skip("performance comparison is opt-in")
	}
	fixturePool := integrationPool(t)
	ctx := context.Background()
	owner := createVerifiedLocalAccount(ctx, t, fixturePool, "performance@example.test", "performanceowner", "correct password")
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
	for iteration := 0; iteration < 6; iteration++ { // First sample is warmup.
		league, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Maximum league", Sport: tournaments.SportFootball, Teams: teams})
		if err != nil {
			t.Fatal(err)
		}
		measurement.insertTimes = nil
		started := time.Now()
		league, err = service.Start(ctx, owner, league.ID, tournaments.StartInput{Format: "league", RoundRobinLegs: 2})
		elapsed := time.Since(started)
		if err != nil || len(league.Matches) != 4032 {
			t.Fatalf("maximum calendar: %d %v", len(league.Matches), err)
		}
		var sum, maximum time.Duration
		for _, duration := range measurement.insertTimes {
			sum += duration
			maximum = max(maximum, duration)
		}
		t.Logf("MEASURE league sample=%d total_ms=%.3f inserts=%d insert_total_ms=%.3f insert_max_ms=%.3f", iteration, float64(elapsed)/float64(time.Millisecond), len(measurement.insertTimes), float64(sum)/float64(time.Millisecond), float64(maximum)/float64(time.Millisecond))

		bracket, err := service.Create(ctx, owner, tournaments.CreateInput{Name: "Maximum bracket", Sport: tournaments.SportFootball, Teams: teams})
		if err != nil {
			t.Fatal(err)
		}
		bracket, err = service.Start(ctx, owner, bracket.ID, tournaments.StartInput{Format: "single_elimination"})
		if err != nil {
			t.Fatal(err)
		}
		measurement.updates = 0
		started = time.Now()
		bracket, err = service.RecordResult(ctx, owner, bracket.ID, bracket.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 1})
		if err != nil || len(bracket.Matches) != 63 {
			t.Fatalf("bracket result: %v", err)
		}
		t.Logf("MEASURE bracket sample=%d total_ms=%.3f updates=%d", iteration, float64(time.Since(started))/float64(time.Millisecond), measurement.updates)
	}
	for iteration := 0; iteration < 6; iteration++ {
		for _, table := range []string{"google_risc_events", "apple_login_challenges", "legal_account_acceptances"} {
			if _, err := pool.Exec(ctx, "TRUNCATE "+table); err != nil {
				t.Fatal(err)
			}
		}
		for _, query := range []string{
			`INSERT INTO google_risc_events(id,received_at,expires_at) SELECT 'measure-'||i,now()-interval '2 days',now()-interval '1 day' FROM generate_series(1,1000) i`,
			`INSERT INTO apple_login_challenges(state_hash,nonce_hash,proof_hash,platform,created_at,expires_at) SELECT decode(md5(i::text)||md5(i::text),'hex'),decode(repeat('00',32),'hex'),decode(repeat('00',32),'hex'),'web',now()-interval '2 days',now()-interval '1 day' FROM generate_series(1,1000) i`,
			`INSERT INTO legal_account_acceptances(email_hash,terms_version,terms_content_hash,source,retention_until) SELECT decode(repeat('00',32),'hex'),'test',decode(repeat('00',32),'hex'),'password_registration',now()-interval '1 day' FROM generate_series(1,1000)`,
		} {
			if _, err := pool.Exec(ctx, query); err != nil {
				t.Fatal(err)
			}
		}
		var purgeTime, firstTime time.Duration
		firstDeleted, passes := 0, 0
		remaining := 3000
		for remaining > 0 {
			started := time.Now()
			if _, err := NewAccountTournamentRepository(pool).PurgeExpired(ctx, 100); err != nil {
				t.Fatal(err)
			}
			duration := time.Since(started)
			purgeTime += duration
			passes++
			var next int
			if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM google_risc_events)+(SELECT count(*) FROM apple_login_challenges)+(SELECT count(*) FROM legal_account_acceptances)`).Scan(&next); err != nil {
				t.Fatal(err)
			}
			if passes == 1 {
				firstTime, firstDeleted = duration, remaining-next
			}
			if next >= remaining {
				t.Fatal("purge made no progress")
			}
			remaining = next
		}
		t.Logf("MEASURE purge sample=%d first_ms=%.3f first_deleted=%d drain_ms=%.3f passes=%d", iteration, float64(firstTime)/float64(time.Millisecond), firstDeleted, float64(purgeTime)/float64(time.Millisecond), passes)
	}
}
