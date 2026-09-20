package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

func TestIntegrationPasswordResetConsumesTokenRevokesSessionsAndCreatesNewSession(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "old correct password")
	if _, err := pool.Exec(ctx, `INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at) VALUES ($1, decode(repeat('01', 32), 'hex'), now() + interval '1 day', now() + interval '1 day')`, accountID); err != nil {
		t.Fatalf("crear sesión previa: %v", err)
	}
	mailer := &integrationMailer{}
	service := registration.NewService(NewRegistrationRepository(pool), mailer)
	if err := service.RequestPasswordReset(ctx, " person@example.test "); err != nil {
		t.Fatalf("solicitar restablecimiento: %v", err)
	}
	if mailer.passwordResetToken == "" {
		t.Fatal("no se entregó el token de restablecimiento")
	}
	if email, err := service.InspectPasswordReset(ctx, mailer.passwordResetToken); err != nil || email != "person@example.test" {
		t.Fatalf("inspeccionar enlace = (%q, %v), se esperaba email válido", email, err)
	}
	if _, _, _, err := service.ResetPassword(ctx, mailer.passwordResetToken, "new correct password"); err != nil {
		t.Fatalf("restablecer contraseña: %v", err)
	}
	var passwordHash string
	if err := pool.QueryRow(ctx, `SELECT password_hash FROM local_credentials WHERE account_id = $1`, accountID).Scan(&passwordHash); err != nil {
		t.Fatalf("leer credencial cambiada: %v", err)
	}
	if !registration.VerifyPassword("new correct password", passwordHash) || registration.VerifyPassword("old correct password", passwordHash) {
		t.Fatal("la credencial persistida no contiene únicamente la nueva contraseña")
	}
	var activeSessions, revokedSessions int
	if err := pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE revoked_at IS NULL), count(*) FILTER (WHERE revoked_at IS NOT NULL) FROM sessions WHERE account_id = $1`, accountID).Scan(&activeSessions, &revokedSessions); err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}
	if activeSessions != 1 || revokedSessions != 1 {
		t.Fatalf("sesiones activas/revocadas = %d/%d, se esperaba 1/1", activeSessions, revokedSessions)
	}
	if _, err := service.InspectPasswordReset(ctx, mailer.passwordResetToken); !errors.Is(err, registration.ErrPasswordResetInvalid) {
		t.Fatalf("inspeccionar token consumido = %v, se esperaba %v", err, registration.ErrPasswordResetInvalid)
	}
	if _, _, _, err := service.ResetPassword(ctx, mailer.passwordResetToken, "another correct password"); !errors.Is(err, registration.ErrPasswordResetInvalid) {
		t.Fatalf("reutilizar token = %v, se esperaba %v", err, registration.ErrPasswordResetInvalid)
	}
}

func TestIntegrationLocalLoginCreatesTournamentAndSessionAtomically(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "login-draft@example.test", "login_draft", "correct password")
	service := registration.NewService(NewRegistrationRepository(pool), nil)

	draft := &registration.Draft{
		ID:   "019abcde-1111-7111-8111-111111111112",
		Name: "Torneo recuperado", Sport: tournaments.SportBasketball, Teams: []string{"Norte", "Sur"},
	}
	result, err := service.Login(ctx, "login-draft@example.test", "correct password", draft)
	if err != nil || result.Session.AccountID != accountID || result.Session.LastTeamName != "Norte" {
		t.Fatalf("login con borrador = %#v, %v", result, err)
	}
	if _, err := service.Login(ctx, "login-draft@example.test", "correct password", draft); err != nil {
		t.Fatalf("reintentar login con el mismo borrador: %v", err)
	}
	var sessions, tournamentsCount, teams int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE account_id = $1 AND revoked_at IS NULL`, accountID).Scan(&sessions); err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournaments WHERE organizer_account_id = $1 AND name = 'Torneo recuperado' AND sport = 'basketball'`, accountID).Scan(&tournamentsCount); err != nil {
		t.Fatalf("contar torneos: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournament_teams JOIN tournaments ON tournaments.id = tournament_teams.tournament_id WHERE tournaments.organizer_account_id = $1`, accountID).Scan(&teams); err != nil {
		t.Fatalf("contar equipos: %v", err)
	}
	if sessions != 2 || tournamentsCount != 1 || teams != 2 {
		t.Fatalf("sesiones/torneos/equipos = %d/%d/%d, se esperaba 2/1/2", sessions, tournamentsCount, teams)
	}
}

