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
