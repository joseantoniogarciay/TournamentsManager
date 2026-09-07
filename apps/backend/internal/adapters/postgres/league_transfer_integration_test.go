package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationTransferTournamentOwnershipWithPostgres(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	previous := createVerifiedLocalAccount(t, ctx, pool, "previous@example.test", "previous", "correct password")
	recipient := createVerifiedLocalAccount(t, ctx, pool, "recipient@example.test", "recipient", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, previous, tournaments.CreateInput{Name: "Liga transferida", Teams: []tournaments.TeamInput{{Name: "Uno"}, {Name: "Dos"}}})
	if err != nil {
		t.Fatalf("crear liga: %v", err)
	}
	if err := service.AssignAdministrator(ctx, previous, created.ID, "recipient"); err != nil {
		t.Fatalf("delegar destinataria: %v", err)
	}
	if err := service.TransferOwnership(ctx, previous, created.ID, "recipient"); err != nil {
		t.Fatalf("transferir: %v", err)
	}
	if _, err := service.ListAdministrators(ctx, previous, created.ID); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("anterior organizadora conserva acceso: %v", err)
	}
	if _, err := service.ListAdministrators(ctx, recipient, created.ID); err != nil {
		t.Fatalf("nueva organizadora no puede administrar: %v", err)
	}
	var delegated, notifications int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournament_administrators WHERE tournament_id = $1 AND account_id = $2`, created.ID, recipient).Scan(&delegated); err != nil || delegated != 0 {
		t.Fatalf("delegación destinataria = %d, %v", delegated, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM account_notifications WHERE tournament_id = $1 AND account_id = $2 AND kind = 'tournament_ownership_transferred'`, created.ID, recipient).Scan(&notifications); err != nil || notifications != 1 {
		t.Fatalf("notificación = %d, %v", notifications, err)
	}
}
