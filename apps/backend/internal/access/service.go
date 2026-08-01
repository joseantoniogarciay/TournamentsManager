// Package access coordina la reautenticación y los cambios de credenciales locales.
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

// ErrReauthenticationInvalid indica que la credencial o el ticket no son válidos.
var ErrReauthenticationInvalid = errors.New("reautenticación inválida")

// Repository define la persistencia necesaria para reautenticar y cambiar contraseñas.
type Repository interface {
	CurrentPasswordHash(context.Context, string) (string, error)
	CreateReauthenticationTicket(context.Context, string, []byte) error
	ConsumeReauthenticationTicketAndSetPassword(context.Context, string, []byte, string) error
}

// Service coordina la reautenticación y los cambios de contraseña locales.
type Service struct{ repository Repository }

// NewService construye el servicio de acceso con la persistencia indicada.
func NewService(repository Repository) Service { return Service{repository: repository} }

// ReauthenticateWithPassword valida la contraseña y emite un ticket de corta duración.
func (s Service) ReauthenticateWithPassword(ctx context.Context, sessionToken, password string) (string, string, error) {
	stored, err := s.repository.CurrentPasswordHash(ctx, sessionToken)
	if err != nil || !registration.VerifyPassword(password, stored) {
		return "", "", ErrReauthenticationInvalid
	}
	ticket, hash, err := newTicket()
	if err != nil {
		return "", "", err
	}
	if err := s.repository.CreateReauthenticationTicket(ctx, sessionToken, hash); err != nil {
		return "", "", ErrReauthenticationInvalid
	}
	return ticket, time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339Nano), nil
}

// SetPassword consume un ticket de reautenticación y actualiza la contraseña.
func (s Service) SetPassword(ctx context.Context, sessionToken, ticket, password string) error {
	hash := sha256.Sum256([]byte("reauthentication-ticket:" + ticket))
	passwordHash, err := registration.HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.repository.ConsumeReauthenticationTicketAndSetPassword(ctx, sessionToken, hash[:], passwordHash); err != nil {
		return ErrReauthenticationInvalid
	}
	return nil
}

func newTicket() (string, []byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", nil, err
	}
	ticket := base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte("reauthentication-ticket:" + ticket))
	return ticket, hash[:], nil
}
