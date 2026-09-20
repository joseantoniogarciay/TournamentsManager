package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationTournamentTeamInvitationCreatesTeamAndFollowWithoutAdministration(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	organizerID := createVerifiedLocalAccount(t, ctx, pool, "invite-organizer@example.test", "inviteorganizer", "correct password")
	participantID := createVerifiedLocalAccount(t, ctx, pool, "invite-participant@example.test", "inviteparticipant", "correct password")
	service := tournaments.NewCreationService(NewAccountTournamentRepository(pool))
	created, err := service.Create(ctx, organizerID, tournaments.CreateInput{
		Name: "Copa por invitación", Sport: tournaments.SportFootball,
		Teams: []tournaments.TeamInput{{Name: "Organizadores"}},
	})
	if err != nil || len(created.Teams) != 1 {
		t.Fatalf("crear torneo con equipo propio = %#v, %v", created, err)
	}
	var organizerPreference string
	if err := pool.QueryRow(ctx, `SELECT last_team_name FROM accounts WHERE id = $1`, organizerID).Scan(&organizerPreference); err != nil || organizerPreference != "Organizadores" {
		t.Fatalf("preferencia organizadora = %q, %v", organizerPreference, err)
	}
	firstToken, err := service.CreateTeamInvitation(ctx, organizerID, created.ID)
	if err != nil {
		t.Fatalf("crear primera invitación = %v", err)
	}
	secondToken, err := service.CreateTeamInvitation(ctx, organizerID, created.ID)
	if err != nil {
		t.Fatalf("rotar invitación = %v", err)
	}
	if _, err := service.InspectTeamInvitation(ctx, firstToken); !errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
		t.Fatalf("inspeccionar invitación rotada = %v; se esperaba no disponible", err)
	}
	invitation, err := service.InspectTeamInvitation(ctx, secondToken)
	if err != nil || invitation.TournamentID != created.ID || invitation.TournamentName != created.Name {
		t.Fatalf("inspeccionar invitación = %#v, %v", invitation, err)
	}
	joined, err := service.JoinTeamInvitation(ctx, participantID, secondToken, tournaments.TeamInput{Name: " Invitados "})
	if err != nil || joined.TournamentID != created.ID || joined.Team.Name != "Invitados" {
		t.Fatalf("registrar equipo invitado = %#v, %v", joined, err)
	}
	if _, err := service.JoinTeamInvitation(ctx, participantID, secondToken, tournaments.TeamInput{Name: "Otro"}); !errors.Is(err, tournaments.ErrTournamentInvitationConflict) {
		t.Fatalf("segunda inscripción de la cuenta = %v; se esperaba conflicto", err)
	}
	var follows, administers, ownsTeam bool
	var participantPreference string
	if err := pool.QueryRow(ctx, `SELECT
		EXISTS (SELECT 1 FROM tournament_followers WHERE tournament_id=$1 AND account_id=$2),
		EXISTS (SELECT 1 FROM tournament_administrators WHERE tournament_id=$1 AND account_id=$2),
		EXISTS (SELECT 1 FROM tournament_team_accounts WHERE tournament_id=$1 AND team_id=$3 AND account_id=$2),
		(SELECT last_team_name FROM accounts WHERE id=$2)`, created.ID, participantID, joined.Team.ID).Scan(&follows, &administers, &ownsTeam, &participantPreference); err != nil {
		t.Fatalf("comprobar relaciones = %v", err)
	}
	if !follows || administers || !ownsTeam || participantPreference != "Invitados" {
		t.Fatalf("relaciones = follows %v, administers %v, owns team %v, preferencia %q", follows, administers, ownsTeam, participantPreference)
	}
	if _, err := service.Start(ctx, organizerID, created.ID, tournaments.StartInput{RoundRobinLegs: 1}); err != nil {
		t.Fatalf("iniciar torneo = %v", err)
	}
	if _, err := service.InspectTeamInvitation(ctx, secondToken); !errors.Is(err, tournaments.ErrTournamentInvitationNotFound) {
		t.Fatalf("inspeccionar invitación tras inicio = %v; se esperaba no disponible", err)
	}
	repository := NewAccountTournamentRepository(pool)
	if _, err := repository.ScheduleAccountDeletion(ctx, participantID); err != nil {
		t.Fatalf("programar baja de participante = %v", err)
	}
	var linkedAfterDeletion, followsAfterDeletion, teamRemains bool
	if err := pool.QueryRow(ctx, `SELECT
		EXISTS (SELECT 1 FROM tournament_team_accounts WHERE tournament_id=$1 AND account_id=$2),
		EXISTS (SELECT 1 FROM tournament_followers WHERE tournament_id=$1 AND account_id=$2),
		EXISTS (SELECT 1 FROM tournament_teams WHERE tournament_id=$1 AND id=$3)`, created.ID, participantID, joined.Team.ID).Scan(&linkedAfterDeletion, &followsAfterDeletion, &teamRemains); err != nil {
		t.Fatalf("comprobar baja de participante = %v", err)
	}
	if linkedAfterDeletion || followsAfterDeletion || !teamRemains {
		t.Fatalf("baja = vínculo %v, seguimiento %v, equipo %v; se esperaban relaciones retiradas y equipo conservado", linkedAfterDeletion, followsAfterDeletion, teamRemains)
	}
}
