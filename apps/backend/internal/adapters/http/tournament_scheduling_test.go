package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type httpSchedulingRepository struct {
	err   error
	calls *int
}

func (r httpSchedulingRepository) SetVenue(_ context.Context, _, _, venue string) (tournaments.Tournament, error) {
	*r.calls++
	return tournaments.Tournament{Venue: venue}, r.err
}
func (r httpSchedulingRepository) SetMatchSchedule(_ context.Context, _, _, _ string, input tournaments.ScheduleInput) (tournaments.Tournament, error) {
	*r.calls++
	return tournaments.Tournament{Matches: []tournaments.Match{{ScheduledDate: input.ScheduledDate, Venue: input.Venue}}}, r.err
}

func TestSchedulingHTTPContractAndSafeFailureReasons(t *testing.T) {
	t.Parallel()
	const id = "019abcde-1111-7111-8111-111111111111"
	for _, match := range []bool{false, true} {
		for _, tc := range []struct {
			name, body string
			err        error
			status     int
			reason     string
		}{
			{"success", "", nil, 200, ""},
			{"missing", "{}", nil, 400, "validation.rejected"},
			{"null", "{\"venue\":null}", nil, 400, "validation.rejected"},
			{"unknown", "{\"venue\":\"Private court\",\"extra\":true}", nil, 400, "validation.rejected"},
			{"validation", "", tournaments.ErrInvalidTournamentInput, 400, "validation.rejected"},
			{"forbidden", "", tournaments.ErrTournamentForbidden, 403, "tournament.forbidden"},
			{"unavailable", "", tournaments.ErrTournamentNotFound, 404, "tournament.not_found"},
			{"closed", "", tournaments.ErrSchedulingConflict, 409, "tournament.scheduling_conflict"},
			{"connection", "", &pgconn.PgError{Code: "08006", Message: "private database credentials"}, 500, "database.unavailable"},
			{"constraint", "", &pgconn.PgError{Code: "23514", Message: "Private court"}, 500, "database.constraint_failed"},
			{"query", "", errors.New("private SQL"), 500, "database.query_failed"},
			{"timeout", "", context.DeadlineExceeded, 500, "request.timeout"},
			{"cancel", "", context.Canceled, 500, "request.cancelled"},
		} {
			t.Run(tc.name+map[bool]string{false: " venue", true: " match"}[match], func(t *testing.T) {
				calls := 0
				scheduling := tournaments.NewSchedulingService(httpSchedulingRepository{err: tc.err, calls: &calls})
				handler := NewHandlerWithConfig(HandlerConfig{CookieSecure: true, CORSAllowedOrigins: testAllowedOrigins}, HandlerDependencies{Authenticator: testAuthenticator{accountID: id}, TournamentScheduling: &scheduling})
				path := "/v1/tournaments/" + id + "/venue"
				template := "PUT /v1/tournaments/{tournamentId}/venue"
				payload := `{"venue":"Private court"}`
				if match {
					path = "/v1/tournaments/" + id + "/matches/" + id + "/schedule"
					template = "PUT /v1/tournaments/{tournamentId}/matches/{matchId}/schedule"
					payload = `{"scheduledDate":"2026-10-18","scheduledTime":"18:30","timeZone":"Europe/Madrid","venue":"Private court"}`
				}
				if tc.body != "" {
					payload = tc.body
				}
				exporter := tracetest.NewInMemoryExporter()
				provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
				ctx, span := provider.Tracer("test").Start(t.Context(), template)
				request := httptest.NewRequestWithContext(ctx, http.MethodPut, path, strings.NewReader(payload))
				request.Header.Set("Authorization", "Bearer session-token")
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				span.End()
				if recorder.Code != tc.status {
					t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
				}
				spans := exporter.GetSpans()
				if len(spans) != 1 || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != tc.reason {
					t.Fatalf("unsafe trace %+v", spans)
				}
				for _, attr := range spans[0].Attributes {
					if strings.Contains(attr.Value.AsString(), "Private") || strings.Contains(attr.Value.AsString(), "private") || strings.Contains(attr.Value.AsString(), "2026") {
						t.Fatal("input leaked in trace")
					}
				}
				if tc.status != 200 && (strings.Contains(recorder.Body.String(), "private") || strings.Contains(recorder.Body.String(), "Private")) {
					t.Fatal("raw failure leaked")
				}
				if tc.status == 400 && tc.err == nil && calls != 0 {
					t.Fatal("invalid request reached persistence")
				}
			})
		}
	}
}

func TestSchedulingRequiresSessionAndStrictBody(t *testing.T) {
	t.Parallel()
	const id = "019abcde-1111-7111-8111-111111111111"
	calls := 0
	service := tournaments.NewSchedulingService(httpSchedulingRepository{calls: &calls})
	var handler http.Handler
	for _, tc := range []struct {
		body, auth string
		status     int
	}{
		{`{"venue":"Club"}`, "", 401},
		{`{"venue":"Club"} {}`, "Bearer session-token", 400},
		{`{"venue":"` + strings.Repeat("x", 5000) + `"}`, "Bearer session-token", 400},
	} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/v1/tournaments/"+id+"/venue", strings.NewReader(tc.body))
		if tc.auth == "" {
			handler = NewHandlerWithConfig(HandlerConfig{CookieSecure: true, CORSAllowedOrigins: testAllowedOrigins}, HandlerDependencies{Authenticator: testAuthenticator{}, TournamentScheduling: &service})
		} else {
			handler = NewHandlerWithConfig(HandlerConfig{CookieSecure: true, CORSAllowedOrigins: testAllowedOrigins}, HandlerDependencies{Authenticator: testAuthenticator{accountID: id}, TournamentScheduling: &service})
		}
		request.Header.Set("Authorization", tc.auth)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != tc.status {
			t.Fatalf("status %d", recorder.Code)
		}
	}
	if calls != 0 {
		t.Fatal("rejected request reached persistence")
	}
}
