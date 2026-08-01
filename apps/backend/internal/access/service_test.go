package access

import (
	"context"
	"errors"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type repositoryStub struct {
	passwordHash      string
	created, consumed bool
}

func (r *repositoryStub) CurrentPasswordHash(context.Context, string) (string, error) {
	return r.passwordHash, nil
}
func (r *repositoryStub) CreateReauthenticationTicket(context.Context, string, []byte) error {
	r.created = true
	return nil
}
func (r *repositoryStub) ConsumeReauthenticationTicketAndSetPassword(context.Context, string, []byte, string) error {
	r.consumed = true
	return nil
}

func TestReauthenticateWithPasswordVerifiesArgon2id(t *testing.T) {
	hash, err := registration.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{passwordHash: hash}
	service := NewService(repository)
	if _, _, err := service.ReauthenticateWithPassword(context.Background(), "session", "wrong"); !errors.Is(err, ErrReauthenticationInvalid) {
		t.Fatalf("wrong password error = %v", err)
	}
	if repository.created {
		t.Fatal("ticket created for wrong password")
	}
	if _, _, err := service.ReauthenticateWithPassword(context.Background(), "session", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if !repository.created {
		t.Fatal("ticket was not created")
	}
}

func TestSetPasswordConsumesTicketOnce(t *testing.T) {
	service := NewService(&repositoryStub{})
	if err := service.SetPassword(context.Background(), "session", "ticket", "new correct password"); err != nil {
		t.Fatal(err)
	}
}
