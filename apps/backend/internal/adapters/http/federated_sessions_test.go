package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCreateGoogleChallengeReturnsCreated(t *testing.T) {
	service := federated.NewService(testFederatedRepository{}, nil)
	recorder := httptest.NewRecorder()

	createGoogleChallenge(service).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/google-login-challenges", nil))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestFederatedHandlersRecordOnlySafeRootFailureReasons(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	validIdentity := federated.Identity{Issuer: federated.GoogleIssuer, Subject: "google-subject", Email: "person@example.test", Nonce: "nonce", EmailVerified: true}
	const challengeID = "019abcde-1111-7111-8111-111111111111"
	for _, test := range []struct {
		name    string
		route   string
		handler http.Handler
		body    string
		reason  string
	}{
		{
			name: "challenge database failure", route: "POST /v1/google-login-challenges", body: "",
			handler: createGoogleChallenge(federated.NewService(testFederatedRepository{challengeErr: errors.New("database refused secret-nonce")}, nil)), reason: "database.query_failed",
		},
		{
			name: "google unavailable", route: "POST /v1/google-sessions", body: "",
			handler: http.HandlerFunc(unavailableFederatedLogin), reason: "identity.google_unavailable",
		},
		{
			name: "session validation", route: "POST /v1/google-sessions", body: `{"challengeId":"not-a-uuid","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{}, testGoogleVerifier{identity: validIdentity}), sessionCookies(true)), reason: "validation.rejected",
		},
		{
			name: "session email conflict", route: "POST /v1/google-sessions", body: `{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{authenticateErr: federated.ErrEmailConflict}, testGoogleVerifier{identity: validIdentity}), sessionCookies(true)), reason: "identity.email_conflict",
		},
		{
			name: "session invalid google proof", route: "POST /v1/google-sessions", body: `{"challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token","sessionTransport":"bearer"}`,
			handler: createGoogleSession(federated.NewService(testFederatedRepository{}, testGoogleVerifier{}), sessionCookies(true)), reason: "credential.google_challenge_invalid",
		},
		{
			name: "link identity conflict", route: "POST /v1/me/google-identities", body: `{"ticket":"secret-ticket","challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token"}`,
			handler: createGoogleIdentity(federated.NewService(testFederatedRepository{addWithTicketErr: federated.ErrIdentityConflict}, testGoogleVerifier{identity: validIdentity})), reason: "identity.google_conflict",
		},
		{
			name: "link invalid reauthentication", route: "POST /v1/me/google-identities", body: `{"ticket":"secret-ticket","challengeId":"019abcde-1111-7111-8111-111111111111","idToken":"secret-google-token"}`,
			handler: createGoogleIdentity(federated.NewService(testFederatedRepository{addWithTicketErr: federated.ErrChallengeInvalid}, testGoogleVerifier{identity: validIdentity})), reason: "credential.reauthentication_invalid",
		},
		{
			name: "unlink invalid reauthentication", route: "DELETE /v1/me/google-identities", body: `{"ticket":"secret-ticket"}`,
			handler: deleteGoogleIdentity(federated.NewService(testFederatedRepository{removeWithTicketErr: federated.ErrChallengeInvalid}, nil)), reason: "credential.reauthentication_invalid",
		},
		{
			name: "unlink database failure", route: "DELETE /v1/me/google-identities", body: `{"ticket":"secret-ticket"}`,
			handler: deleteGoogleIdentity(federated.NewService(testFederatedRepository{removeWithTicketErr: errors.New("database rejected session-token")}, nil)), reason: "database.query_failed",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer secret-session-token")
			ctx, span := provider.Tracer("test").Start(request.Context(), test.route)
			test.handler.ServeHTTP(httptest.NewRecorder(), request.WithContext(ctx))
			span.End()

			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); got != test.reason {
				t.Fatalf("failure reason = %q, want %q", got, test.reason)
			}
			if got := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); strings.Contains(got, "secret") || strings.Contains(got, "person@example.test") {
				t.Fatalf("failure reason leaked request data: %q", got)
			}
			exporter.Reset()
		})
	}
}

func testSpanAttribute(attributes []attribute.KeyValue, key string) string {
	for _, item := range attributes {
		if string(item.Key) == key {
			return item.Value.AsString()
		}
	}
	return ""
}
