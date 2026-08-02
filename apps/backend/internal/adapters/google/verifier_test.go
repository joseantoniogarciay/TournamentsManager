package google

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

const testAudience = "test-client.apps.googleusercontent.com"

type certificateTransport struct{ body []byte }

func (t certificateTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Cache-Control": []string{"max-age=600"}},
		Body:       ioNopCloser{Reader: strings.NewReader(string(t.body))},
	}, nil
}

type ioNopCloser struct{ *strings.Reader }

func (ioNopCloser) Close() error { return nil }

func TestVerifyAcceptsLocallySignedGoogleIDToken(t *testing.T) {
	privateKey, verifier := localVerifier(t)
	token := signedToken(t, privateKey, map[string]any{
		"iss":            " https://accounts.google.com ",
		"aud":            testAudience,
		"sub":            " subject ",
		"email":          " person@example.test ",
		"nonce":          "nonce",
		"email_verified": true,
	})

	identity, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if identity.Issuer != "https://accounts.google.com" || identity.Subject != "subject" || identity.Email != "person@example.test" || identity.Nonce != "nonce" || !identity.EmailVerified {
		t.Fatalf("identity = %#v", identity)
	}
}

func TestVerifyRejectsInvalidSignatureAndAudience(t *testing.T) {
	privateKey, verifier := localVerifier(t)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		token string
	}{
		{"signature", signedToken(t, otherKey, map[string]any{"iss": "https://accounts.google.com", "aud": testAudience, "sub": "subject", "email": "person@example.test", "nonce": "nonce", "email_verified": true})},
		{"audience", signedToken(t, privateKey, map[string]any{"iss": "https://accounts.google.com", "aud": "another-client.apps.googleusercontent.com", "sub": "subject", "email": "person@example.test", "nonce": "nonce", "email_verified": true})},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := verifier.Verify(context.Background(), test.token); err == nil {
				t.Fatal("Verify() succeeded with an invalid token")
			}
		})
	}
}

func localVerifier(t *testing.T) (*rsa.PrivateKey, Verifier) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	modulus := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
	exponent := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())
	certificates, err := json.Marshal(map[string]any{"keys": []map[string]string{{"kid": "test-key", "kty": "RSA", "alg": "RS256", "use": "sig", "n": modulus, "e": exponent}}})
	if err != nil {
		t.Fatal(err)
	}
	validator, err := idtoken.NewValidator(context.Background(), option.WithHTTPClient(&http.Client{Transport: certificateTransport{body: certificates}}))
	if err != nil {
		t.Fatal(err)
	}
	return privateKey, Verifier{audiences: []string{testAudience}, validator: validator}
}

func signedToken(t *testing.T, privateKey *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(encoded))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signature)
}
