package observability

import (
	"context"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
)

// Mailer measures SMTP delivery without recording recipient, locale, token, or message content.
type Mailer struct{ Next registration.Mailer }

func (m Mailer) SendVerification(ctx context.Context, recipient string, locale registration.Locale, token string) error {
	return m.send(ctx, "smtp.send.verification", func() error { return m.Next.SendVerification(ctx, recipient, locale, token) })
}

func (m Mailer) SendPasswordReset(ctx context.Context, recipient string, locale registration.Locale, token string) error {
	return m.send(ctx, "smtp.send.password_reset", func() error { return m.Next.SendPasswordReset(ctx, recipient, locale, token) })
}

func (Mailer) send(ctx context.Context, name string, send func() error) error {
	_, span := otel.Tracer(serviceName+"/smtp").Start(ctx, name)
	defer span.End()
	if err := send(); err != nil {
		recordFailure(span, smtpFailureReason(err), "SMTP delivery failed")
		return err
	}
	return nil
}
