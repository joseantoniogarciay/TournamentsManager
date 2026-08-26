package federated

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubVerifier struct {
	identity Identity
	err      error
}

func (v stubVerifier) Verify(context.Context, string) (Identity, error) { return v.identity, v.err }

type stubRepository struct {
	addErr                        error
	authenticateCalled, addCalled bool
	riscCalled                    bool
}

func (r *stubRepository) CreateChallenge(context.Context, []byte, time.Time) (string, error) {
	return "019abcde-1111-7111-8111-111111111111", nil
}
func (r *stubRepository) AuthenticateGoogle(context.Context, string, []byte, Identity, *Registration, []byte, []byte) (Session, error) {
	r.authenticateCalled = true
	return Session{AccountID: "account", Username: "person"}, nil
}
func (r *stubRepository) AddGoogleIdentity(context.Context, string, string, []byte, Identity) error {
	r.addCalled = true
	return r.addErr
}
func (*stubRepository) ReauthenticateGoogle(context.Context, string, string, string, []byte, Identity, []byte) error {
	return nil
}
func (*stubRepository) AddGoogleIdentityWithTicket(context.Context, string, string, []byte, Identity, []byte) error {
	return nil
}
func (*stubRepository) RemoveGoogleIdentityWithTicket(context.Context, string, []byte) error {
	return nil
}
func (r *stubRepository) RevokeSessionsForGoogleIdentity(context.Context, RISCEvent) error {
	r.riscCalled = true
	return nil
}

func TestAuthenticateRejectsUnverifiedGoogleEmailBeforeRepository(t *testing.T) {
	repository := &stubRepository{}
	service := NewService(repository, stubVerifier{identity: Identity{Issuer: GoogleIssuer, Subject: "subject", Email: "person@example.test", Nonce: "nonce"}})
	_, err := service.Authenticate(context.Background(), "019abcde-1111-7111-8111-111111111111", "token", nil)
	if !errors.Is(err, ErrChallengeInvalid) {
		t.Fatalf("Authenticate() error = %v, want challenge invalid", err)
	}
	if repository.authenticateCalled {
		t.Fatal("repository was called for an unverified email")
	}
}

func TestAuthenticateRejectsInvalidGoogleIdentityBeforeRepository(t *testing.T) {
	valid := Identity{Issuer: GoogleIssuer, Subject: "subject", Email: "person@example.test", Nonce: "nonce", EmailVerified: true}
	for _, test := range []struct {
		name     string
		identity Identity
	}{
		{"issuer", Identity{Issuer: "https://issuer.example", Subject: valid.Subject, Email: valid.Email, Nonce: valid.Nonce, EmailVerified: true}},
		{"subject", Identity{Issuer: valid.Issuer, Email: valid.Email, Nonce: valid.Nonce, EmailVerified: true}},
		{"email", Identity{Issuer: valid.Issuer, Subject: valid.Subject, Email: " ", Nonce: valid.Nonce, EmailVerified: true}},
		{"nonce", Identity{Issuer: valid.Issuer, Subject: valid.Subject, Email: valid.Email, EmailVerified: true}},
		{"email not verified", Identity{Issuer: valid.Issuer, Subject: valid.Subject, Email: valid.Email, Nonce: valid.Nonce}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &stubRepository{}
			service := NewService(repository, stubVerifier{identity: test.identity})
			_, err := service.Authenticate(context.Background(), "019abcde-1111-7111-8111-111111111111", "token", nil)
			if !errors.Is(err, ErrChallengeInvalid) {
				t.Fatalf("Authenticate() error = %v, want %v", err, ErrChallengeInvalid)
			}
			if repository.authenticateCalled {
				t.Fatal("repository was called for an invalid identity")
			}
		})
	}
}

func TestAddGooglePreservesForeignSubjectConflict(t *testing.T) {
	repository := &stubRepository{addErr: ErrIdentityConflict}
	service := NewService(repository, stubVerifier{identity: Identity{Issuer: GoogleIssuer, Subject: "foreign-subject", Email: "person@example.test", Nonce: "nonce", EmailVerified: true}})
	err := service.AddGoogle(context.Background(), "019abcde-1111-7111-8111-111111111111", "019abcde-2222-7222-8222-222222222222", "token")
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("AddGoogle() error = %v, want identity conflict", err)
	}
	if !repository.addCalled {
		t.Fatal("repository was not called")
	}
}

func TestHandleRISCEventRevokesOnlyRequiredSecurityEvents(t *testing.T) {
	for _, test := range []struct {
		name, eventType, reason string
		want                    bool
	}{
		{"sessions revoked", RISCSessionsRevoked, "", true},
		{"tokens revoked", RISCTokensRevoked, "", true},
		{"hijacked account", RISCAccountDisabled, "hijacking", true},
		{"bulk account", RISCAccountDisabled, "bulk-account", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &stubRepository{}
			service := NewService(repository, stubVerifier{})
			err := service.HandleRISCEvent(context.Background(), RISCEvent{ID: "event", Issuer: GoogleIssuer, Subject: "subject", Type: test.eventType, Reason: test.reason})
			if err != nil || repository.riscCalled != test.want {
				t.Fatalf("HandleRISCEvent() = %v, called = %t", err, repository.riscCalled)
			}
		})
	}
}
