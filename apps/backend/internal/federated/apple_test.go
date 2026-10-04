package federated

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

type appleRepositoryStub struct {
	nonceHash, proofHash                      []byte
	platform                                  string
	callbackErr, completeErr, authenticateErr error
	completed                                 bool
}

func (r *appleRepositoryStub) CreateAppleChallenge(_ context.Context, _ []byte, nonce, proof []byte, platform string, _ time.Time) (string, error) {
	r.nonceHash = nonce
	r.proofHash = proof
	r.platform = platform
	return "challenge", nil
}
func (r *appleRepositoryStub) ClaimAppleCallback(context.Context, []byte) (AppleCallback, error) {
	return AppleCallback{ID: "challenge", NonceHash: r.nonceHash, Platform: r.platform}, r.callbackErr
}
func (r *appleRepositoryStub) CompleteAppleCallback(context.Context, string, Identity) error {
	r.completed = true
	return r.completeErr
}
func (r *appleRepositoryStub) AuthenticateApple(_ context.Context, _ string, proof []byte, _ *Registration, _ *Draft, _, _ []byte) (Session, error) {
	if !equalAppleTestHash(proof, r.proofHash) {
		return Session{}, ErrChallengeInvalid
	}
	return Session{AccountID: "account"}, r.authenticateErr
}
func equalAppleTestHash(a, b []byte) bool { return hex.EncodeToString(a) == hex.EncodeToString(b) }

type appleProviderStub struct {
	identity Identity
	err      error
	called   bool
	nonce    string
}

func (p *appleProviderStub) AuthorizationURL(_, nonce string) string {
	p.nonce = nonce
	return "https://appleid.apple.com/auth/authorize"
}
func (p *appleProviderStub) Exchange(context.Context, string) (Identity, error) {
	p.called = true
	return p.identity, p.err
}

func TestAppleChallengeBindsNonceAndClientProof(t *testing.T) {
	repo := new(appleRepositoryStub)
	provider := new(appleProviderStub)
	service := NewAppleService(repo, provider)
	proof, err := secret()
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte("apple-login-proof:" + proof))
	challenge, err := service.CreateChallenge(context.Background(), "android", hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}
	provider.identity = Identity{Issuer: AppleIssuer, Subject: "apple-subject", Nonce: provider.nonce}
	if _, err := service.Complete(context.Background(), "state", "code", false); err != nil || !repo.completed {
		t.Fatalf("callback=%v", err)
	}
	if _, err := service.Authenticate(context.Background(), challenge.ID, proof, nil, nil); err != nil {
		t.Fatal(err)
	}
	other, err := secret()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), challenge.ID, other, nil, nil); !errors.Is(err, ErrChallengeInvalid) {
		t.Fatalf("wrong app proof=%v", err)
	}
}

func TestAppleCallbackNeverTrustsWrongNonceIssuerOrSubject(t *testing.T) {
	for _, test := range []struct {
		name     string
		identity Identity
	}{{"wrong nonce", Identity{Issuer: AppleIssuer, Subject: "subject", Nonce: "other"}}, {"wrong issuer", Identity{Issuer: GoogleIssuer, Subject: "subject", Nonce: "nonce"}}, {"missing subject", Identity{Issuer: AppleIssuer, Nonce: "nonce"}}} {
		t.Run(test.name, func(t *testing.T) {
			hash := sha256.Sum256([]byte("apple-login-nonce:nonce"))
			repo := &appleRepositoryStub{nonceHash: hash[:]}
			provider := &appleProviderStub{identity: test.identity}
			if _, err := NewAppleService(repo, provider).Complete(context.Background(), "state", "code", false); !errors.Is(err, ErrChallengeInvalid) || repo.completed {
				t.Fatalf("completed=%v, error=%v", repo.completed, err)
			}
		})
	}
}

func TestAppleCancellationDoesNotContactProviderOrStoreIdentity(t *testing.T) {
	repo := new(appleRepositoryStub)
	provider := new(appleProviderStub)
	callback, err := NewAppleService(repo, provider).Complete(context.Background(), "state", "", true)
	if err != nil || callback.ID == "" || provider.called || repo.completed {
		t.Fatalf("cancel=%#v, %v", callback, err)
	}
}

func TestAppleInvalidChallengeAndPlatformDoNotReachProvider(t *testing.T) {
	repo := &appleRepositoryStub{callbackErr: ErrChallengeInvalid}
	provider := new(appleProviderStub)
	service := NewAppleService(repo, provider)
	if _, err := service.Complete(context.Background(), "state", "code", false); !errors.Is(err, ErrChallengeInvalid) || provider.called {
		t.Fatal("invalid state contacted Apple")
	}
	hash := sha256.Sum256([]byte("proof"))
	if _, err := service.CreateChallenge(context.Background(), "https://evil.test", hex.EncodeToString(hash[:])); !errors.Is(err, ErrChallengeInvalid) {
		t.Fatal("accepted arbitrary redirect")
	}
	if _, err := service.CreateChallenge(context.Background(), "ios", "invalid-digest"); !errors.Is(err, ErrChallengeInvalid) {
		t.Fatal("accepted invalid proof digest")
	}
}
