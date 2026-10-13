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

func TestTournamentBatchedWritesKeepSafeHTTPFailures(t *testing.T) {
	const id = "019abcde-1111-7111-8111-111111111111"
	for _, operation := range []string{"start", "elimination", "result"} {
		for _, tc := range []struct {
			name, reason string
			err          error
			status       int
		}{
			{"success", "", nil, 200},
			{"validation", "validation.rejected", tournaments.ErrInvalidTournamentInput, 400},
			{"not found", "tournament.not_found", tournaments.ErrTournamentNotFound, 404},
			{"forbidden", "tournament.forbidden", tournaments.ErrTournamentForbidden, 403},
			{"business conflict", "tournament.start_conflict", tournaments.ErrTournamentConflict, 409},
			{"connection", "database.unavailable", &pgconn.PgError{Code: "08006", Message: "private SQL and token"}, 500},
			{"constraint", "database.constraint_failed", &pgconn.PgError{Code: "23503", Message: "private SQL and token"}, 500},
			{"query", "database.query_failed", errors.New("private SQL and token"), 500},
			{"statement timeout", "request.timeout", &pgconn.PgError{Code: "57014", Message: "private SQL and token"}, 500},
			{"lock timeout", "request.timeout", &pgconn.PgError{Code: "55P03", Message: "private SQL and token"}, 500},
			{"deadline", "request.timeout", context.DeadlineExceeded, 500},
			{"cancellation", "request.cancelled", context.Canceled, 500},
		} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				if operation == "result" && tc.name == "forbidden" {
					tc.err = tournaments.ErrMatchResultForbidden
				}
				if tc.name == "business conflict" {
					switch operation {
					case "elimination":
						tc.err, tc.reason = tournaments.ErrTournamentStageTransitionConflict, "tournament.stage_transition_conflict"
					case "result":
						tc.err, tc.reason = tournaments.ErrMatchResultConflict, "tournament.result_conflict"
					}
				}
				repo := testCreationRepository{startErr: tc.err, eliminationErr: tc.err, resultErr: tc.err}
				service := tournaments.NewCreationService(repo)
				handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: testAuthenticator{accountID: id}, TournamentCreation: &service})
				method, path, template, payload := http.MethodPost, "/v1/tournaments/"+id+"/start", "POST /v1/tournaments/{tournamentId}/start", `{"format":"league","roundRobinLegs":2}`
				switch operation {
				case "elimination":
					path, template, payload = "/v1/tournaments/"+id+"/stages/elimination/start", "POST /v1/tournaments/{tournamentId}/stages/elimination/start", ""
					if tc.name == "validation" {
						path = strings.ReplaceAll(path, id, "invalid")
					}
				case "result":
					method, path, template, payload = http.MethodPut, "/v1/tournaments/"+id+"/matches/"+id+"/result", "PUT /v1/tournaments/{tournamentId}/matches/{matchId}/result", `{"homeScore":1,"awayScore":0}`
				}
				exporter := tracetest.NewInMemoryExporter()
				provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
				t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
				ctx, span := provider.Tracer("test").Start(t.Context(), template)
				request := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(payload))
				request.Header.Set("Authorization", "Bearer session-token")
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)
				span.End()
				if recorder.Code != tc.status || strings.Contains(recorder.Body.String(), "private") {
					t.Fatalf("unsafe response: status=%d body=%s", recorder.Code, recorder.Body.String())
				}
				spans := exporter.GetSpans()
				if len(spans) != 1 || spans[0].Name != template || testSpanAttribute(spans[0].Attributes, "tournaments_manager.failure.reason") != tc.reason {
					t.Fatalf("unexpected root span: %+v", spans)
				}
				for _, attr := range spans[0].Attributes {
					if strings.Contains(attr.Value.AsString(), "private") || strings.Contains(attr.Value.AsString(), id) {
						t.Fatal("internal error or identifier leaked to trace")
					}
				}
			})
		}
	}
}
