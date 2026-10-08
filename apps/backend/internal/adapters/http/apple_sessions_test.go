package http

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type httpAppleRepository struct {
	challengeErr, callbackErr, completeErr, sessionErr error
	platform                                           string
}

func (r httpAppleRepository) CreateAppleChallenge(context.Context, []byte, []byte, []byte, string, time.Time) (string, error) {
	return "019abcde-1111-7111-8111-111111111111", r.challengeErr
}
func (r httpAppleRepository) ClaimAppleCallback(context.Context, []byte) (federated.AppleCallback, error) {
	hash := sha256.Sum256([]byte("apple-login-nonce:nonce"))
	return federated.AppleCallback{ID: "019abcde-1111-7111-8111-111111111111", Platform: r.platform, NonceHash: hash[:]}, r.callbackErr
}
func (r httpAppleRepository) CompleteAppleCallback(context.Context, string, federated.Identity) error {
	return r.completeErr
}
func (r httpAppleRepository) AuthenticateApple(context.Context, string, []byte, *federated.Registration, *federated.Draft, []byte, []byte) (federated.Session, error) {
	return federated.Session{AccountID: "account", Username: "person"}, r.sessionErr
}

type httpAppleProvider struct{ err error }

func (httpAppleProvider) AuthorizationURL(string, string) string {
	return "https://appleid.apple.com/auth/authorize"
}
func (p httpAppleProvider) Exchange(context.Context, string) (federated.Identity, error) {
	return federated.Identity{Issuer: federated.AppleIssuer, Subject: "subject", Email: "private@example.test", EmailVerified: true, Nonce: "nonce"}, p.err
}

