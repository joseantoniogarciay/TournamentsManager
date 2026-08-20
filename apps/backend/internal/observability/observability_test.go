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

func TestMailerCreatesSafeSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	err := (Mailer{Next: mailerStub{}}).SendVerification(context.Background(), "person@example.test", registration.LocaleSpanish, "secret-token")
	if err != nil {
		t.Fatalf("SendVerification() error = %v", err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 || spans[0].Name != "smtp.send.verification" {
		t.Fatalf("spans = %#v, want one smtp.send.verification span", spans)
	}
	if len(spans[0].Attributes) != 0 {
		t.Fatalf("attributes = %#v, want no potentially sensitive attributes", spans[0].Attributes)
	}
}

type mailerStub struct{}

func (mailerStub) SendVerification(context.Context, string, registration.Locale, string) error {
	return nil
}
func (mailerStub) SendPasswordReset(context.Context, string, registration.Locale, string) error {
	return nil
}
