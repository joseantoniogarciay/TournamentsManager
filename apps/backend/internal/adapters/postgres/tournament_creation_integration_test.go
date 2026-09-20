package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationTournamentCreationAndStartWithPostgres(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	var accountID string
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, verified_at) VALUES ('organizer@example.test', 'es', 'verified', 'organizer', now()) RETURNING id::text`).Scan(&accountID); err != nil {
		t.Fatalf("crear organizadora: %v", err)
	}
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, accountID, tournaments.CreateInput{Name: "Liga de verano", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}, {Name: "Amarillos"}}})
	if err != nil {
		t.Fatalf("crear liga: %v", err)
	}
	publicCreated, err := service.GetPublic(ctx, created.ID)
	if err != nil || publicCreated.Matches == nil || publicCreated.Teams == nil {
		t.Fatalf("consultar liga recién creada = %#v, %v; se esperaban arrays no nulos", publicCreated, err)
	}
	administratorID := createVerifiedLocalAccount(t, ctx, pool, "administrator@example.test", "administrator", "correct password")
	if err := service.AssignAdministrator(ctx, accountID, created.ID, "administrator"); err != nil {
		t.Fatalf("asignar administradora: %v", err)
	}
	administrators, err := service.ListAdministrators(ctx, accountID, created.ID)
	if err != nil || len(administrators) != 1 || administrators[0] != "administrator" {
		t.Fatalf("listar administradoras = %#v, %v", administrators, err)
	}
	if err := service.RemoveAdministrator(ctx, accountID, created.ID, "administrator"); err != nil {
		t.Fatalf("retirar administradora: %v", err)
	}
	if err := service.AssignAdministrator(ctx, accountID, created.ID, "administrator"); err != nil {
		t.Fatalf("reasignar administradora: %v", err)
	}
	var assigned bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournament_administrators WHERE tournament_id = $1 AND account_id = $2)`, created.ID, administratorID).Scan(&assigned); err != nil || !assigned {
		t.Fatalf("comprobar administradora asignada = %v, %v", assigned, err)
	}
	if created.State != "published" || len(created.Teams) != 4 || len(created.Matches) != 0 {
		t.Fatalf("liga creada = %#v, se esperaba publicada sin partidos", created)
	}
	started, err := service.Start(ctx, accountID, created.ID, tournaments.StartInput{RoundRobinLegs: 2})
	if err != nil {
		t.Fatalf("iniciar liga: %v", err)
	}
	if started.State != "in_progress" || len(started.Matches) != 12 {
		t.Fatalf("liga iniciada = state %q, partidos %d; se esperaba in_progress y 12", started.State, len(started.Matches))
	}
	withOrganizerResult, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1})
	organizerMatch, found := leagueMatchByID(withOrganizerResult.Matches, started.Matches[0].ID)
	if err != nil || !found || organizerMatch.State != "completed" || organizerMatch.HomeScore == nil || *organizerMatch.HomeScore != 2 || organizerMatch.AwayScore == nil || *organizerMatch.AwayScore != 1 {
		t.Fatalf("resultado de organizadora = %#v, %v; se esperaba marcador 2-1", organizerMatch, err)
	}
	withResult, err := service.RecordResult(ctx, administratorID, created.ID, started.Matches[1].ID, tournaments.MatchResultInput{HomeScore: 2, AwayScore: 1})
	administratorMatch, found := leagueMatchByID(withResult.Matches, started.Matches[1].ID)
	if err != nil || !found || administratorMatch.State != "completed" || administratorMatch.HomeScore == nil || *administratorMatch.HomeScore != 2 || administratorMatch.AwayScore == nil || *administratorMatch.AwayScore != 1 {
		t.Fatalf("registrar resultado = %#v, %v; se esperaba marcador 2-1", administratorMatch, err)
	}
	corrected, err := service.RecordResult(ctx, administratorID, created.ID, started.Matches[1].ID, tournaments.MatchResultInput{HomeScore: 3, AwayScore: 0})
	correctedMatch, found := leagueMatchByID(corrected.Matches, started.Matches[1].ID)
	if err != nil || !found || correctedMatch.HomeScore == nil || *correctedMatch.HomeScore != 3 || correctedMatch.AwayScore == nil || *correctedMatch.AwayScore != 0 {
		t.Fatalf("corregir resultado = %#v, %v; se esperaba marcador 3-0", correctedMatch, err)
	}
	var historyCount, previousHome, previousAway int
	if err := pool.QueryRow(ctx, `SELECT count(*), max(previous_home_score), max(previous_away_score) FROM match_result_changes WHERE match_id = $1`, started.Matches[1].ID).Scan(&historyCount, &previousHome, &previousAway); err != nil || historyCount != 2 || previousHome != 2 || previousAway != 1 {
		t.Fatalf("historial = count %d, previo %d-%d, %v; se esperaban dos cambios y previo 2-1", historyCount, previousHome, previousAway, err)
	}
	if _, err := service.Start(ctx, accountID, created.ID, tournaments.StartInput{RoundRobinLegs: 1}); err != tournaments.ErrTournamentConflict {
		t.Fatalf("segundo inicio = %v, se esperaba %v", err, tournaments.ErrTournamentConflict)
	}
	cancelled, err := service.Cancel(ctx, accountID, created.ID)
	if err != nil {
		t.Fatalf("cancelar liga: %v", err)
	}
	if cancelled.State != "cancelled" || len(cancelled.Teams) != 4 || len(cancelled.Matches) != 12 {
		t.Fatalf("liga cancelada = %#v, se esperaban datos conservados y estado cancelled", cancelled)
	}
	if _, err := service.Cancel(ctx, accountID, created.ID); err != tournaments.ErrTournamentCancellationConflict {
		t.Fatalf("segunda cancelación = %v, se esperaba %v", err, tournaments.ErrTournamentCancellationConflict)
	}
	if _, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, tournaments.MatchResultInput{HomeScore: 4, AwayScore: 0}); !errors.Is(err, tournaments.ErrMatchResultConflict) {
		t.Fatalf("registrar resultado tras cancelar = %v, se esperaba %v", err, tournaments.ErrMatchResultConflict)
	}
	if _, err := service.Complete(ctx, accountID, created.ID); !errors.Is(err, tournaments.ErrTournamentCompletionConflict) {
		t.Fatalf("finalizar tras cancelar = %v, se esperaba %v", err, tournaments.ErrTournamentCompletionConflict)
	}
	published, err := service.Create(ctx, accountID, tournaments.CreateInput{Name: "Liga sin empezar", Sport: tournaments.SportFootball, Teams: []tournaments.TeamInput{{Name: "Norte"}, {Name: "Sur"}}})
	if err != nil {
		t.Fatalf("crear segunda liga: %v", err)
	}
	cancelledPublished, err := service.Cancel(ctx, accountID, published.ID)
	if err != nil || cancelledPublished.State != "cancelled" {
		t.Fatalf("cancelar liga publicada = %#v, %v; se esperaba cancelled", cancelledPublished, err)
	}
}
