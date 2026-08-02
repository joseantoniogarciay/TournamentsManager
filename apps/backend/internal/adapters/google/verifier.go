// Package google adapta la validación de ID tokens de Google al puerto federado.
package google

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

// Verifier valida ID tokens de Google para las audiencias configuradas.
type Verifier struct {
	audiences []string
	validator tokenValidator
}

type tokenValidator interface {
	Validate(context.Context, string, string) (*idtoken.Payload, error)
}

type defaultTokenValidator struct{}

func (defaultTokenValidator) Validate(ctx context.Context, raw, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, raw, audience)
}

// NewVerifier construye un verificador limitado a las audiencias permitidas.
func NewVerifier(audiences []string) Verifier {
	return Verifier{audiences: audiences, validator: defaultTokenValidator{}}
}

// Verify transforma un ID token Google validado en la identidad del dominio.
func (v Verifier) Verify(ctx context.Context, raw string) (federated.Identity, error) {
	var payload *idtoken.Payload
	var err error
	for _, audience := range v.audiences {
		payload, err = v.validator.Validate(ctx, raw, audience)
		if err == nil {
			break
		}
	}
	if err != nil || payload == nil {
		return federated.Identity{}, fmt.Errorf("ID token no válido")
	}
	issuer, _ := payload.Claims["iss"].(string)
	subject, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	nonce, _ := payload.Claims["nonce"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	return federated.Identity{Issuer: strings.TrimSpace(issuer), Subject: strings.TrimSpace(subject), Email: strings.TrimSpace(email), Nonce: nonce, EmailVerified: verified}, nil
}
