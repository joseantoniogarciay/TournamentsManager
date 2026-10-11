package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type recoveryContractRepository struct {
	testRegistrationRepository
	eligible bool
	err      error
}

func (r recoveryContractRepository) CreatePasswordReset(context.Context, string, []byte) (string, registration.Locale, bool, error) {
	return "private-marker@example.test", registration.LocaleSpanish, r.eligible, r.err
}

type recoveryContractMailer struct {
	err   error
	calls int
}

func (*recoveryContractMailer) SendVerification(context.Context, string, registration.Locale, string) error {
	return nil
}
func (m *recoveryContractMailer) SendPasswordReset(context.Context, string, registration.Locale, string) error {
	m.calls++
	return m.err
}

func TestPasswordResetRequestResponsesDoNotExposeDependencyDetails(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous); _ = provider.Shutdown(context.Background()) })

	for _, test := range []struct {
		name, body               string
		eligible                 bool
		limit, status, mailCalls int
		databaseErr, smtpErr     error
	}{
		{name: "eligible", body: `{"email":"person@example.test"}`, eligible: true, limit: 10, status: 202, mailCalls: 1},
		{name: "unknown account", body: `{"email":"person@example.test"}`, limit: 10, status: 202},
		{name: "validation", body: `{"email":"invalid"}`, limit: 10, status: 400},
		{name: "rate limit", body: `{"email":"person@example.test"}`, limit: 0, status: 429},
		{name: "database failure", body: `{"email":"person@example.test"}`, limit: 10, status: 500, databaseErr: errors.New("private-marker SQL failure")},
		{name: "database cancellation", body: `{"email":"person@example.test"}`, limit: 10, status: 500, databaseErr: context.Canceled},
		{name: "database timeout", body: `{"email":"person@example.test"}`, limit: 10, status: 500, databaseErr: context.DeadlineExceeded},
		{name: "SMTP failure", body: `{"email":"person@example.test"}`, eligible: true, limit: 10, status: 500, mailCalls: 1, smtpErr: errors.New("private-marker SMTP failure")},
		{name: "SMTP cancellation", body: `{"email":"person@example.test"}`, eligible: true, limit: 10, status: 500, mailCalls: 1, smtpErr: context.Canceled},
		{name: "SMTP timeout", body: `{"email":"person@example.test"}`, eligible: true, limit: 10, status: 500, mailCalls: 1, smtpErr: context.DeadlineExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			mailer := &recoveryContractMailer{err: test.smtpErr}
			service := registration.NewService(recoveryContractRepository{eligible: test.eligible, err: test.databaseErr}, mailer)
			handler := requestPasswordReset(service, newRequestLimiter(test.limit, time.Minute), func(*http.Request) string { return "127.0.0.1" })
			response := httptest.NewRecorder()
			exporter.Reset()
			ctx, span := provider.Tracer("test").Start(context.Background(), "POST /v1/password-resets")
			handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/password-resets", strings.NewReader(test.body)).WithContext(ctx))
			span.End()
			wantReason := ""
			switch test.status {
			case 400:
				wantReason = "validation.rejected"
			case 429:
				wantReason = "rate_limit.exceeded"
			case 500:
				wantReason = "request.failed"
				if errors.Is(test.databaseErr, context.Canceled) || errors.Is(test.smtpErr, context.Canceled) {
					wantReason = "request.cancelled"
				}
				if errors.Is(test.databaseErr, context.DeadlineExceeded) || errors.Is(test.smtpErr, context.DeadlineExceeded) {
					wantReason = "request.timeout"
				}
			}
			spans := exporter.GetSpans()
			if len(spans) != 1 || failureReason(spans[0]) != wantReason {
				t.Fatalf("unexpected root failure reason: want %q", wantReason)
			}
			if len(spans[0].Events) != 0 {
				t.Fatal("raw error exported as span event")
			}
			for _, attribute := range spans[0].Attributes {
				if strings.Contains(attribute.Value.AsString(), "private-marker") {
					t.Fatal("private dependency detail exported")
				}
			}

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if mailer.calls != test.mailCalls {
				t.Fatalf("SMTP calls = %d, want %d", mailer.calls, test.mailCalls)
			}
			for _, secret := range []string{"private-marker", "SQL failure", "SMTP failure", "context canceled", "deadline exceeded"} {
				if strings.Contains(response.Body.String(), secret) {
					t.Errorf("response leaked %q", secret)
				}
			}
			if test.status == 202 && response.Body.Len() != 0 {
				t.Fatal("accepted response should be identical and empty for eligible and unknown accounts")
			}
			if test.status == 429 && response.Header().Get("Retry-After") == "" {
				t.Fatal("missing Retry-After")
			}
		})
	}
}
