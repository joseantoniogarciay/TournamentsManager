package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

func TestIntegrationAccountOptionsRequireSingleUseReauthenticationTicket(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "old correct password")
	methods, err := NewAccountTournamentRepository(pool).GetAccessMethods(ctx, accountID)
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
	service := access.NewService(NewAccountTournamentRepository(pool))
	ticket, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "old correct password", access.SetLocalPassword)
	if err != nil {
		t.Fatalf("reautenticar: %v", err)
	}
	if err := service.SetPassword(ctx, sessionToken, ticket, "new correct password"); err != nil {
		t.Fatalf("cambiar contraseña: %v", err)
	}
	if err := service.SetPassword(ctx, sessionToken, ticket, "another correct password"); !errors.Is(err, access.ErrReauthenticationInvalid) {
		t.Fatalf("reutilizar ticket = %v, se esperaba %v", err, access.ErrReauthenticationInvalid)
	}
	if _, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "old correct password", access.SetLocalPassword); !errors.Is(err, access.ErrReauthenticationInvalid) {
		t.Fatalf("reautenticar con contraseña previa = %v, se esperaba %v", err, access.ErrReauthenticationInvalid)
	}
	if _, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "new correct password", access.SetLocalPassword); err != nil {
		t.Fatalf("reautenticar con contraseña cambiada: %v", err)
	}
	verifier := &integrationGoogleVerifier{}
	google := federated.NewService(NewFederatedRepository(pool), verifier)
	localTicket, _, err := service.ReauthenticateWithPassword(ctx, sessionToken, "new correct password", access.LinkGoogle)
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
	methods, err = NewAccountTournamentRepository(pool).GetAccessMethods(ctx, accountID)
	if err != nil || !methods.HasGoogle || !methods.HasPassword {
		t.Fatalf("métodos tras vincular Google = %#v, %v; se esperaban contraseña y Google", methods, err)
	}
	reauthenticationChallenge, err := google.CreateChallenge(ctx)
	if err != nil {
		t.Fatalf("crear challenge para reautenticar con Google: %v", err)
	}
	verifier.identity.Nonce = reauthenticationChallenge.Nonce
	googleTicket, _, err := google.ReauthenticateGoogle(ctx, accountID, sessionToken, reauthenticationChallenge.ID, "google-id-token", string(access.RemoveLocalPassword))
	if err != nil {
		t.Fatalf("reautenticar con Google vinculada: %v", err)
	}
	if err := service.RemovePassword(ctx, sessionToken, googleTicket); err != nil {
		t.Fatalf("eliminar contraseña con ticket Google: %v", err)
	}
	if err := service.RemovePassword(ctx, sessionToken, googleTicket); !errors.Is(err, access.ErrReauthenticationInvalid) {
		t.Fatalf("reutilizar ticket Google = %v, se esperaba %v", err, access.ErrReauthenticationInvalid)
	}
}
