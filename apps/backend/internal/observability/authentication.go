package observability

import (
	"context"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PasswordProtector measures local Argon2id operations without recording a
// password, verifier, or account identifier.
type PasswordProtector struct{}

func (PasswordProtector) Hash(ctx context.Context, password string) (string, error) {
	_, span := otel.Tracer(serviceName+"/authentication").Start(ctx, "auth.password.hash", trace.WithAttributes(attribute.String("auth.password.algorithm", "argon2id")))
	defer span.End()
	return registration.HashPassword(password)
}

func (PasswordProtector) Verify(ctx context.Context, password, encoded string) bool {
	_, span := otel.Tracer(serviceName+"/authentication").Start(ctx, "auth.password.verify", trace.WithAttributes(attribute.String("auth.password.algorithm", "argon2id")))
	defer span.End()
	return registration.VerifyPassword(password, encoded)
}
