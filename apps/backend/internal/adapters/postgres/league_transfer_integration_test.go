package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
)

func TestIntegrationTransferLeagueOwnershipWithPostgres(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	previous := createVerifiedLocalAccount(t, ctx, pool, "previous@example.test", "previous", "correct password")
	recipient := createVerifiedLocalAccount(t, ctx, pool, "recipient@example.test", "recipient", "correct password")
	service := leagues.NewCreationService(NewAccountLeagueRepository(pool))
	created, err := service.Create(ctx, previous, leagues.CreateInput{Name: "Liga transferida", Teams: []leagues.TeamInput{{Name: "Uno"}, {Name: "Dos"}}})
	if err != nil {
		t.Fatalf("crear liga: %v", err)
	}
	if err := service.AssignAdministrator(ctx, previous, created.ID, "recipient"); err != nil {
		t.Fatalf("delegar destinataria: %v", err)
	}
	if err := service.TransferOwnership(ctx, previous, created.ID, "recipient"); err != nil {
		t.Fatalf("transferir: %v", err)
	}
	if _, err := service.ListAdministrators(ctx, previous, created.ID); !errors.Is(err, leagues.ErrLeagueForbidden) {
		t.Fatalf("anterior organizadora conserva acceso: %v", err)
	}
	if _, err := service.ListAdministrators(ctx, recipient, created.ID); err != nil {
		t.Fatalf("nueva organizadora no puede administrar: %v", err)
	}
	var delegated, notifications int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM league_administrators WHERE league_id = $1 AND account_id = $2`, created.ID, recipient).Scan(&delegated); err != nil || delegated != 0 {
		t.Fatalf("delegación destinataria = %d, %v", delegated, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM account_notifications WHERE league_id = $1 AND account_id = $2 AND kind = 'league_ownership_transferred'`, created.ID, recipient).Scan(&notifications); err != nil || notifications != 1 {
		t.Fatalf("notificación = %d, %v", notifications, err)
	}
}
