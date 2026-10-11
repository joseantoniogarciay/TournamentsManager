package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type repetitionHTTPRepository struct {
	testCreationRepository
	err   error
	input *tournaments.RepeatInput
}

func (r repetitionHTTPRepository) Repeat(_ context.Context, _, _ string, input tournaments.RepeatInput) (tournaments.Tournament, error) {
	if r.input != nil {
		*r.input = input
	}
	return r.created, r.err
}

func TestRepeatTournamentHTTPResponsesAndSafeRootSpan(t *testing.T) {
	const sourceID = "019abcde-2222-7222-8222-222222222222"
	const accountID = "019abcde-1111-7111-8111-111111111111"
	valid := `{"requestId":"019abcde-3333-7333-8333-333333333333","name":" New edition ","teams":[{"name":" Blue "},{"name":"Red"}]}`
	cases := []struct {
		name, body, reason string
		err                error
		want               int
	}{
		{"success", valid, "", nil, 201},
		{"forbidden", valid, "tournament.forbidden", tournaments.ErrTournamentForbidden, 403},
		{"missing", valid, "tournament.not_found", tournaments.ErrTournamentNotFound, 404},
		{"not completed", valid, "tournament.repetition_conflict", tournaments.ErrTournamentRepetitionConflict, 409},
		{"bad UUID", strings.Replace(valid, "019abcde-3333-7333-8333-333333333333", "invalid", 1), "validation.rejected", nil, 400},
		{"duplicate names", strings.Replace(valid, "Red", "blue", 1), "validation.rejected", nil, 400},
		{"unknown field", strings.Replace(valid, "\"name\":\" New edition \"", "\"name\":\" New edition \",\"secret\":\"private value\"", 1), "validation.rejected", nil, 400},
		{"begin unavailable", valid, "database.unavailable", &pgconn.PgError{Code: "08006", Message: "private value"}, 500},
		{"query failed", valid, "database.query_failed", errors.New("private value: SQL failed"), 500},
		{"constraint failed", valid, "database.constraint_failed", &pgconn.PgError{Code: "23505", Message: "private value"}, 500},
		{"cancelled", valid, "request.cancelled", context.Canceled, 500},
		{"timeout", valid, "request.timeout", context.DeadlineExceeded, 500},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spans := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(spans))
			defer func() { _ = provider.Shutdown(t.Context()) }()
			ctx, span := provider.Tracer("test").Start(context.WithValue(t.Context(), accountContextKey{}, accountID), "POST /v1/tournaments/{tournamentId}/repeat")
			var received tournaments.RepeatInput
			repo := repetitionHTTPRepository{input: &received, err: c.err, testCreationRepository: testCreationRepository{created: tournaments.Tournament{ID: sourceID, Name: "New edition", Sport: tournaments.SportFootball, State: "published"}}}
			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/v1/tournaments/"+sourceID+"/repeat", strings.NewReader(c.body))
			req.SetPathValue("tournamentId", sourceID)
			recorder := httptest.NewRecorder()
			repeatTournament(tournaments.NewCreationService(repo)).ServeHTTP(recorder, req)
			span.End()
			if recorder.Code != c.want {
				t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
			}
			if strings.Contains(recorder.Body.String(), "private value") {
				t.Fatal("internal details leaked")
			}
			ended := spans.Ended()
			if len(ended) != 1 || leagueFailureReason(ended[0].Attributes()) != c.reason {
				t.Fatalf("root spans: %#v", ended)
			}
			if c.want == 201 && (received.Name != "New edition" || received.Teams[0].Name != "Blue") {
				t.Fatalf("not normalized: %#v", received)
			}
		})
	}
}

func TestRepeatTournamentRequiresSessionAndCookieCSRF(t *testing.T) {
	const accountID = "019abcde-1111-7111-8111-111111111111"
	path := "/v1/tournaments/019abcde-2222-7222-8222-222222222222/repeat"
	for _, c := range []struct {
		name   string
		cookie bool
		want   int
	}{{"no session", false, 401}, {"cookie without CSRF", true, 403}} {
		t.Run(c.name, func(t *testing.T) {
			handler := NewHandler(registration.Service{}, nil, testAuthenticator{accountID: accountID}, tournaments.NewService(testTournamentRepository{}), testAllowedOrigins, tournaments.NewCreationService(testCreationRepository{}))
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(`{}`))
			if c.cookie {
				request.AddCookie(&http.Cookie{Name: "__Host-tm_session", Value: "opaque-session", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
				request.Header.Set("Origin", "https://untrusted.example.test")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != c.want {
				t.Fatalf("status=%d want=%d", response.Code, c.want)
			}
		})
	}
}