func TestIntegrationGoogleLoginCreatesTournamentAndSessionAtomically(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "google-draft@example.test", "google_draft", "correct password")
	if _, err := pool.Exec(ctx, `INSERT INTO external_identities (account_id, provider, issuer, subject) VALUES ($1, 'google', $2, 'draft-subject')`, accountID, federated.GoogleIssuer); err != nil {
		t.Fatalf("crear identidad Google: %v", err)
	}
	verifier := &integrationGoogleVerifier{}
	service := federated.NewService(NewFederatedRepository(pool), verifier)
	challenge, err := service.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge: %v", err)
	}
	verifier.identity = federated.Identity{Issuer: federated.GoogleIssuer, Subject: "draft-subject", Email: "google-draft@example.test", Nonce: challenge.Nonce, EmailVerified: true}

	draft := &federated.Draft{
		ID:   "019abcde-1111-7111-8111-111111111112",
		Name: "Torneo Google", Sport: tournaments.SportFootball, Teams: []string{"Uno", "Dos"},
	}
	result, err := service.Authenticate(ctx, challenge.ID, "google-token", nil, draft)
	if err != nil || result.AccountID != accountID || result.LastTeamName != "Uno" {
		t.Fatalf("login Google con borrador = %#v, %v", result, err)
	}
	retryChallenge, err := service.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge de reintento: %v", err)
	}
	verifier.identity.Nonce = retryChallenge.Nonce
	if _, err := service.Authenticate(ctx, retryChallenge.ID, "google-token", nil, draft); err != nil {
		t.Fatalf("reintentar login Google con el mismo borrador: %v", err)
	}
	var sessions, tournamentsCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE account_id = $1 AND revoked_at IS NULL`, accountID).Scan(&sessions); err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournaments WHERE organizer_account_id = $1 AND name = 'Torneo Google'`, accountID).Scan(&tournamentsCount); err != nil {
		t.Fatalf("contar torneos: %v", err)
	}
	if sessions != 2 || tournamentsCount != 1 {
		t.Fatalf("sesiones/torneos = %d/%d, se esperaba 2/1", sessions, tournamentsCount)
	}
}

func TestIntegrationLoginDraftFailureRollsBackSession(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "rollback-draft@example.test", "rollback_draft", "correct password")
	repository := NewRegistrationRepository(pool)

	_, err := repository.CreateLocalLoginSession(ctx, accountID, sessionHash("rollback-session"), sessionHash("rollback-refresh"), &registration.Draft{
		ID:   "019abcde-1111-7111-8111-111111111112",
		Name: "Torneo inválido", Sport: tournaments.SportFootball, Teams: []string{"Duplicado", "Duplicado"},
	})
	if err == nil {
		t.Fatal("el borrador inválido no falló")
	}
	var sessions, tournamentsCount int
	var lastTeamName *string
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE account_id = $1`, accountID).Scan(&sessions); err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournaments WHERE organizer_account_id = $1`, accountID).Scan(&tournamentsCount); err != nil {
		t.Fatalf("contar torneos: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT last_team_name FROM accounts WHERE id = $1`, accountID).Scan(&lastTeamName); err != nil {
		t.Fatalf("consultar preferencia: %v", err)
	}
	if sessions != 0 || tournamentsCount != 0 || lastTeamName != nil {
		t.Fatalf("rollback dejó sesiones/torneos/preferencia = %d/%d/%v", sessions, tournamentsCount, lastTeamName)
	}
}

func TestIntegrationGoogleLoginDraftFailureRollsBackChallengeAndSession(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "google-rollback@example.test", "google_rollback", "correct password")
	if _, err := pool.Exec(ctx, `INSERT INTO external_identities (account_id, provider, issuer, subject) VALUES ($1, 'google', $2, 'rollback-subject')`, accountID, federated.GoogleIssuer); err != nil {
		t.Fatalf("crear identidad Google: %v", err)
	}
	verifier := &integrationGoogleVerifier{}
	service := federated.NewService(NewFederatedRepository(pool), verifier)
	challenge, err := service.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge: %v", err)
	}
	verifier.identity = federated.Identity{Issuer: federated.GoogleIssuer, Subject: "rollback-subject", Email: "google-rollback@example.test", Nonce: challenge.Nonce, EmailVerified: true}

	_, err = service.Authenticate(ctx, challenge.ID, "google-token", nil, &federated.Draft{
		ID:   "019abcde-1111-7111-8111-111111111112",
		Name: "Torneo inválido", Sport: tournaments.SportFootball, Teams: []string{"Duplicado", "Duplicado"},
	})
	if err == nil {
		t.Fatal("el borrador Google inválido no falló")
	}
	var sessions, tournamentsCount int
	var challengeConsumed bool
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE account_id = $1`, accountID).Scan(&sessions); err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tournaments WHERE organizer_account_id = $1`, accountID).Scan(&tournamentsCount); err != nil {
		t.Fatalf("contar torneos: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT consumed_at IS NOT NULL FROM federated_login_challenges WHERE id = $1`, challenge.ID).Scan(&challengeConsumed); err != nil {
		t.Fatalf("consultar challenge: %v", err)
	}
	if sessions != 0 || tournamentsCount != 0 || challengeConsumed {
		t.Fatalf("rollback dejó sesiones/torneos/challenge consumido = %d/%d/%v", sessions, tournamentsCount, challengeConsumed)
	}
}