func TestAppleHandlersStatusesAndSafeRootReasons(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous); _ = provider.Shutdown(context.Background()) })
	config := HandlerConfig{AppleWebReturnURL: "https://dev.fasttourney.com/oauth/apple-complete", AppleNativeReturnURL: "fasttourney-dev://oauth/apple-complete"}
	sessionBody := `{"challengeId":"019abcde-1111-7111-8111-111111111111","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","sessionTransport":"bearer"}`
	tests := []struct {
		name, route, body, contentType, reason string
		status                                 int
		handler                                http.Handler
	}{
		{name: "challenge success", route: "POST /v1/apple-login-challenges", body: `{"platform":"android","proofChallenge":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`, status: 201, handler: createAppleChallenge(federated.NewAppleService(httpAppleRepository{}, httpAppleProvider{}), config)},
		{name: "validation", route: "POST /v1/apple-login-challenges", body: `{"platform":"https://evil.test","proofChallenge":"secret"}`, reason: "validation.rejected", status: 400, handler: createAppleChallenge(federated.NewAppleService(httpAppleRepository{}, httpAppleProvider{}), config)},
		{name: "challenge database failure", route: "POST /v1/apple-login-challenges", body: `{"platform":"web","proofChallenge":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`, reason: "database.query_failed", status: 500, handler: createAppleChallenge(federated.NewAppleService(httpAppleRepository{challengeErr: errors.New("private database detail")}, httpAppleProvider{}), config)},
		{name: "disabled", route: "POST /v1/apple-sessions", reason: "identity.apple_unavailable", status: 503, handler: http.HandlerFunc(unavailableAppleLogin)},
		{name: "session success", route: "POST /v1/apple-sessions", body: sessionBody, status: 200, handler: createAppleSession(federated.NewAppleService(httpAppleRepository{}, httpAppleProvider{}), sessionCookies(true))},
		{name: "registration required", route: "POST /v1/apple-sessions", body: sessionBody, status: 202, handler: createAppleSession(federated.NewAppleService(httpAppleRepository{sessionErr: federated.ErrRegistration}, httpAppleProvider{}), sessionCookies(true))},
		{name: "email conflict", route: "POST /v1/apple-sessions", body: sessionBody, reason: "identity.email_conflict", status: 409, handler: createAppleSession(federated.NewAppleService(httpAppleRepository{sessionErr: federated.ErrEmailConflict}, httpAppleProvider{}), sessionCookies(true))},
		{name: "invalid proof", route: "POST /v1/apple-sessions", body: sessionBody, reason: "credential.apple_challenge_invalid", status: 400, handler: createAppleSession(federated.NewAppleService(httpAppleRepository{sessionErr: federated.ErrChallengeInvalid}, httpAppleProvider{}), sessionCookies(true))},
		{name: "session database failure", route: "POST /v1/apple-sessions", body: sessionBody, reason: "database.query_failed", status: 500, handler: createAppleSession(federated.NewAppleService(httpAppleRepository{sessionErr: errors.New("private database detail")}, httpAppleProvider{}), sessionCookies(true))},
		{name: "callback success", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&code=private-code", contentType: "application/x-www-form-urlencoded", status: 303, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{platform: "android"}, httpAppleProvider{}), config)},
		{name: "callback denial", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&error=access_denied", contentType: "application/x-www-form-urlencoded", status: 303, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{platform: "ios"}, httpAppleProvider{}), config)},
		{name: "provider failure", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&code=private-code", contentType: "application/x-www-form-urlencoded", reason: "identity.provider_unavailable", status: 303, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{platform: "web"}, httpAppleProvider{err: federated.ErrAppleUnavailable}), config)},
		{name: "JWKS rejected", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&code=private-code", contentType: "application/x-www-form-urlencoded", reason: "credential.apple_challenge_invalid", status: 303, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{platform: "web"}, httpAppleProvider{err: federated.ErrChallengeInvalid}), config)},
		{name: "callback database failure", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&code=private-code", contentType: "application/x-www-form-urlencoded", reason: "database.query_failed", status: 303, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{platform: "web", completeErr: errors.New("private subject detail")}, httpAppleProvider{}), config)},
		{name: "callback replay", route: "POST /v1/apple-callback", body: "state=" + strings.Repeat("s", 43) + "&code=private-code", contentType: "application/x-www-form-urlencoded", reason: "credential.apple_challenge_invalid", status: 400, handler: receiveAppleCallback(federated.NewAppleService(httpAppleRepository{callbackErr: federated.ErrChallengeInvalid}, httpAppleProvider{}), config)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			ctx, span := provider.Tracer("test").Start(request.Context(), test.route)
			test.handler.ServeHTTP(recorder, request.WithContext(ctx))
			span.End()
			if recorder.Code != test.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("spans=%d", len(spans))
			}
			if reason := testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason"); reason != test.reason {
				t.Fatalf("reason=%s want=%s", reason, test.reason)
			}
			for _, value := range []string{recorder.Body.String(), recorder.Header().Get("Location"), testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason")} {
				if strings.Contains(value, "private") {
					t.Fatalf("unsafe output=%s", value)
				}
			}
			if recorder.Code == 303 {
				location, err := url.Parse(recorder.Header().Get("Location"))
				if err != nil {
					t.Fatal(err)
				}
				if location.Query().Get("challengeId") == "" || recorder.Header().Get("Set-Cookie") != "" {
					t.Fatal("callback issued credentials")
				}
				if test.name == "callback denial" && location.Query().Get("status") != "cancelled" {
					t.Fatal("denial not silent")
				}
			}
			exporter.Reset()
		})
	}
}

func TestSocialRateLimitAndCancellation(t *testing.T) {
	called := 0
	handler := socialRateLimit(func(w http.ResponseWriter, r *http.Request) { called++; w.WriteHeader(201) }, newRequestLimiter(1, time.Minute), func(*http.Request) string { return "loopback" })
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
	limited := httptest.NewRecorder()
	handler(limited, httptest.NewRequest(http.MethodPost, "/", nil))
	if called != 1 || limited.Code != 429 || limited.Header().Get("Retry-After") == "" {
		t.Fatalf("rate limit=%d %d", called, limited.Code)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	recorder := httptest.NewRecorder()
	writeAppleFailure(recorder, httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx), context.Canceled)
	if recorder.Body.Len() != 0 {
		t.Fatal("intentional cancellation wrote feedback")
	}
}

func TestAppleCallbackBypassesOnlyItsOwnProviderOrigin(t *testing.T) {
	service := federated.NewAppleService(httpAppleRepository{}, httpAppleProvider{})
	handler := NewHandlerWithConfig(HandlerConfig{CORSAllowedOrigins: []string{"https://dev.fasttourney.com"}}, HandlerDependencies{Apple: &service})
	for _, path := range []string{"/v1/apple-callback", "/v1/apple-sessions"} {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader("bad"))
		request.Header.Set("Origin", "https://appleid.apple.com")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		expected := 400
		if path != "/v1/apple-callback" {
			expected = 403
		}
		if recorder.Code != expected {
			t.Fatalf("%s status=%d", path, recorder.Code)
		}
		if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("provider callback expanded CORS")
		}
	}
}
