package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/sessions"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const managementID = "019abcde-1111-7111-8111-111111111111"

type managementAuthenticator struct {
	testAuthenticator
	err    error
	calls  int
	target string
}

func (a *managementAuthenticator) ListSessions(context.Context, string, string) (sessions.Page, error) {
	a.calls++
	return sessions.Page{Items: []sessions.Item{{ID: managementID, CreatedAt: "2026-10-12T12:00:00Z", LastSeenAt: "2026-10-12T12:00:00Z", Platform: "unknown", Client: "unknown", Current: true}}}, a.err
}
func (a *managementAuthenticator) RevokeOtherSessions(_ context.Context, _, target string) error {
	a.calls++
	a.target = target
	return a.err
}

func TestSessionManagementContractAndGuards(t *testing.T) {
	a := &managementAuthenticator{testAuthenticator: testAuthenticator{accountID: managementID}}
	handler := NewHandler(registration.Service{}, nil, a, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins)
	for _, tc := range []struct {
		method, path, auth, origin string
		status                     int
	}{
		{"GET", "/v1/me/sessions", "", "", 401},
		{"GET", "/v1/me/sessions", "bearer", "", 200},
		{"GET", "/v1/me/sessions?cursor=", "bearer", "", 400},
		{"GET", "/v1/me/sessions?cursor=private", "bearer", "", 400},
		{"DELETE", "/v1/me/sessions/invalid", "bearer", "", 400},
		{"DELETE", "/v1/me/sessions/" + managementID, "bearer", "", 204},
		{"DELETE", "/v1/me/sessions/others", "bearer", "", 204},
		{"DELETE", "/v1/me/sessions/others", "cookie", "https://evil.example", 403},
		{"DELETE", "/v1/me/sessions/" + managementID, "cookie", "http://localhost:8082", 204},
	} {
		t.Run(tc.method+tc.path+tc.auth+tc.origin, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			if tc.auth == "bearer" {
				req.Header.Set("Authorization", "Bearer secret")
			}
			if tc.auth == "cookie" {
				req.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "secret", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
				req.Header.Set("Origin", tc.origin)
			}
			before := a.calls
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != tc.status {
				t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
			}
			if tc.status >= 400 && a.calls != before {
				t.Fatal("rejected request reached repository")
			}
			if tc.status == 200 {
				var page sessions.Page
				if err := json.Unmarshal(recorder.Body.Bytes(), &page); err != nil || len(page.Items) != 1 || !page.Items[0].Current {
					t.Fatalf("invalid projection: %v", err)
				}
				if recorder.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("session list must not be cached")
				}
			}
			if tc.path == "/v1/me/sessions/others" && tc.status == 204 && a.target != "" {
				t.Fatal("others interpreted as an ID")
			}
			if len(recorder.Result().Cookies()) != 0 {
				t.Fatal("remote revoke must preserve current cookies")
			}
		})
	}
}

func TestSessionManagementFailuresAreSafeAndCategorized(t *testing.T) {
	for _, revoke := range []bool{false, true} {
		for _, tc := range []struct {
			err    error
			reason string
			status int
		}{
			{errors.New("private SQL token email"), "database.query_failed", 500},
			{&pgconn.PgError{Code: "08006", Message: "private connection"}, "database.unavailable", 500},
			{&pgconn.PgError{Code: "23514", Message: "private constraint"}, "database.constraint_failed", 500},
			{context.DeadlineExceeded, "request.timeout", 500},
			{context.Canceled, "request.cancelled", 200},
			{tournaments.ErrUnauthenticated, "session.invalid", 401},
		} {
			t.Run(tc.reason, func(t *testing.T) {
				exporter := tracetest.NewInMemoryExporter()
				provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
				ctx, span := provider.Tracer("test").Start(t.Context(), "DELETE /v1/me/sessions/{sessionId}")
				if errors.Is(tc.err, context.Canceled) {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				}
				ctx = context.WithValue(ctx, sessionTokenContextKey{}, "secret")
				a := &managementAuthenticator{err: tc.err}
				req := httptest.NewRequestWithContext(ctx, "DELETE", "/v1/me/sessions/others", nil)
				recorder := httptest.NewRecorder()
				manageSessions(a, revoke, true)(recorder, req)
				span.End()
				defer func() { _ = provider.Shutdown(t.Context()) }()
				if recorder.Code != tc.status || strings.Contains(recorder.Body.String(), "private") {
					t.Fatalf("unsafe response %d: %s", recorder.Code, recorder.Body.String())
				}
				if errors.Is(tc.err, context.Canceled) && recorder.Body.Len() != 0 {
					t.Fatal("cancellation wrote a body")
				}
				spans := exporter.GetSpans()
				if len(spans) != 1 {
					t.Fatalf("unexpected child spans: %d", len(spans))
				}
				found := false
				for _, attr := range spans[0].Attributes {
					if string(attr.Key) == "tournaments_manager.failure.reason" && attr.Value.AsString() == tc.reason {
						found = true
					}
					if strings.Contains(attr.Value.AsString(), "private") {
						t.Fatal("private details exported")
					}
				}
				if !found {
					t.Fatalf("missing reason %s", tc.reason)
				}
			})
		}
	}
}

func TestSessionManagementRateLimits(t *testing.T) {
	for _, revoke := range []bool{false, true} {
		a := &managementAuthenticator{}
		limiter := newRequestLimiter(1, time.Minute)
		handler := resourceRateLimit(manageSessions(a, revoke, true), limiter, accountResourceKey)
		for i := 0; i < 2; i++ {
			ctx := context.WithValue(t.Context(), accountContextKey{}, managementID)
			req := httptest.NewRequestWithContext(ctx, "DELETE", "/v1/me/sessions/others", nil)
			rec := httptest.NewRecorder()
			handler(rec, req)
			if i == 1 && (rec.Code != 429 || rec.Header().Get("Retry-After") == "") {
				t.Fatal("missing limit/retry")
			}
		}
		if a.calls != 1 {
			t.Fatal("limited request reached storage")
		}
	}
}

func TestSessionClientCategoriesNeverUseRawHints(t *testing.T) {
	for _, tc := range []struct {
		ua, hint       string
		transport      sessionTransport
		platform, kind string
	}{
		{"Mozilla iPhone Safari/605", "", cookieSession, "ios", "safari"},
		{"Mozilla Windows Chrome/120 Edg/120", "", cookieSession, "windows", "edge"},
		{"private device email", "private token", bearerSession, "unknown", "app"},
		{"private device", "android", bearerSession, "android", "app"},
		{"Mozilla Linux Firefox/120", "", cookieSession, "linux", "firefox"},
	} {
		r := httptest.NewRequestWithContext(t.Context(), "GET", "/", nil)
		r.Header.Set("User-Agent", tc.ua)
		r.Header.Set("X-TM-Client-Platform", tc.hint)
		got := sessionClient(r, tc.transport)
		if got.Platform != tc.platform || got.Kind != tc.kind {
			t.Fatalf("categories %#v", got)
		}
	}
}
