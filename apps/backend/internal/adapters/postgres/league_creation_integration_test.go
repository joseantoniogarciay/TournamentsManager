package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type integrationMailer struct{ passwordResetToken string }

type integrationGoogleVerifier struct{ identity federated.Identity }

func (v *integrationGoogleVerifier) Verify(context.Context, string) (federated.Identity, error) {
	return v.identity, nil
}

func (*integrationMailer) SendVerification(context.Context, string, registration.Locale, string) error {
	return nil
}

func (m *integrationMailer) SendPasswordReset(_ context.Context, _ string, _ registration.Locale, token string) error {
	m.passwordResetToken = token
	return nil
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TM_INTEGRATION_DATABASE_URL")
	if databaseURL == "" || os.Getenv("TM_RUN_INTEGRATION") != "1" {
		t.Skip("la integración PostgreSQL no está activada")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("conectar PostgreSQL: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(context.Background(), "TRUNCATE accounts CASCADE"); err != nil {
		t.Fatalf("limpiar base: %v", err)
	}
	return pool
}

func createVerifiedLocalAccount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email, username, password string) string {
	t.Helper()
	passwordHash, err := registration.HashPassword(password)
	if err != nil {
		t.Fatalf("crear hash de contraseña: %v", err)
	}
	var accountID string
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, verified_at) VALUES ($1, 'es', 'verified', $2, now()) RETURNING id::text`, email, username).Scan(&accountID); err != nil {
		t.Fatalf("crear cuenta: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO local_credentials (account_id, password_hash) VALUES ($1, $2)`, accountID, passwordHash); err != nil {
		t.Fatalf("crear credencial: %v", err)
	}
	return accountID
}

func sessionHash(token string) []byte {
	hash := sha256.Sum256([]byte("session:" + token))
	return hash[:]
}

