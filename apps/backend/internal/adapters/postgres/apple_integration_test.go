package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

type integrationAppleProvider struct {
	state, nonce string
	identity     federated.Identity
}

func (p *integrationAppleProvider) AuthorizationURL(state, nonce string) string {
	p.nonce = nonce
	p.state = state
	return "https://appleid.apple.com/auth/authorize"
}
func (p *integrationAppleProvider) Exchange(context.Context, string) (federated.Identity, error) {
	identity := p.identity
	identity.Nonce = p.nonce
	return identity, nil
}

func prepareAppleIntegration(t *testing.T, repository FederatedRepository, provider *integrationAppleProvider) (federated.AppleService, string, string) {
	t.Helper()
	service := federated.NewAppleService(repository, provider)
	proof := strings.Repeat("a", 43)
	hash := sha256.Sum256([]byte("apple-login-proof:" + proof))
	challenge, err := service.CreateChallenge(context.Background(), "android", hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(context.Background(), provider.state, "code", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Complete(context.Background(), provider.state, "code", false); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("callback replay=%v", err)
	}
	return service, challenge.ID, proof
}

func TestIntegrationAppleRegistrationSessionAndDraftAreAtomicAndProofIsSingleUse(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repository := NewFederatedRepository(pool)
	provider := &integrationAppleProvider{identity: federated.Identity{Issuer: federated.AppleIssuer, Subject: "apple-first", Email: "relay@privaterelay.appleid.com", EmailVerified: true}}
	service, id, proof := prepareAppleIntegration(t, repository, provider)
	if _, err := service.Authenticate(ctx, id, proof, nil, nil); !errors.Is(err, federated.ErrRegistration) {
		t.Fatalf("new identity=%v", err)
	}
	if _, err := service.Authenticate(ctx, id, strings.Repeat("b", 43), nil, nil); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("intercepted callback proof=%v", err)
	}
	registration := &federated.Registration{Username: "apple_person", Locale: "es", TermsVersion: legal.CurrentTermsVersion}
	draft := &federated.Draft{ID: "019abcde-1111-7111-8111-111111111151", Name: "Apple tournament", Sport: tournaments.SportFootball, Teams: []string{"A", "B"}}
	session, err := service.Authenticate(ctx, id, proof, registration, draft)
	if err != nil {
		t.Fatal(err)
	}
	account, err := NewAccountTournamentRepository(pool).Authenticate(ctx, session.AccessToken)
	if err != nil || account != session.AccountID {
		t.Fatalf("session=%v %v", account, err)
	}
	methods, err := NewAccountTournamentRepository(pool).GetAccessMethods(ctx, session.AccountID)
	if err != nil || !methods.HasApple || methods.HasGoogle || methods.HasPassword {
		t.Fatalf("methods=%#v %v", methods, err)
	}
	for _, query := range []string{
		`SELECT count(*) FROM external_identities WHERE account_id=$1 AND provider='apple'`,
		`SELECT count(*) FROM legal_account_acceptances WHERE account_id=$1 AND source='apple_registration'`,
		`SELECT count(*) FROM tournaments WHERE organizer_account_id=$1`,
	} {
		var count int
		if err := pool.QueryRow(ctx, query, session.AccountID).Scan(&count); err != nil || count != 1 {
			t.Fatalf("atomic result=%d %v", count, err)
		}
	}
	if _, err := service.Authenticate(ctx, id, proof, registration, draft); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("session replay=%v", err)
	}
	provider.identity.Email = ""
	provider.identity.EmailVerified = false
	service, id, proof = prepareAppleIntegration(t, repository, provider)
	if next, err := service.Authenticate(ctx, id, proof, nil, nil); err != nil || next.AccountID != session.AccountID {
		t.Fatalf("returning subject without email=%#v %v", next, err)
	}
}

func TestIntegrationAppleRejectsEmailConflictExpiryAndConcurrentConsumption(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	repository := NewFederatedRepository(pool)
	_ = createVerifiedLocalAccount(t, ctx, pool, "person@example.test", "person", "correct password")
	provider := &integrationAppleProvider{identity: federated.Identity{Issuer: federated.AppleIssuer, Subject: "apple-conflict", Email: "person@example.test", EmailVerified: true}}
	service, id, proof := prepareAppleIntegration(t, repository, provider)
	input := &federated.Registration{Username: "another_person", Locale: "en", TermsVersion: legal.CurrentTermsVersion}
	if _, err := service.Authenticate(ctx, id, proof, input, nil); !errors.Is(err, federated.ErrEmailConflict) {
		t.Fatalf("email conflict=%v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM external_identities WHERE provider='apple'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("automatic link count=%d %v", count, err)
	}
	provider.identity.Subject = "apple-other"
	provider.identity.Email = "other@example.test"
	service, id, proof = prepareAppleIntegration(t, repository, provider)
	var group sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		group.Go(func() { _, err := service.Authenticate(ctx, id, proof, input, nil); results <- err })
	}
	group.Wait()
	close(results)
	successes, replays := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, federated.ErrChallengeInvalid) {
			replays++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || replays != 1 {
		t.Fatalf("concurrent successes/replays=%d/%d", successes, replays)
	}
	service, id, proof = prepareAppleIntegration(t, repository, provider)
	if _, err := pool.Exec(ctx, `UPDATE apple_login_challenges SET created_at=$2, expires_at=$3 WHERE id=$1`, id, time.Now().Add(-time.Hour), time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(ctx, id, proof, nil, nil); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("expired proof=%v", err)
	}
}
