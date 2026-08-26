package google

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

const riscConfigurationURL = "https://accounts.google.com/.well-known/risc-configuration"

// RISCVerifier validates Google Cross-Account Protection security event tokens.
// It fetches Google's current RISC configuration and signing keys, caching both
// briefly to avoid turning each event into a remote dependency burst.
type RISCVerifier struct {
	audiences map[string]struct{}
	client    *http.Client
	configURL string
	now       func() time.Time
	mu        sync.Mutex
	config    riscConfiguration
	configTTL time.Time
	keys      map[string]rsa.PublicKey
	keysURI   string
	keysTTL   time.Time
}

type riscConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

type riscJWKS struct {
	Keys []riscJWK `json:"keys"`
}

type riscJWK struct {
	Algorithm string `json:"alg"`
	Exponent  string `json:"e"`
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Modulus   string `json:"n"`
}

type riscHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

type riscClaims struct {
	Audience json.RawMessage            `json:"aud"`
	Events   map[string]json.RawMessage `json:"events"`
	ID       string                     `json:"jti"`
	Issuer   string                     `json:"iss"`
}

type riscEventBody struct {
	Reason  string      `json:"reason"`
	Subject riscSubject `json:"subject"`
}

type riscSubject struct {
	Issuer  string `json:"iss"`
	Subject string `json:"sub"`
	Type    string `json:"subject_type"`
}

// NewRISCVerifier limits events to the OAuth audiences accepted by the API.
func NewRISCVerifier(audiences []string) *RISCVerifier {
	allowed := make(map[string]struct{}, len(audiences))
	for _, audience := range audiences {
		allowed[audience] = struct{}{}
	}
	return &RISCVerifier{audiences: allowed, client: &http.Client{Timeout: 5 * time.Second}, configURL: riscConfigurationURL, now: time.Now}
}

// Verify parses and validates a RISC SET before exposing its small safe domain
// representation. A SET may contain several events; RISC delivery from Google
// sends one, so multiple requested events are rejected rather than guessed.
func (v *RISCVerifier) Verify(ctx context.Context, raw string) (federated.RISCEvent, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 || anyEmpty(parts) {
		return federated.RISCEvent{}, errors.New("SET RISC mal formado")
	}
	var header riscHeader
	var claims riscClaims
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return federated.RISCEvent{}, errors.New("cabecera SET inválida")
	}
	if err := decodeJWTPart(parts[1], &claims); err != nil {
		return federated.RISCEvent{}, errors.New("claims SET inválidos")
	}
	if header.Algorithm != "RS256" || header.KeyID == "" {
		return federated.RISCEvent{}, errors.New("cabecera SET no permitida")
	}
	if claims.ID == "" {
		return federated.RISCEvent{}, errors.New("SET RISC sin jti")
	}
	if claims.Issuer == "" {
		return federated.RISCEvent{}, errors.New("SET RISC sin emisor")
	}
	if len(claims.Audience) == 0 {
		return federated.RISCEvent{}, errors.New("SET RISC sin audiencia")
	}
	if len(claims.Events) == 0 {
		return federated.RISCEvent{}, errors.New("SET RISC sin evento")
	}
	if !v.acceptsAudience(claims.Audience) {
		return federated.RISCEvent{}, errors.New("audiencia SET no permitida")
	}
	configuration, key, err := v.keyFor(ctx, header.KeyID)
	if err != nil {
		return federated.RISCEvent{}, fmt.Errorf("%w: obtener clave RISC", federated.ErrRISCUnavailable)
	}
	if claims.Issuer != configuration.Issuer || strings.TrimSuffix(claims.Issuer, "/") != federated.GoogleIssuer {
		return federated.RISCEvent{}, errors.New("emisor SET no permitido")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return federated.RISCEvent{}, errors.New("firma SET inválida")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(&key, crypto.SHA256, digest[:], signature); err != nil {
		return federated.RISCEvent{}, errors.New("firma SET inválida")
	}
	if rawEvent, ok := claims.Events[federated.RISCVerification]; ok {
		if len(claims.Events) != 1 {
			return federated.RISCEvent{}, errors.New("SET RISC de verificación combinado")
		}
		var verification struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(rawEvent, &verification); err != nil || verification.State == "" {
			return federated.RISCEvent{}, errors.New("evento RISC de verificación inválido")
		}
		return federated.RISCEvent{ID: claims.ID, Issuer: federated.GoogleIssuer, Type: federated.RISCVerification}, nil
	}

	var event federated.RISCEvent
	for eventType, rawEvent := range claims.Events {
		if eventType != federated.RISCSessionsRevoked && eventType != federated.RISCTokensRevoked && eventType != federated.RISCAccountDisabled {
			continue
		}
		var body riscEventBody
		if err := json.Unmarshal(rawEvent, &body); err != nil || body.Subject.Type != "iss-sub" || body.Subject.Issuer != claims.Issuer || body.Subject.Subject == "" {
			return federated.RISCEvent{}, errors.New("evento RISC inválido")
		}
		if event.Subject != "" && event.Subject != body.Subject.Subject {
			return federated.RISCEvent{}, errors.New("SET RISC con sujetos distintos")
		}
		candidate := federated.RISCEvent{ID: claims.ID, Issuer: federated.GoogleIssuer, Subject: body.Subject.Subject, Type: eventType, Reason: body.Reason}
		if event.Subject == "" || riscEventPriority(candidate) > riscEventPriority(event) {
			event = candidate
		}
	}
	if event.Subject == "" {
		return federated.RISCEvent{}, errors.New("SET RISC sin evento admitido")
	}
	return event, nil
}

