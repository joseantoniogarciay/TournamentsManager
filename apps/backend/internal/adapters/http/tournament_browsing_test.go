package http

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type browsingFake struct{ err error }

func (r browsingFake) PersonalMatches(context.Context, string, string, string, int) (tournaments.PersonalMatchPage, error) {
	return tournaments.PersonalMatchPage{Items: []tournaments.PersonalMatch{}}, r.err
}
func (r browsingFake) Editions(context.Context, string, string, int) ([]tournaments.Edition, error) {
	return []tournaments.Edition{}, r.err
}
func TestBrowsingSafeResponsesAndReasons(t *testing.T) {
	for _, personal := range []bool{true, false} {
		for _, tc := range []struct {
			name   string
			err    error
			status int
			reason string
		}{
			{"success", nil, 200, ""}, {"validation", tournaments.ErrInvalidPage, 400, "validation.rejected"}, {"missing", tournaments.ErrTournamentNotFound, 404, "tournament.not_found"},
			{"database", &pgconn.PgError{Code: "08006", Message: "private password"}, 500, "database.unavailable"}, {"constraint", &pgconn.PgError{Code: "23514", Message: "private SQL"}, 500, "database.constraint_failed"}, {"query", errors.New("private SQL"), 500, "database.query_failed"}, {"timeout", context.DeadlineExceeded, 500, "request.timeout"}, {"cancel", context.Canceled, 500, "request.cancelled"},
		} {
			t.Run(tc.name+map[bool]string{true: " personal", false: " editions"}[personal], func(t *testing.T) {
				service := tournaments.NewBrowsingService(browsingFake{tc.err})
				handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
				path := "/v1/tournaments/019abcde-1111-7111-8111-111111111111/editions"
				if personal {
					path = "/v1/me/matches?view=pending"
				}
				exporter := tracetest.NewInMemoryExporter()
				provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
				defer func() { _ = provider.Shutdown(context.Background()) }()
				ctx, span := provider.Tracer("test").Start(context.Background(), "GET /v1/collection")
				request := httptest.NewRequestWithContext(ctx, "GET", path, nil)
				request.Header.Set("Authorization", "Bearer token")
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				span.End()
				if recorder.Code != tc.status {
					t.Fatalf("%d %s", recorder.Code, recorder.Body.String())
				}
				spans := exporter.GetSpans()
				if len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != tc.reason {
					t.Fatalf("spans %+v", spans)
				}
				if strings.Contains(recorder.Body.String(), "private") {
					t.Fatal("raw error leaked")
				}
			})
		}
	}
}
func TestPersonalBrowsingRequiresSessionAndValidQuery(t *testing.T) {
	service := tournaments.NewBrowsingService(browsingFake{})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
	for _, tc := range []struct {
		query, auth string
		want        int
	}{{"?view=pending", "", 401}, {"?view=other", "Bearer token", 400}, {"?view=pending&limit=51", "Bearer token", 400}, {"?view=pending&cursor=invalid", "Bearer token", 400}} {
		request := httptest.NewRequestWithContext(context.Background(), "GET", "/v1/me/matches"+tc.query, nil)
		request.Header.Set("Authorization", tc.auth)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Fatalf("status %d expected %d", response.Code, tc.want)
		}
	}
}

func TestBrowsingCancellationDoesNotWriteFeedbackBody(t *testing.T) {
	service := tournaments.NewBrowsingService(browsingFake{context.Canceled})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
	for _, path := range []string{"/v1/me/matches?view=pending", "/v1/tournaments/019abcde-1111-7111-8111-111111111111/editions"} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		request := httptest.NewRequestWithContext(ctx, "GET", path, nil)
		request.Header.Set("Authorization", "Bearer token")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Body.Len() != 0 {
			t.Fatalf("cancelled request wrote feedback: %s", response.Body.String())
		}
	}
}

