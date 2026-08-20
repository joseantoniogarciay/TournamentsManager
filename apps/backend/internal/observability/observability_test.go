package observability

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestQueryMetadataUsesSQLCNameWithoutExportingSQL(t *testing.T) {
	t.Parallel()

	name, operation := queryMetadata(`-- name: FindLocalAccountForLogin :one
SELECT id, password_hash FROM local_credentials WHERE email = $1`)

	if name != "FindLocalAccountForLogin" {
		t.Fatalf("name = %q, want FindLocalAccountForLogin", name)
	}
	if operation != "SELECT" {
		t.Fatalf("operation = %q, want SELECT", operation)
	}
}

func TestQueryMetadataSkipsOrdinaryComments(t *testing.T) {
	t.Parallel()

	name, operation := queryMetadata("-- internal explanation\n\nWITH created AS (SELECT 1) SELECT * FROM created")

	if name != "" {
		t.Fatalf("name = %q, want empty", name)
	}
	if operation != "WITH" {
		t.Fatalf("operation = %q, want WITH", operation)
	}
}

func TestQueryMetadataUsesManualQueryName(t *testing.T) {
	t.Parallel()

	name, operation := queryMetadata(`-- name: SearchPublicUsernames :many
SELECT username FROM accounts WHERE username LIKE '%' || $1 || '%'`)

	if name != "SearchPublicUsernames" {
		t.Fatalf("name = %q, want SearchPublicUsernames", name)
	}
	if operation != "SELECT" {
		t.Fatalf("operation = %q, want SELECT", operation)
	}
}

func TestPasswordProtectorCreatesSafeSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	protector := PasswordProtector{}
	if _, err := protector.Hash(context.Background(), "test-password"); err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if protector.Verify(context.Background(), "test-password", "not-an-argon2id-verifier") {
		t.Fatal("Verify returned true for an invalid verifier")
	}

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("span count = %d, want 2", len(spans))
	}
	if spans[0].Name != "auth.password.hash" || spans[1].Name != "auth.password.verify" {
		t.Fatalf("span names = %q, %q, want auth.password.hash, auth.password.verify", spans[0].Name, spans[1].Name)
	}
	if got := spans[0].Attributes[0].Value.AsString(); got != "argon2id" {
		t.Fatalf("algorithm = %q, want argon2id", got)
	}
}

func TestMailerCreatesSafeSpans(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		send func(Mailer) error
		span string
	}{
		{name: "verification", send: func(mailer Mailer) error {
			return mailer.SendVerification(context.Background(), "person@example.test", registration.LocaleSpanish, "secret-token")
		}, span: "smtp.send.verification"},
		{name: "password reset", send: func(mailer Mailer) error {
			return mailer.SendPasswordReset(context.Background(), "person@example.test", registration.LocaleSpanish, "secret-token")
		}, span: "smtp.send.password_reset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			previous := otel.GetTracerProvider()
			otel.SetTracerProvider(provider)
			t.Cleanup(func() {
				otel.SetTracerProvider(previous)
				_ = provider.Shutdown(context.Background())
			})

			if err := test.send(Mailer{Next: mailerStub{}}); err != nil {
				t.Fatalf("send() error = %v", err)
			}
			spans := exporter.GetSpans()
			if len(spans) != 1 || spans[0].Name != test.span {
				t.Fatalf("spans = %#v, want one %s span", spans, test.span)
			}
			if len(spans[0].Attributes) != 0 {
				t.Fatalf("attributes = %#v, want no potentially sensitive attributes", spans[0].Attributes)
			}
		})
	}
}

type mailerStub struct{}

func (mailerStub) SendVerification(context.Context, string, registration.Locale, string) error {
	return nil
}
func (mailerStub) SendPasswordReset(context.Context, string, registration.Locale, string) error {
	return nil
}
