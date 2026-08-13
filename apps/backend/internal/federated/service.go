// Package federated contains external identity use cases.
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
	// GoogleIssuer is the accepted canonical issuer for Google ID tokens.
	GoogleIssuer      = "https://accounts.google.com"
	challengeLifetime = 5 * time.Minute
)

var (
	// ErrChallengeInvalid indicates that the challenge or its nonce is invalid.
	ErrChallengeInvalid = errors.New("challenge federado inválido")
	// ErrIdentityConflict indicates that the external identity belongs to another account.
	ErrIdentityConflict = errors.New("identidad federada en otra cuenta")
	// ErrEmailConflict prevents linking through a matching email across accounts.
	ErrEmailConflict = errors.New("email ya pertenece a una cuenta")
	// ErrRegistration indicates that a new account requires registration data.
	ErrRegistration = errors.New("alta Google requiere username")
)

// Identity is the verified external identity received from an OIDC provider.
type Identity struct {
	Issuer, Subject, Email, Nonce string
	EmailVerified                 bool
}

// Verifier validates the OIDC artifact outside the domain.
type Verifier interface {
	Verify(context.Context, string) (Identity, error)
}

// Challenge is the single-use proof delivered to the client before OAuth.
type Challenge struct{ ID, Nonce, ExpiresAt string }

// Registration contains the fields required to create a social account.
type Registration struct {
	Username, Locale string
	Draft            *Draft
}

// Draft is the complete tournament that can be created with a new account in the same transaction.
type Draft struct {
	Name  string
	Teams []string
}

// Session describes the persisted session without exposing its sensitive tokens.
type Session struct{ AccountID, Username, IdleExpiresAt, RefreshExpiresAt string }

// EstablishedSession joins a session with its one-time-delivered tokens.
type EstablishedSession struct {
	Session
	AccessToken, RefreshToken string
}

// Repository preserves atomic invariants across challenge, identity, and session.
type Repository interface {
	CreateChallenge(context.Context, []byte, time.Time) (string, error)
	AuthenticateGoogle(context.Context, string, []byte, Identity, *Registration, []byte, []byte) (Session, error)
	AddGoogleIdentity(context.Context, string, string, []byte, Identity) error
	ReauthenticateGoogle(context.Context, string, string, string, []byte, Identity, []byte) error
	AddGoogleIdentityWithTicket(context.Context, string, string, []byte, Identity, []byte) error
}

// ReauthenticateGoogle proves an already linked identity and issues a single-use ticket.
func (s Service) ReauthenticateGoogle(ctx context.Context, accountID, sessionToken, challengeID, idToken string) (string, string, error) {
	identity, err := s.verify(ctx, idToken)
	if err != nil {
		return "", "", err
	}
	ticket, err := secret()
	if err != nil {
		return "", "", err
	}
	ticketDigest := sha256.Sum256([]byte("reauthentication-ticket:" + ticket))
	challengeHash := sha256.Sum256([]byte("google-login-nonce:" + identity.Nonce))
	if err := s.repository.ReauthenticateGoogle(ctx, accountID, sessionToken, challengeID, challengeHash[:], identity, ticketDigest[:]); err != nil {
		return "", "", err
	}
	return ticket, s.now().Add(challengeLifetime).UTC().Format(time.RFC3339Nano), nil
}

// AddGoogleWithTicket links Google and consumes the ticket in the same transaction.
func (s Service) AddGoogleWithTicket(ctx context.Context, sessionToken, ticket, challengeID, idToken string) error {
	identity, err := s.verify(ctx, idToken)
	if err != nil {
		return err
	}
	ticketHash := sha256.Sum256([]byte("reauthentication-ticket:" + ticket))
	challengeHash := sha256.Sum256([]byte("google-login-nonce:" + identity.Nonce))
	return s.repository.AddGoogleIdentityWithTicket(ctx, sessionToken, challengeID, challengeHash[:], identity, ticketHash[:])
}

// Service coordinates the federated sign-in use case.
type Service struct {
	repository Repository
	verifier   Verifier
	now        func() time.Time
}

// NewService builds the use case with its persistence and verification ports.
func NewService(repository Repository, verifier Verifier) Service {
	return Service{repository: repository, verifier: verifier, now: time.Now}
}

// CreateChallenge issues the opaque single-use nonce required before Google.
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

// Authenticate validates a Google proof and obtains a session or requests registration.
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

// AddGoogle adds only an identity still unlinked to the already authenticated account.
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
