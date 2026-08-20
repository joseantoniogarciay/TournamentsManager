package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type accessObservabilityRepository struct{ removeErr, setErr error }

func (r accessObservabilityRepository) CurrentPasswordHash(context.Context, string) (string, error) {
	return "", errors.New("not used")
}
func (accessObservabilityRepository) CreateReauthenticationTicket(context.Context, string, []byte) error {
	return errors.New("not used")
}
func (r accessObservabilityRepository) ConsumeReauthenticationTicketAndSetPassword(context.Context, string, []byte, string) error {
	return r.setErr
}
func (r accessObservabilityRepository) ConsumeReauthenticationTicketAndRemovePassword(context.Context, string, []byte) error {
	return r.removeErr
}

type deletionObservabilityAuthenticator struct{ testAuthenticator }

func (deletionObservabilityAuthenticator) ScheduleAccountDeletion(context.Context, string) (time.Time, error) {
	return time.Time{}, postgres.ErrAccountHasOwnedLeagues
}

func TestAccessEndpointFailuresUseClosedSafeReasons(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	registrationService := registration.NewService(testRegistrationRepository{}, nil)
	federatedService := federated.NewService(
		testFederatedRepository{reauthenticationErr: federated.ErrIdentityConflict},
		testGoogleVerifier{identity: federated.Identity{Email: "person@example.test", EmailVerified: true, Issuer: federated.GoogleIssuer, Nonce: "nonce", Subject: "other-google-account"}},
	)
	accessService := access.NewService(accessObservabilityRepository{setErr: errors.New("ticket absent"), removeErr: access.ErrLastAccessMethod})

	tests := []struct {
		name    string
		handler http.Handler
		request *http.Request
		want    string
	}{
		{
			name:    "local session validation",
			handler: createLocalSession(registrationService, newLoginLimiter(10, time.Minute), sessionCookies(false), func(*http.Request) string { return "127.0.0.1" }),
			request: httptest.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(`{"email":"invalid"}`)),
			want:    "validation.rejected",
		},
		{
			name:    "local session credentials",
			handler: createLocalSession(registrationService, newLoginLimiter(10, time.Minute), sessionCookies(false), func(*http.Request) string { return "127.0.0.1" }),
			request: httptest.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","sessionTransport":"bearer"}`)),
			want:    "authentication.credentials_rejected",
		},
		{
			name:    "local session rate limited",
			handler: createLocalSession(registrationService, newLoginLimiter(0, time.Minute), sessionCookies(false), func(*http.Request) string { return "127.0.0.1" }),
			request: httptest.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(`{"email":"person@example.test","password":"correct horse battery staple","sessionTransport":"bearer"}`)),
			want:    "rate_limit.exceeded",
		},
		{
			name:    "refresh invalid",
			handler: refreshSession(registrationService, sessionCookies(false)),
			request: withBearer(httptest.NewRequest(http.MethodPost, "/v1/sessions/refresh", nil)),
			want:    "session.refresh_invalid",
		},
		{
			name:    "reauthentication identity conflict",
			handler: createReauthenticationTicket(access.Service{}, &federatedService),
			request: withAccount(withBearer(httptest.NewRequest(http.MethodPost, "/v1/me/reauthentication-tickets", strings.NewReader(`{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"google-id-token","purpose":"set-local-password"}`)))),
			want:    "reauthentication.identity_conflict",
		},
		{
			name:    "set local credential invalid ticket",
			handler: putLocalCredential(accessService),
			request: withBearer(httptest.NewRequest(http.MethodPut, "/v1/me/local-credential", strings.NewReader(`{"ticket":"ticket","password":"correct horse battery staple"}`))),
			want:    "reauthentication.invalid",
		},
		{
			name:    "remove final access method",
			handler: deleteLocalCredential(accessService),
			request: withBearer(httptest.NewRequest(http.MethodDelete, "/v1/me/local-credential", strings.NewReader(`{"ticket":"ticket"}`))),
			want:    "access_method.last_remaining",
		},
		{
			name:    "account deletion owns leagues",
			handler: scheduleAccountDeletion(deletionObservabilityAuthenticator{testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}}, sessionCookies(false)),
			request: withAccount(httptest.NewRequest(http.MethodDelete, "/v1/me/account", nil)),
			want:    "account.deletion_owned_leagues",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exporter.Reset()
			ctx, span := provider.Tracer("test").Start(test.request.Context(), test.request.Method+" "+test.request.URL.Path)
			test.handler.ServeHTTP(httptest.NewRecorder(), test.request.WithContext(ctx))
			span.End()
			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := failureReason(spans[0]); got != test.want {
				t.Errorf("failure reason = %q, want %q", got, test.want)
			}
			if len(spans[0].Events) != 0 {
				t.Errorf("events = %#v, want no raw error event", spans[0].Events)
			}
		})
	}
}

func withBearer(request *http.Request) *http.Request {
	request.Header.Set("Authorization", "Bearer session-token")
	return request
}

func withAccount(request *http.Request) *http.Request {
	return request.WithContext(context.WithValue(request.Context(), accountContextKey{}, "019abcde-2222-7222-8222-222222222222"))
}

func failureReason(span tracetest.SpanStub) string {
	for _, attribute := range span.Attributes {
		if string(attribute.Key) == "tournaments_manager.failure.reason" {
			return attribute.Value.AsString()
		}
	}
	return ""
}
