package registration

import (
	"context"
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

func (r *passwordResetRepositoryStub) ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (Session, error) {
	r.consumed = true
	return Session{AccountID: "account-id", Username: "person"}, nil
}
