package federated

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
)

// AppleIssuer is the sole accepted Apple identity issuer.
const AppleIssuer = "https://appleid.apple.com"

// ErrAppleUnavailable marks a provider network/key boundary, never its raw error.
var ErrAppleUnavailable = errors.New("apple identity provider unavailable")

// AppleProvider owns authorization URLs, code exchange and OIDC verification.
type AppleProvider interface {
	AuthorizationURL(state, nonce string) string
	Exchange(context.Context, string) (Identity, error)
}

// AppleCallback is a claimed, single-use callback; it never contains provider tokens.
type AppleCallback struct {
	ID, Platform string
	NonceHash    []byte
}

// AppleRepository stores short-lived proofs and atomically establishes sessions.
type AppleRepository interface {
	CreateAppleChallenge(context.Context, []byte, []byte, []byte, string, time.Time) (string, error)
	ClaimAppleCallback(context.Context, []byte) (AppleCallback, error)
	CompleteAppleCallback(context.Context, string, Identity) error
	AuthenticateApple(context.Context, string, []byte, *Registration, *Draft, []byte, []byte) (Session, error)
}

// AppleService coordinates Apple proofs without importing HTTP or persistence.
type AppleService struct {
	repository AppleRepository
	provider   AppleProvider
	now        func() time.Time
}

// NewAppleService composes the Apple ports.
func NewAppleService(repository AppleRepository, provider AppleProvider) AppleService {
	return AppleService{repository: repository, provider: provider, now: time.Now}
}

// AppleChallenge is a short-lived authorization URL with a client-bound proof.
type AppleChallenge struct{ ID, AuthorizationURL, ExpiresAt string }

// CreateChallenge binds an Apple request to a platform and client proof digest.
func (s AppleService) CreateChallenge(ctx context.Context, platform, proofChallenge string) (AppleChallenge, error) {
	proofHash, err := hex.DecodeString(proofChallenge)
	if err != nil || len(proofHash) != 32 || hex.EncodeToString(proofHash) != proofChallenge || (platform != "web" && platform != "ios" && platform != "android") {
		return AppleChallenge{}, ErrChallengeInvalid
	}
	state, err := secret()
	if err != nil {
		return AppleChallenge{}, err
	}
	nonce, err := secret()
	if err != nil {
		return AppleChallenge{}, err
	}
	stateHash := sha256.Sum256([]byte("apple-login-state:" + state))
	nonceHash := sha256.Sum256([]byte("apple-login-nonce:" + nonce))
	expires := s.now().Add(challengeLifetime)
	id, err := s.repository.CreateAppleChallenge(ctx, stateHash[:], nonceHash[:], proofHash, platform, expires)
	if err != nil {
		return AppleChallenge{}, err
	}
	return AppleChallenge{ID: id, AuthorizationURL: s.provider.AuthorizationURL(state, nonce), ExpiresAt: expires.UTC().Format(time.RFC3339Nano)}, nil
}

// Complete claims state before contacting Apple, validates nonce and stores only
// the verified identity. A denial consumes the callback without opening a session.
func (s AppleService) Complete(ctx context.Context, state, code string, denied bool) (AppleCallback, error) {
	if state == "" {
		return AppleCallback{}, ErrChallengeInvalid
	}
	hash := sha256.Sum256([]byte("apple-login-state:" + state))
	callback, err := s.repository.ClaimAppleCallback(ctx, hash[:])
	if err != nil {
		return AppleCallback{}, err
	}
	if denied {
		return callback, nil
	}
	if code == "" {
		return callback, ErrChallengeInvalid
	}
	identity, err := s.provider.Exchange(ctx, code)
	if err != nil {
		return callback, err
	}
	nonceHash := sha256.Sum256([]byte("apple-login-nonce:" + identity.Nonce))
	if identity.Issuer != AppleIssuer || identity.Subject == "" || identity.Nonce == "" || subtle.ConstantTimeCompare(callback.NonceHash, nonceHash[:]) != 1 {
		return callback, ErrChallengeInvalid
	}
	if err := s.repository.CompleteAppleCallback(ctx, callback.ID, identity); err != nil {
		return callback, err
	}
	return callback, nil
}

// Authenticate requires the initiating client's secret, never the callback URL.
func (s AppleService) Authenticate(ctx context.Context, id, proof string, registration *Registration, draft *Draft) (EstablishedSession, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(proof); err != nil || len(decoded) != 32 || len(proof) != 43 {
		return EstablishedSession{}, ErrChallengeInvalid
	}
	if registration != nil && (registration.Username == "" || registration.Locale == "" || registration.TermsVersion != legal.CurrentTermsVersion) {
		return EstablishedSession{}, ErrRegistration
	}
	hash := sha256.Sum256([]byte("apple-login-proof:" + proof))
	access, refresh, accessHash, refreshHash, err := sessionTokens()
	if err != nil {
		return EstablishedSession{}, err
	}
	session, err := s.repository.AuthenticateApple(ctx, id, hash[:], registration, draft, accessHash, refreshHash)
	if err != nil {
		return EstablishedSession{}, err
	}
	return EstablishedSession{Session: session, AccessToken: access, RefreshToken: refresh}, nil
}