// Esta prueba usa una base efímera preparada por el comando de integración.
func TestIntegrationLeagueCreationAndStartWithPostgres(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	var accountID string
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, verified_at) VALUES ('organizer@example.test', 'es', 'verified', 'organizer', now()) RETURNING id::text`).Scan(&accountID); err != nil {
		t.Fatalf("crear organizadora: %v", err)
	}
	service := leagues.NewCreationService(NewAccountLeagueRepository(pool))
	created, err := service.Create(ctx, accountID, leagues.CreateInput{Name: "Liga de verano", Teams: []leagues.TeamInput{{Name: "Azules"}, {Name: "Rojos"}, {Name: "Verdes"}, {Name: "Amarillos"}}})
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
	var assigned bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM league_administrators WHERE league_id = $1 AND account_id = $2)`, created.ID, administratorID).Scan(&assigned); err != nil || !assigned {
		t.Fatalf("comprobar administradora asignada = %v, %v", assigned, err)
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
	withOrganizerResult, err := service.RecordResult(ctx, accountID, created.ID, started.Matches[0].ID, leagues.MatchResultInput{HomeScore: 2, AwayScore: 1})
	if err != nil || withOrganizerResult.Matches[0].State != "completed" || withOrganizerResult.Matches[0].HomeScore == nil || *withOrganizerResult.Matches[0].HomeScore != 2 || withOrganizerResult.Matches[0].AwayScore == nil || *withOrganizerResult.Matches[0].AwayScore != 1 {
		t.Fatalf("resultado de organizadora = %#v, %v; se esperaba marcador 2-1", withOrganizerResult.Matches[0], err)
	}
	withResult, err := service.RecordResult(ctx, administratorID, created.ID, started.Matches[1].ID, leagues.MatchResultInput{HomeScore: 2, AwayScore: 1})
	if err != nil || withResult.Matches[0].State != "completed" || withResult.Matches[0].HomeScore == nil || *withResult.Matches[0].HomeScore != 2 || withResult.Matches[0].AwayScore == nil || *withResult.Matches[0].AwayScore != 1 {
		t.Fatalf("registrar resultado = %#v, %v; se esperaba marcador 2-1", withResult.Matches[0], err)
	}
	corrected, err := service.RecordResult(ctx, administratorID, created.ID, started.Matches[1].ID, leagues.MatchResultInput{HomeScore: 3, AwayScore: 0})
	if err != nil || corrected.Matches[0].HomeScore == nil || *corrected.Matches[0].HomeScore != 3 || corrected.Matches[0].AwayScore == nil || *corrected.Matches[0].AwayScore != 0 {
		t.Fatalf("corregir resultado = %#v, %v; se esperaba marcador 3-0", corrected.Matches[0], err)
	}
	var historyCount, previousHome, previousAway int
	if err := pool.QueryRow(ctx, `SELECT count(*), max(previous_home_score), max(previous_away_score) FROM match_result_changes WHERE match_id = $1`, started.Matches[1].ID).Scan(&historyCount, &previousHome, &previousAway); err != nil || historyCount != 2 || previousHome != 2 || previousAway != 1 {
		t.Fatalf("historial = count %d, previo %d-%d, %v; se esperaban dos cambios y previo 2-1", historyCount, previousHome, previousAway, err)
	}
	if _, err := service.Start(ctx, accountID, created.ID, leagues.StartInput{RoundRobinLegs: 1}); err != leagues.ErrLeagueConflict {
		t.Fatalf("segundo inicio = %v, se esperaba %v", err, leagues.ErrLeagueConflict)
	}
	cancelled, err := service.Cancel(ctx, accountID, created.ID)
	if err != nil {
		t.Fatalf("cancelar liga: %v", err)
	}
	if cancelled.State != "cancelled" || len(cancelled.Teams) != 4 || len(cancelled.Matches) != 12 {
		t.Fatalf("liga cancelada = %#v, se esperaban datos conservados y estado cancelled", cancelled)
	}
	if _, err := service.Cancel(ctx, accountID, created.ID); err != leagues.ErrLeagueCancellationConflict {
		t.Fatalf("segunda cancelación = %v, se esperaba %v", err, leagues.ErrLeagueCancellationConflict)
	}
	published, err := service.Create(ctx, accountID, leagues.CreateInput{Name: "Liga sin empezar", Teams: []leagues.TeamInput{{Name: "Norte"}, {Name: "Sur"}}})
	if err != nil {
		t.Fatalf("crear segunda liga: %v", err)
	}
	cancelledPublished, err := service.Cancel(ctx, accountID, published.ID)
	if err != nil || cancelledPublished.State != "cancelled" {
		t.Fatalf("cancelar liga publicada = %#v, %v; se esperaba cancelled", cancelledPublished, err)
	}
}

