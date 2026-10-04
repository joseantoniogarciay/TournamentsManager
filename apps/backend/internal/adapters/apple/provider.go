// Package apple implements Apple's server-side authorization code and OIDC boundary.
package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

const keysURL = "https://appleid.apple.com/auth/keys"
const tokenURL = "https://appleid.apple.com/auth/token" // #nosec G101 -- public protocol endpoint, not a credential.

// Provider never exposes its private signing key or Apple's tokens to callers.
type Provider struct {
	serviceID, teamID, keyID, redirectURI string
	privateKey                            *ecdsa.PrivateKey
	client                                *http.Client
	now                                   func() time.Time
	keysURL, tokenURL                     string
	mu                                    sync.Mutex
	keys                                  map[string]*rsa.PublicKey
	keysUnavailable                       bool
	keysExpire, keysFetched               time.Time
}

// NewProvider validates a configured ES256 signing key at composition time.
func NewProvider(serviceID, teamID, keyID, redirectURI string, privatePEM []byte) (*Provider, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM(privatePEM)
	if err != nil || key.Curve != elliptic.P256() {
		return nil, errors.New("apple signing key must be a PKCS8 P-256 key")
	}
	return &Provider{serviceID: serviceID, teamID: teamID, keyID: keyID, redirectURI: redirectURI, privateKey: key, client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, now: time.Now, keysURL: keysURL, tokenURL: tokenURL}, nil
}

// AuthorizationURL includes independent unpredictable state and nonce.
func (p *Provider) AuthorizationURL(state, nonce string) string {
	query := url.Values{"client_id": {p.serviceID}, "redirect_uri": {p.redirectURI}, "response_type": {"code"}, "response_mode": {"form_post"}, "scope": {"email"}, "state": {state}, "nonce": {nonce}}
	return "https://appleid.apple.com/auth/authorize?" + query.Encode()
}

func (p *Provider) clientSecret() (string, error) {
	now := p.now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{Issuer: p.teamID, Subject: p.serviceID, Audience: jwt.ClaimStrings{federated.AppleIssuer}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute))})
	token.Header["kid"] = p.keyID
	return token.SignedString(p.privateKey)
}

// Exchange obtains and verifies an ID token; access/refresh tokens are discarded.
func (p *Provider) Exchange(ctx context.Context, code string) (federated.Identity, error) {
	secret, err := p.clientSecret()
	if err != nil {
		return federated.Identity{}, federated.ErrAppleUnavailable
	}
	form := url.Values{"client_id": {p.serviceID}, "client_secret": {secret}, "code": {code}, "grant_type": {"authorization_code"}, "redirect_uri": {p.redirectURI}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return federated.Identity{}, federated.ErrAppleUnavailable
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := p.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return federated.Identity{}, ctx.Err()
		}
		return federated.Identity{}, federated.ErrAppleUnavailable
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusBadRequest {
		return federated.Identity{}, federated.ErrChallengeInvalid
	}
	if response.StatusCode != http.StatusOK {
		return federated.Identity{}, federated.ErrAppleUnavailable
	}
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 128*1024)).Decode(&body); err != nil || body.IDToken == "" {
		return federated.Identity{}, federated.ErrAppleUnavailable
	}
	return p.verify(ctx, body.IDToken)
}

type appleClaims struct {
	jwt.RegisteredClaims
	Nonce         string          `json:"nonce"`
	Email         string          `json:"email"`
	EmailVerified json.RawMessage `json:"email_verified"`
}

func (p *Provider) verify(ctx context.Context, raw string) (federated.Identity, error) {
	claims := new(appleClaims)
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, federated.ErrChallengeInvalid
		}
		return p.key(ctx, kid)
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(federated.AppleIssuer), jwt.WithAudience(p.serviceID), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithTimeFunc(p.now))
	if err != nil {
		if errors.Is(err, federated.ErrAppleUnavailable) {
			return federated.Identity{}, federated.ErrAppleUnavailable
		}
		if ctx.Err() != nil {
			return federated.Identity{}, ctx.Err()
		}
		return federated.Identity{}, federated.ErrChallengeInvalid
	}
	if claims.Subject == "" || claims.Nonce == "" {
		return federated.Identity{}, federated.ErrChallengeInvalid
	}
	verified := string(claims.EmailVerified) == "true" || string(claims.EmailVerified) == `"true"`
	return federated.Identity{Issuer: claims.Issuer, Subject: claims.Subject, Email: claims.Email, EmailVerified: verified, Nonce: claims.Nonce}, nil
}

func (p *Provider) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.now()
	if key := p.keys[kid]; key != nil && now.Before(p.keysExpire) {
		return key, nil
	}
	// Unknown kids cannot force an unbounded burst of remote JWKS fetches.
	if now.Before(p.keysFetched.Add(time.Minute)) {
		if p.keysUnavailable {
			return nil, federated.ErrAppleUnavailable
		}
		return nil, federated.ErrChallengeInvalid
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.keysURL, nil)
	if err != nil {
		return nil, federated.ErrAppleUnavailable
	}
	p.keysFetched = now
	p.keysUnavailable = true
	response, err := p.client.Do(request)
	if err != nil {
		return nil, federated.ErrAppleUnavailable
	}
	defer func() { _ = response.Body.Close() }()
	p.keysFetched = now
	if response.StatusCode != http.StatusOK {
		return nil, federated.ErrAppleUnavailable
	}
	var document struct {
		Keys []struct {
			KeyID     string `json:"kid"`
			KeyType   string `json:"kty"`
			Algorithm string `json:"alg"`
			Use       string `json:"use"`
			Modulus   string `json:"n"`
			Exponent  string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 128*1024)).Decode(&document); err != nil {
		return nil, federated.ErrAppleUnavailable
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, item := range document.Keys {
		if item.KeyType != "RSA" || item.Algorithm != "RS256" || item.Use != "sig" || item.KeyID == "" {
			continue
		}
		n, nErr := base64.RawURLEncoding.DecodeString(item.Modulus)
		e, eErr := base64.RawURLEncoding.DecodeString(item.Exponent)
		if nErr != nil || eErr != nil || len(n) < 256 || len(e) == 0 || len(e) > 4 {
			continue
		}
		exponent := new(big.Int).SetBytes(e).Int64()
		if exponent < 3 || exponent > 2147483647 || exponent%2 == 0 {
			continue
		}
		keys[item.KeyID] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent)}
	}
	if len(keys) == 0 {
		return nil, federated.ErrAppleUnavailable
	}
	p.keys, p.keysExpire = keys, now.Add(time.Hour)
	p.keysUnavailable = false
	if key := keys[kid]; key != nil {
		return key, nil
	}
	return nil, federated.ErrChallengeInvalid
}
