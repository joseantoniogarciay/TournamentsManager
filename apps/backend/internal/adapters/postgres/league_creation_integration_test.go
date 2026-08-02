package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
)

// Esta prueba usa una base efímera preparada por el comando de integración.
func TestIntegrationLeagueCreationAndStartWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TM_INTEGRATION_DATABASE_URL")
	if databaseURL == "" || os.Getenv("TM_RUN_INTEGRATION") != "1" {
		t.Skip("la integración PostgreSQL no está activada")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("conectar PostgreSQL: %v", err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, "TRUNCATE accounts CASCADE"); err != nil {
		t.Fatalf("limpiar base: %v", err)
	}
	var accountID string
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, verified_at) VALUES ('organizer@example.test', 'es', 'verified', 'organizer', now()) RETURNING id::text`).Scan(&accountID); err != nil {
		t.Fatalf("crear organizadora: %v", err)
	}
	service := leagues.NewCreationService(NewAccountLeagueRepository(pool))
	created, err := service.Create(ctx, accountID, leagues.CreateInput{Name: "Liga de verano", Teams: []leagues.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}, {Name: "Amarillos"}}})
	if err != nil {
		t.Fatalf("crear liga: %v", err)
	}
	if created.State != "published" || len(created.Teams) != 4 || len(created.Matches) != 0 {
		t.Fatalf("liga creada = %#v, se esperaba publicada sin partidos", created)
	}
	started, err := service.Start(ctx, accountID, created.ID, leagues.StartInput{RoundRobinLegs: 2})
	if err != nil {
		t.Fatalf("iniciar liga: %v", err)
	}
	if started.State != "in_progress" || len(started.Matches) != 12 {
		t.Fatalf("liga iniciada = state %q, partidos %d; se esperaba in_progress y 12", started.State, len(started.Matches))
	}
	if _, err := service.Start(ctx, accountID, created.ID, leagues.StartInput{RoundRobinLegs: 1}); err != leagues.ErrLeagueConflict {
		t.Fatalf("segundo inicio = %v, se esperaba %v", err, leagues.ErrLeagueConflict)
	}
}
