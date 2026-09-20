package postgres

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/accounts"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationPurgeExpiredAccountsDeletesOnlyExpiredAccountsAndAnonymizesHistory(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repository := NewAccountTournamentRepository(pool)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "organizer@example.test", "organizer", "correct password")
	administratorID := createVerifiedLocalAccount(t, ctx, pool, "administrator@example.test", "administrator", "correct password")
	notDueID := createVerifiedLocalAccount(t, ctx, pool, "not-due@example.test", "notdue", "correct password")
	creation := tournaments.NewCreationService(repository)
	created, err := creation.Create(ctx, organizerID, tournaments.CreateInput{Name: "Liga de purga", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	if err := creation.AssignAdministrator(ctx, organizerID, created.ID, "administrator"); err != nil {
		t.Fatalf("asignar administradora = %v", err)
	}
	started, err := creation.Start(ctx, organizerID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar liga = %v", err)
	}
	if _, err := creation.RecordResult(ctx, administratorID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1}); err != nil {
		t.Fatalf("registrar resultado = %v", err)
	}
	if _, err := repository.ScheduleAccountDeletion(ctx, administratorID); err != nil {
		t.Fatalf("programar baja vencida = %v", err)
	}
	if _, err := repository.ScheduleAccountDeletion(ctx, notDueID); err != nil {
		t.Fatalf("programar baja no vencida = %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE accounts SET deletion_requested_at = now() - interval '30 days' WHERE id = $1`, administratorID); err != nil {
		t.Fatalf("vencer baja = %v", err)
	}
	deleted, err := accounts.NewService(repository).PurgeExpired(ctx)
	if err != nil {
		t.Fatalf("purgar cuentas = %v", err)
	}
	if deleted != 1 {
		t.Fatalf("cuentas eliminadas = %d, se esperaba 1", deleted)
	}
	var accountExists, notDueExists, historyExists, historyAnonymized bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE id = $1)`, administratorID).Scan(&accountExists); err != nil {
		t.Fatalf("comprobar cuenta eliminada = %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE id = $1)`, notDueID).Scan(&notDueExists); err != nil {
		t.Fatalf("comprobar cuenta no vencida = %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM match_result_changes WHERE match_id = $1), EXISTS (SELECT 1 FROM match_result_changes WHERE match_id = $1 AND changed_by_account_id IS NULL)`, started.Matches[0].ID).Scan(&historyExists, &historyAnonymized); err != nil {
		t.Fatalf("comprobar historial = %v", err)
	}
	if accountExists || !notDueExists || !historyExists || !historyAnonymized {
		t.Fatalf("purga: eliminada=%v noVencida=%v historial=%v anonimizado=%v", accountExists, notDueExists, historyExists, historyAnonymized)
	}
	deleted, err = accounts.NewService(repository).PurgeExpired(ctx)
	if err != nil || deleted != 0 {
		t.Fatalf("segunda purga = (%d, %v), se esperaba (0, nil)", deleted, err)
	}
}
