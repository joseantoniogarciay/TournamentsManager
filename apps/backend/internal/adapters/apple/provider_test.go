package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

func testProvider(t *testing.T) (*Provider, *rsa.PrivateKey) {
	t.Helper()
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(ec)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewProvider("com.example.web", "TEAM123456", "KEY1234567", "https://api.example.test/v1/apple-callback", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	if err != nil {
		t.Fatal(err)
	}
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	provider.keys = map[string]*rsa.PublicKey{"apple-key": &rsaKey.PublicKey}
	provider.keysExpire = time.Now().Add(time.Hour)
	return provider, rsaKey
}

func signedIDToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "apple-key"
	raw, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{"iss": federated.AppleIssuer, "aud": "com.example.web", "sub": "verified-apple-subject", "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Add(-time.Minute).Unix(), "nonce": "nonce", "email": "relay@privaterelay.appleid.com", "email_verified": "true"}
}

func TestAppleVerifierRejectsUntrustedClaimsAndSignatures(t *testing.T) {
	provider, key := testProvider(t)
	cases := []struct {
		name, field string
		value       any
	}{
		{"wrong issuer", "iss", "https://evil.test"}, {"wrong audience", "aud", "another-service"}, {"expired", "exp", time.Now().Add(-time.Second).Unix()}, {"missing expiry", "exp", nil}, {"future issuance", "iat", time.Now().Add(time.Hour).Unix()}, {"missing subject", "sub", ""}, {"missing nonce", "nonce", ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			claims := validClaims()
			if test.value == nil {
				delete(claims, test.field)
			} else {
				claims[test.field] = test.value
			}
			if _, err := provider.verify(context.Background(), signedIDToken(t, key, claims)); !errors.Is(err, federated.ErrChallengeInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	raw := signedIDToken(t, key, validClaims())
	parts := strings.Split(raw, ".")
	parts[2] = strings.Repeat("A", len(parts[2]))
	if _, err := provider.verify(context.Background(), strings.Join(parts, ".")); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("tamper error = %v", err)
	}
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims()).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.verify(context.Background(), none); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatalf("unsigned token error=%v", err)
	}
}

func TestAppleVerifierAcceptsBooleanAndStringEmailAndAbsentEmailForExistingIdentity(t *testing.T) {
	provider, key := testProvider(t)
	for _, verified := range []any{true, "true", false, "false"} {
		claims := validClaims()
		claims["email_verified"] = verified
		identity, err := provider.verify(context.Background(), signedIDToken(t, key, claims))
		if err != nil {
			t.Fatal(err)
		}
		if identity.EmailVerified != (verified == true || verified == "true") {
			t.Fatal("verification claim misinterpreted")
		}
	}
	claims := validClaims()
	delete(claims, "email")
	delete(claims, "email_verified")
	identity, err := provider.verify(context.Background(), signedIDToken(t, key, claims))
	if err != nil || identity.Subject == "" || identity.EmailVerified {
		t.Fatalf("identity=%#v, error=%v", identity, err)
	}
}

func TestAppleExchangeVerifiesCodeAndSignsServerSecret(t *testing.T) {
	provider, key := testProvider(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.ParseForm() != nil {
			t.Error("not a token POST")
		}
		if r.Form.Get("code") != "one-use-code" || r.Form.Get("redirect_uri") != provider.redirectURI || r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("client_id") != provider.serviceID {
			t.Error("wrong code exchange")
		}
		claims := new(jwt.RegisteredClaims)
		_, err := jwt.ParseWithClaims(r.Form.Get("client_secret"), claims, func(*jwt.Token) (any, error) { return &provider.privateKey.PublicKey, nil }, jwt.WithValidMethods([]string{"ES256"}), jwt.WithIssuer(provider.teamID), jwt.WithAudience(federated.AppleIssuer))
		if err != nil || claims.Subject != provider.serviceID {
			t.Errorf("invalid client secret: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": signedIDToken(t, key, validClaims()), "access_token": "discard-this-token", "refresh_token": "discard-this-refresh"})
	}))
	defer server.Close()
	provider.tokenURL = server.URL
	identity, err := provider.Exchange(context.Background(), "one-use-code")
	if err != nil || identity.Issuer != federated.AppleIssuer {
		t.Fatalf("identity=%#v, err=%v", identity, err)
	}
}

func TestAppleJWKSCacheAndUnknownKeyThrottle(t *testing.T) {
	provider, key := testProvider(t)
	provider.keys = nil
	provider.keysExpire = time.Time{}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kid": "apple-key", "kty": "RSA", "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
	}))
	defer server.Close()
	provider.keysURL = server.URL
	if _, err := provider.verify(context.Background(), signedIDToken(t, key, validClaims())); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.key(context.Background(), "unknown-key"); !errors.Is(err, federated.ErrChallengeInvalid) {
		t.Fatal(err)
	}
	if _, err := provider.key(context.Background(), "apple-key"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("JWKS fetches=%d", calls)
	}
}

func TestAppleExchangeBoundariesAreSafe(t *testing.T) {
	for _, status := range []int{400, 429, 500, 200} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			provider, _ := testProvider(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"internal":"private-token-and-email"}`))
			}))
			defer server.Close()
			provider.tokenURL = server.URL
			_, err := provider.Exchange(context.Background(), "code")
			expected := federated.ErrAppleUnavailable
			if status == 400 {
				expected = federated.ErrChallengeInvalid
			}
			if !errors.Is(err, expected) || strings.Contains(err.Error(), "private") {
				t.Fatalf("unsafe error: %v", err)
			}
		})
	}
	provider, _ := testProvider(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Exchange(ctx, "code"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
}
