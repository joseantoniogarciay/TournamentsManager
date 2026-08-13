// Package access coordinates reauthentication and local credential changes.
package access

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

// ErrReauthenticationInvalid indicates that the credential or ticket is invalid.
var ErrReauthenticationInvalid = errors.New("reautenticación inválida")

// ErrLastAccessMethod reports an attempt to remove the sole remaining login method.
var ErrLastAccessMethod = errors.New("no se puede retirar el último método de acceso")

// Purpose binds a short-lived reauthentication ticket to one sensitive mutation.
type Purpose string

// Ticket purposes identify the only mutation a reauthentication can authorize.
const (
	SetLocalPassword    Purpose = "set-local-password" // #nosec G101 -- fixed authorization purpose, not a secret
	LinkGoogle          Purpose = "link-google"
	UnlinkGoogle        Purpose = "unlink-google"
	RemoveLocalPassword Purpose = "remove-local-password"
)

// IsPurpose reports whether a ticket purpose is accepted by the access service.
func IsPurpose(value Purpose) bool {
	return value == SetLocalPassword || value == LinkGoogle || value == UnlinkGoogle || value == RemoveLocalPassword
}

// Repository defines the persistence required to reauthenticate and change passwords.
type Repository interface {
	CurrentPasswordHash(context.Context, string) (string, error)
	CreateReauthenticationTicket(context.Context, string, []byte) error
	ConsumeReauthenticationTicketAndSetPassword(context.Context, string, []byte, string) error
	ConsumeReauthenticationTicketAndRemovePassword(context.Context, string, []byte) error
}

// Service coordinates reauthentication and local password changes.
type Service struct{ repository Repository }

// NewService builds the access service with the given persistence.
func NewService(repository Repository) Service { return Service{repository: repository} }

// ReauthenticateWithPassword validates the password and issues a short-lived ticket.
func (s Service) ReauthenticateWithPassword(ctx context.Context, sessionToken, password string, purpose Purpose) (string, string, error) {
	if !IsPurpose(purpose) {
		return "", "", ErrReauthenticationInvalid
	}
	stored, err := s.repository.CurrentPasswordHash(ctx, sessionToken)
	if err != nil || !registration.VerifyPassword(password, stored) {
		return "", "", ErrReauthenticationInvalid
	}
	ticket, hash, err := newTicket(purpose)
	if err != nil {
		return "", "", err
	}
	if err := s.repository.CreateReauthenticationTicket(ctx, sessionToken, hash); err != nil {
		return "", "", ErrReauthenticationInvalid
	}
	return ticket, time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339Nano), nil
}

// SetPassword consumes a reauthentication ticket and updates the password.
func (s Service) SetPassword(ctx context.Context, sessionToken, ticket, password string) error {
	hash := ticketHash(ticket, SetLocalPassword)
	passwordHash, err := registration.HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.repository.ConsumeReauthenticationTicketAndSetPassword(ctx, sessionToken, hash[:], passwordHash); err != nil {
		return ErrReauthenticationInvalid
	}
	return nil
}

// RemovePassword consumes a Google-authorized ticket and removes the local method.
func (s Service) RemovePassword(ctx context.Context, sessionToken, ticket string) error {
	hash := ticketHash(ticket, RemoveLocalPassword)
	if err := s.repository.ConsumeReauthenticationTicketAndRemovePassword(ctx, sessionToken, hash[:]); errors.Is(err, ErrLastAccessMethod) {
		return ErrLastAccessMethod
	} else if err != nil {
		return ErrReauthenticationInvalid
	}
	return nil
}

func newTicket(purpose Purpose) (string, []byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", nil, err
	}
	ticket := base64.RawURLEncoding.EncodeToString(secret)
	hash := ticketHash(ticket, purpose)
	return ticket, hash[:], nil
}

func ticketHash(ticket string, purpose Purpose) [32]byte {
	return sha256.Sum256([]byte("reauthentication-ticket:" + string(purpose) + ":" + ticket))
}