func TestBrowsingExplicitLimitMatchesContract(t *testing.T) {
	service := tournaments.NewBrowsingService(browsingFake{})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
	for _, path := range []string{"/v1/me/matches?view=pending", "/v1/tournaments/019abcde-1111-7111-8111-111111111111/editions?"} {
		for _, tc := range []struct {
			query string
			want  int
		}{{"", 200}, {"&limit=1", 200}, {"&limit=50", 200}, {"&limit=", 400}, {"&limit=0", 400}, {"&limit=-1", 400}, {"&limit=51", 400}} {
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			ctx, span := provider.Tracer("test").Start(context.Background(), "GET /v1/collection")
			request := httptest.NewRequestWithContext(ctx, "GET", path+tc.query, nil)
			request.Header.Set("Authorization", "Bearer token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			span.End()
			spans := exporter.GetSpans()
			if response.Code != tc.want {
				t.Fatalf("%s%s: status %d expected %d", path, tc.query, response.Code, tc.want)
			}
			if tc.want == 400 && (testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != "validation.rejected" || response.Header().Get("Content-Type") != "application/problem+json") {
				t.Fatalf("unsafe validation response or missing reason: %s%s", path, tc.query)
			}
			_ = provider.Shutdown(context.Background())
		}
	}
}

func (r browsingFake) PersonalProfile(context.Context, string, string, string, string, int) (tournaments.PersonalProfile, error) {
	return tournaments.PersonalProfile{Items: []tournaments.ProfileMatch{}, Statistics: []tournaments.SportStatistics{}}, r.err
}

func TestProfileSafeResponsesAndReasons(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		reason string
	}{
		{"success", nil, 200, ""}, {"validation", tournaments.ErrInvalidPage, 400, "validation.rejected"},
		{"database", &pgconn.PgError{Code: "08006", Message: "private password"}, 500, "database.unavailable"},
		{"constraint", &pgconn.PgError{Code: "23514", Message: "private SQL"}, 500, "database.constraint_failed"},
		{"query", errors.New("private SQL"), 500, "database.query_failed"},
		{"timeout", context.DeadlineExceeded, 500, "request.timeout"},
		{"statement", &pgconn.PgError{Code: "57014", Message: "private statement"}, 500, "request.timeout"},
		{"lock", &pgconn.PgError{Code: "55P03", Message: "private lock"}, 500, "request.timeout"},
		{"cancel", context.Canceled, 500, "request.cancelled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := tournaments.NewBrowsingService(browsingFake{tc.err})
			handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
			exporter := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			defer func() { _ = provider.Shutdown(context.Background()) }()
			ctx, span := provider.Tracer("test").Start(t.Context(), "GET /v1/me/profile")
			req := httptest.NewRequestWithContext(ctx, "GET", "/v1/me/profile", nil)
			req.Header.Set("Authorization", "Bearer token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			span.End()
			spans := exporter.GetSpans()
			if response.Code != tc.status || strings.Contains(response.Body.String(), "private") || len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != tc.reason {
				t.Fatalf("response %d %s spans %+v", response.Code, response.Body.String(), spans)
			}
			if tc.status == 200 && response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("profile can be cached")
			}
		})
	}
}

func TestProfileAuthenticationValidationQuotaAndCancellation(t *testing.T) {
	service := tournaments.NewBrowsingService(browsingFake{})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &service})
	for _, tc := range []struct {
		query, auth string
		want        int
	}{{"", "", 401}, {"?sport=unknown", "Bearer token", 400}, {"?cursor=invalid", "Bearer token", 400}, {"?accountId=other", "Bearer token", 400}, {"?limit=0", "Bearer token", 400}, {"?limit=51", "Bearer token", 400}, {"?tournament=" + strings.Repeat("a", 121), "Bearer token", 400}, {"?limit=1", "Bearer token", 200}} {
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/v1/me/profile"+tc.query, nil)
		req.Header.Set("Authorization", tc.auth)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != tc.want {
			t.Fatalf("%s: %d %s", tc.query, res.Code, res.Body.String())
		}
	}
	// Both personal read routes share the same per-account quota.
	for range 60 {
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/v1/me/profile", nil)
		req.Header.Set("Authorization", "Bearer token")
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	req := httptest.NewRequestWithContext(t.Context(), "GET", "/v1/me/profile", nil)
	req.Header.Set("Authorization", "Bearer token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 429 {
		t.Fatalf("quota %d", res.Code)
	}
	cancelledService := tournaments.NewBrowsingService(browsingFake{context.Canceled})
	cancelledHandler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"}, TournamentBrowsing: &cancelledService})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	req = httptest.NewRequestWithContext(ctx, "GET", "/v1/me/profile", nil)
	req.Header.Set("Authorization", "Bearer token")
	res = httptest.NewRecorder()
	cancelledHandler.ServeHTTP(res, req)
	if res.Body.Len() != 0 {
		t.Fatal(res.Body.String())
	}
}
