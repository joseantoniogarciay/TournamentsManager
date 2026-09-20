package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationTournamentMutationsRequireOrganizerOrAdministrator(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "organizer@example.test", "organizer", "correct password")
	outsiderID := createVerifiedLocalAccount(t, ctx, pool, "outsider@example.test", "outsider", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, organizerID, tournaments.CreateInput{Name: "Liga permisos", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}}})
	if err != nil {
		t.Fatalf("crear liga = %v", err)
	}
	if _, err := service.Start(ctx, outsiderID, created.ID, tournaments.StartInput{RoundRobinLegs: 1}); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("iniciar como ajena = %v, se esperaba %v", err, tournaments.ErrTournamentForbidden)
	}
	started, err := service.Start(ctx, organizerID, created.ID, tournaments.StartInput{RoundRobinLegs: 1})
	if err != nil {
		t.Fatalf("iniciar como organizadora = %v", err)
	}
	if _, err := service.RecordResult(ctx, outsiderID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 1, AwayScore: 0}); !errors.Is(err, tournaments.ErrMatchResultForbidden) {
		t.Fatalf("registrar como ajena = %v, se esperaba %v", err, tournaments.ErrMatchResultForbidden)
	}
	if _, err := service.Cancel(ctx, outsiderID, created.ID); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("cancelar como ajena = %v, se esperaba %v", err, tournaments.ErrTournamentForbidden)
	}
	if _, err := service.Complete(ctx, outsiderID, created.ID); !errors.Is(err, tournaments.ErrTournamentForbidden) {
		t.Fatalf("finalizar como ajena = %v, se esperaba %v", err, tournaments.ErrTournamentForbidden)
	}
}