func riscEventPriority(event federated.RISCEvent) int {
	switch {
	case event.Type == federated.RISCSessionsRevoked:
		return 3
	case event.Type == federated.RISCTokensRevoked:
		return 2
	case event.Type == federated.RISCAccountDisabled && event.Reason == "hijacking":
		return 1
	default:
		return 0
	}
}

func (v *RISCVerifier) acceptsAudience(raw json.RawMessage) bool {
	var audience string
	if err := json.Unmarshal(raw, &audience); err == nil {
		_, ok := v.audiences[audience]
		return ok
	}
	var audiences []string
	if err := json.Unmarshal(raw, &audiences); err != nil {
		return false
	}
	for _, audience := range audiences {
		if _, ok := v.audiences[audience]; ok {
			return true
		}
	}
	return false
}

func (v *RISCVerifier) keyFor(ctx context.Context, keyID string) (riscConfiguration, rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := v.now()
	if now.After(v.configTTL) {
		var configuration riscConfiguration
		if err := v.fetchJSON(ctx, v.configURL, &configuration); err != nil || configuration.Issuer == "" || configuration.JWKSURI == "" {
			return riscConfiguration{}, rsa.PublicKey{}, errors.New("configuración RISC no disponible")
		}
		v.config = configuration
		v.configTTL = now.Add(time.Hour)
	}
	if now.After(v.keysTTL) || v.keysURI != v.config.JWKSURI || v.keys[keyID].N == nil {
		var document riscJWKS
		if err := v.fetchJSON(ctx, v.config.JWKSURI, &document); err != nil {
			return riscConfiguration{}, rsa.PublicKey{}, errors.New("JWKS RISC no disponible")
		}
		keys := make(map[string]rsa.PublicKey, len(document.Keys))
		for _, jwk := range document.Keys {
			if jwk.Algorithm != "RS256" || jwk.KeyType != "RSA" || jwk.KeyID == "" {
				continue
			}
			modulus, modulusErr := base64.RawURLEncoding.DecodeString(jwk.Modulus)
			exponent, exponentErr := base64.RawURLEncoding.DecodeString(jwk.Exponent)
			if modulusErr != nil || exponentErr != nil || len(modulus) == 0 || len(exponent) == 0 {
				continue
			}
			keys[jwk.KeyID] = rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(new(big.Int).SetBytes(exponent).Int64())}
		}
		v.keys, v.keysURI, v.keysTTL = keys, v.config.JWKSURI, now.Add(time.Hour)
	}
	key, ok := v.keys[keyID]
	if !ok {
		return riscConfiguration{}, rsa.PublicKey{}, errors.New("clave RISC desconocida")
	}
	return v.config, key, nil
}

func (v *RISCVerifier) fetchJSON(ctx context.Context, url string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := v.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("estado %d", response.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 256*1024)).Decode(target)
}

func decodeJWTPart(raw string, target any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(decoded, target)
}

func anyEmpty(values []string) bool {
	for _, value := range values {
		if value == "" {
			return true
		}
	}
	return false
}
