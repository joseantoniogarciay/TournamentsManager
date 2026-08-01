// Package federated contiene los casos de uso de identidad externa.
package federated

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// GoogleIssuer es el emisor canónico aceptado para los ID tokens de Google.
	GoogleIssuer      = "https://accounts.google.com"
	challengeLifetime = 5 * time.Minute
)

var (
	// ErrChallengeInvalid indica que el challenge o su nonce no son válidos.
	ErrChallengeInvalid = errors.New("challenge federado inválido")
	// ErrIdentityConflict indica que la identidad externa pertenece a otra cuenta.
	ErrIdentityConflict = errors.New("identidad federada en otra cuenta")
	// ErrEmailConflict evita vincular por una coincidencia de correo entre cuentas.
	ErrEmailConflict = errors.New("email ya pertenece a una cuenta")
	// ErrRegistration indica que una cuenta nueva requiere sus datos de alta.
	ErrRegistration = errors.New("alta Google requiere username")
)

// Identity es la identidad externa verificada que llega desde un proveedor OIDC.
type Identity struct {
	Issuer, Subject, Email, Nonce string
	EmailVerified                 bool
}

// Verifier valida el artefacto OIDC fuera del dominio.
type Verifier interface {
	Verify(context.Context, string) (Identity, error)
}

// Challenge es la prueba de un solo uso que se entrega al cliente antes de OAuth.
type Challenge struct{ ID, Nonce, ExpiresAt string }

// Registration contiene los campos propios necesarios para crear una cuenta social.
type Registration struct{ Username, Locale string }

// Session describe la sesión persistida sin exponer sus tokens sensibles.
type Session struct{ AccountID, Username, IdleExpiresAt, RefreshExpiresAt string }

// EstablishedSession une una sesión con los tokens entregados una sola vez.
type EstablishedSession struct {
	Session
	AccessToken, RefreshToken string
}

// Repository preserva las invariantes atómicas entre challenge, identidad y sesión.
type Repository interface {
	CreateChallenge(context.Context, []byte, time.Time) (string, error)
	AuthenticateGoogle(context.Context, string, []byte, Identity, *Registration, []byte, []byte) (Session, error)
	AddGoogleIdentity(context.Context, string, string, []byte, Identity) error
}

// Service coordina el caso de uso de inicio de sesión federado.
type Service struct {
	repository Repository
	verifier   Verifier
	now        func() time.Time
}

// NewService construye el caso de uso con sus puertos de persistencia y verificación.
func NewService(repository Repository, verifier Verifier) Service {
	return Service{repository: repository, verifier: verifier, now: time.Now}
}

// CreateChallenge emite el nonce opaco de un solo uso previo a Google.
func (s Service) CreateChallenge(ctx context.Context) (Challenge, error) {
	nonce, err := secret()
	if err != nil {
		return Challenge{}, err
	}
	hash := sha256.Sum256([]byte("google-login-nonce:" + nonce))
	expiresAt := s.now().Add(challengeLifetime)
	id, err := s.repository.CreateChallenge(ctx, hash[:], expiresAt)
	if err != nil {
		return Challenge{}, err
	}
	return Challenge{ID: id, Nonce: nonce, ExpiresAt: expiresAt.UTC().Format(time.RFC3339Nano)}, nil
}

// Authenticate valida una prueba Google y obtiene una sesión o solicita el alta.
func (s Service) Authenticate(ctx context.Context, challengeID, idToken string, registration *Registration) (EstablishedSession, error) {
	identity, err := s.verify(ctx, idToken)
	if err != nil {
		return EstablishedSession{}, err
	}
	if registration != nil && (registration.Username == "" || registration.Locale == "") {
		return EstablishedSession{}, ErrRegistration
	}
	challengeHash := sha256.Sum256([]byte("google-login-nonce:" + identity.Nonce))
	access, refresh, accessHash, refreshHash, err := sessionTokens()
	if err != nil {
		return EstablishedSession{}, err
	}
	session, err := s.repository.AuthenticateGoogle(ctx, challengeID, challengeHash[:], identity, registration, accessHash, refreshHash)
	if err != nil {
		return EstablishedSession{}, err
	}
	return EstablishedSession{Session: session, AccessToken: access, RefreshToken: refresh}, nil
}

// AddGoogle añade únicamente una identidad todavía libre a la cuenta ya autenticada.
func (s Service) AddGoogle(ctx context.Context, accountID, challengeID, idToken string) error {
	identity, err := s.verify(ctx, idToken)
	if err != nil {
		return err
	}
	challengeHash := sha256.Sum256([]byte("google-login-nonce:" + identity.Nonce))
	return s.repository.AddGoogleIdentity(ctx, accountID, challengeID, challengeHash[:], identity)
}

func (s Service) verify(ctx context.Context, idToken string) (Identity, error) {
	identity, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return Identity{}, fmt.Errorf("verificar identidad Google: %w", err)
	}
	if identity.Issuer != GoogleIssuer || identity.Subject == "" || !identity.EmailVerified || strings.TrimSpace(identity.Email) == "" || identity.Nonce == "" {
		return Identity{}, ErrChallengeInvalid
	}
	return identity, nil
}

func secret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sessionTokens() (string, string, []byte, []byte, error) {
	access, err := secret()
	if err != nil {
		return "", "", nil, nil, err
	}
	refresh, err := secret()
	if err != nil {
		return "", "", nil, nil, err
	}
	a := sha256.Sum256([]byte("session:" + access))
	r := sha256.Sum256([]byte("refresh:" + refresh))
	return access, refresh, a[:], r[:], nil
}