func TestIntegrationRecentLeaguesOrdersActivityAndDeduplicatesRelationships(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "correct password")
	otherAccountID := createVerifiedLocalAccount(t, ctx, pool, "other@example.test", "other", "correct password")
	var administeredID, followedID string
	if err := pool.QueryRow(ctx, `INSERT INTO leagues (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Administrada', now(), now() - interval '2 hours') RETURNING id::text`, accountID).Scan(&administeredID); err != nil {
		t.Fatalf("crear liga administrada: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO leagues (organizer_account_id, name, published_at, last_activity_at) VALUES ($1, 'Seguida', now(), now() - interval '1 hour') RETURNING id::text`, otherAccountID).Scan(&followedID); err != nil {
		t.Fatalf("crear liga seguida: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO league_followers (league_id, account_id) VALUES ($1, $2), ($3, $2)`, administeredID, accountID, followedID); err != nil {
		t.Fatalf("seguir ligas: %v", err)
	}

	items, err := leagues.NewService(NewAccountLeagueRepository(pool)).ListRecent(ctx, accountID)
	if err != nil {
		t.Fatalf("consultar recientes: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ligas recientes = %#v, se esperaban dos sin duplicados", items)
	}
	if items[0].ID != followedID || items[0].Relationship != "follower" {
		t.Fatalf("primera liga = %#v, se esperaba la seguida más reciente", items[0])
	}
	if items[1].ID != administeredID || items[1].Relationship != "organizer" {
		t.Fatalf("segunda liga = %#v, se esperaba la administrada una sola vez", items[1])
	}
}

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

func TestIntegrationAccountOptionsRequireSingleUseReauthenticationTicket(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "old correct password")
	methods, err := NewAccountLeagueRepository(pool).GetAccessMethods(ctx, accountID)
	if err != nil {
		t.Fatalf("consultar métodos de acceso: %v", err)
	}
	if methods.Email != "person@example.test" || methods.Username != "person" || !methods.HasPassword || methods.HasGoogle {
		t.Fatalf("métodos de acceso = %#v, se esperaba solo contraseña", methods)
	}
	const sessionToken = "current-session-token"
	if _, err := pool.Exec(ctx, `INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at) VALUES ($1, $2, now() + interval '1 day', now() + interval '1 day')`, accountID, sessionHash(sessionToken)); err != nil {
		t.Fatalf("crear sesión: %v", err)
	}
	service := access.NewService(NewAccountLeagueRepository(pool))
	ticket, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "old correct password")
	if err != nil {
		t.Fatalf("reautenticar: %v", err)
	}
	if err := service.SetPassword(ctx, sessionToken, ticket, "new correct password"); err != nil {
		t.Fatalf("cambiar contraseña: %v", err)
	}
	if err := service.SetPassword(ctx, sessionToken, ticket, "another correct password"); !errors.Is(err, access.ErrReauthenticationInvalid) {
		t.Fatalf("reutilizar ticket = %v, se esperaba %v", err, access.ErrReauthenticationInvalid)
	}
	if _, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "old correct password"); !errors.Is(err, access.ErrReauthenticationInvalid) {
		t.Fatalf("reautenticar con contraseña previa = %v, se esperaba %v", err, access.ErrReauthenticationInvalid)
	}
	if _, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "new correct password"); err != nil {
		t.Fatalf("reautenticar con contraseña cambiada: %v", err)
	}
	verifier := &integrationGoogleVerifier{}
	google := federated.NewService(NewFederatedRepository(pool), verifier)
	localTicket, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "new correct password")
	if err != nil {
		t.Fatalf("reautenticar para vincular Google: %v", err)
	}
	linkChallenge, err := google.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge para vincular Google: %v", err)
	}
	verifier.identity = federated.Identity{Issuer: federated.GoogleIssuer, Subject: "google-subject", Email: "person@example.test", Nonce: linkChallenge.Nonce, EmailVerified: true}
	if err := google.AddGoogleWithTicket(ctx, sessionToken, localTicket, linkChallenge.ID, "google-id-token"); err != nil {
		t.Fatalf("vincular Google: %v", err)
	}
	methods, err = NewAccountLeagueRepository(pool).GetAccessMethods(ctx, accountID)
	if err != nil || !methods.HasGoogle || !methods.HasPassword {
		t.Fatalf("métodos tras vincular Google = %#v, %v; se esperaban contraseña y Google", methods, err)
	}
	reauthenticationChallenge, err := google.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge para reautenticar con Google: %v", err)
	}
	verifier.identity.Nonce = reauthenticationChallenge.Nonce
	googleTicket, _, err := google.ReauthenticateGoogle(ctx, accountID, sessionToken, reauthenticationChallenge.ID, "google-id-token")
	if err != nil {
		t.Fatalf("reautenticar con Google vinculada: %v", err)
	}
	reuseChallenge, err := google.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge para consumir ticket Google: %v", err)
	}
	verifier.identity.Nonce = reuseChallenge.Nonce
	if err := google.AddGoogleWithTicket(ctx, sessionToken, googleTicket, reuseChallenge.ID, "google-id-token"); err != nil {
		t.Fatalf("consumir ticket Google: %v", err)
	}
	reusedTicketChallenge, err := google.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge para reusar ticket: %v", err)
	}
	verifier.identity.Nonce = reusedTicketChallenge.Nonce
	if err := google.AddGoogleWithTicket(ctx, sessionToken, googleTicket, reusedTicketChallenge.ID, "google-id-token"); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("reutilizar ticket Google = %v, se esperaba %v", err, federated.ErrChallengeInvalid)
	}
}
