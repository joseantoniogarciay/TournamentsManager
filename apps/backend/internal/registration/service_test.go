package registration

import (
	"context"
	"errors"
	"testing"
)

func TestResetPasswordUsesPasswordProtector(t *testing.T) {
	t.Parallel()

	protector := &passwordProtectorStub{}
	repository := &passwordResetRepositoryStub{}
	service := NewService(repository, nil, protector)

	session, access, refresh, err := service.ResetPassword(context.Background(), "valid-reset-token", "new correct password")
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if !protector.hashed {
		t.Fatal("ResetPassword() did not hash through PasswordProtector")
	}
	if !repository.consumed {
		t.Fatal("ResetPassword() did not consume the reset token")
	}
	if session.AccountID != "account-id" || access == "" || refresh == "" {
		t.Fatalf("ResetPassword() = %#v, %q, %q", session, access, refresh)
	}
}

func TestRefreshPreservesTechnicalRepositoryFailures(t *testing.T) {
	want := errors.New("database unavailable")
	service := NewService(&refreshRepositoryStub{err: want}, nil)

	_, _, _, err := service.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, want) {
		t.Fatalf("Refresh() error = %v, want wrapped %v", err, want)
	}
}

type passwordProtectorStub struct{ hashed bool }

func (p *passwordProtectorStub) Hash(context.Context, string) (string, error) {
	p.hashed = true
	return "password-hash", nil
}

func (*passwordProtectorStub) Verify(context.Context, string, string) bool { return false }

type passwordResetRepositoryStub struct {
	Repository
	consumed bool
}

type refreshRepositoryStub struct {
	Repository
	err error
}

func (r *refreshRepositoryStub) RotateSessionTokens(context.Context, []byte, []byte, []byte) (Session, error) {
	return Session{}, r.err
}

func (r *passwordResetRepositoryStub) ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (Session, error) {
	r.consumed = true
	return Session{AccountID: "account-id", Username: "person"}, nil
}
