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
	"net/http/httptest"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

func TestRISCVerifierAcceptsSignedGoogleSessionRevocation(t *testing.T) {
	privateKey, verifier := localRISCVerifier(t)
	token := signedRISC(t, privateKey, map[string]any{
		"iss": "https://accounts.google.com/",
		"aud": testAudience,
		"jti": "risc-event-1",
		"events": map[string]any{federated.RISCSessionsRevoked: map[string]any{
			"subject": map[string]any{"subject_type": "iss-sub", "iss": "https://accounts.google.com/", "sub": "google-subject"},
		}},
	})
	event, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if event.ID != "risc-event-1" || event.Issuer != federated.GoogleIssuer || event.Subject != "google-subject" || event.Type != federated.RISCSessionsRevoked {
		t.Fatalf("event = %#v", event)
	}
}

func TestRISCVerifierAcceptsVerificationWithoutSubject(t *testing.T) {
	privateKey, verifier := localRISCVerifier(t)
	token := signedRISC(t, privateKey, map[string]any{
		"iss": "https://accounts.google.com/",
		"aud": testAudience,
		"jti": "risc-verification-1",
		"events": map[string]any{
			federated.RISCVerification: map[string]any{"state": "verification-state"},
		},
	})

	event, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if event.ID != "risc-verification-1" || event.Issuer != federated.GoogleIssuer || event.Type != federated.RISCVerification || event.Subject != "" {
		t.Fatalf("Verify() event = %#v", event)
	}
}

func TestRISCVerifierAcceptsGoogleDefaultJWTHeaderType(t *testing.T) {
	privateKey, verifier := localRISCVerifier(t)
	token := signedRISCWithHeaderType(t, privateKey, "JWT", map[string]any{
		"iss": "https://accounts.google.com/",
		"aud": testAudience,
		"jti": "risc-default-jwt-header",
		"events": map[string]any{
			federated.RISCVerification: map[string]any{"state": "verification-state"},
		},
	})

	if _, err := verifier.Verify(context.Background(), token); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestRISCVerifierAcceptsAudienceArray(t *testing.T) {
	privateKey, verifier := localRISCVerifier(t)
	token := signedRISC(t, privateKey, map[string]any{
		"iss": "https://accounts.google.com/",
		"aud": []string{"other-client.apps.googleusercontent.com", testAudience},
		"jti": "risc-audience-array",
		"events": map[string]any{
			federated.RISCVerification: map[string]any{"state": "verification-state"},
		},
	})

	if _, err := verifier.Verify(context.Background(), token); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestRISCVerifierRejectsUnexpectedAudienceAndSubjectIssuer(t *testing.T) {
	privateKey, verifier := localRISCVerifier(t)
	for _, test := range []struct {
		name, audience, subjectIssuer string
	}{
		{"audience", "other-client.apps.googleusercontent.com", "https://accounts.google.com/"},
		{"subject issuer", testAudience, "https://issuer.example/"},
	} {
		t.Run(test.name, func(t *testing.T) {
			token := signedRISC(t, privateKey, map[string]any{
				"iss": "https://accounts.google.com/", "aud": test.audience, "jti": "risc-event-2",
				"events": map[string]any{federated.RISCTokensRevoked: map[string]any{
					"subject": map[string]any{"subject_type": "iss-sub", "iss": test.subjectIssuer, "sub": "google-subject"},
				}},
			})
			if _, err := verifier.Verify(context.Background(), token); err == nil {
				t.Fatal("Verify() succeeded")
			}
		})
	}
}

func localRISCVerifier(t *testing.T) (*rsa.PrivateKey, *RISCVerifier) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{"issuer": "https://accounts.google.com/", "jwks_uri": server.URL + "/jwks"})
		case "/jwks":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kid": "risc-test-key", "kty": "RSA", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())}}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	verifier := NewRISCVerifier([]string{testAudience})
	verifier.configURL = server.URL + "/configuration"
	return privateKey, verifier
}

func signedRISC(t *testing.T, privateKey *rsa.PrivateKey, claims map[string]any) string {
	return signedRISCWithHeaderType(t, privateKey, "secevent+jwt", claims)
}

func signedRISCWithHeaderType(t *testing.T, privateKey *rsa.PrivateKey, headerType string, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "kid": "risc-test-key", "typ": headerType})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}
