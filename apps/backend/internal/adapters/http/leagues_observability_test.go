package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRecordLeagueFailureUsesClosedSafeReasons(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"not found", leagues.ErrLeagueNotFound, "league.not_found"},
		{"forbidden", leagues.ErrLeagueForbidden, "league.forbidden"},
		{"start conflict", leagues.ErrLeagueConflict, "league.start_conflict"},
		{"team conflict", leagues.ErrLeagueTeamConflict, "league.team_conflict"},
		{"withdrawal conflict", leagues.ErrLeagueWithdrawalConflict, "league.withdrawal_conflict"},
		{"result conflict", leagues.ErrMatchResultConflict, "league.result_conflict"},
		{"administrator conflict", leagues.ErrLeagueAdministratorConflict, "league.administrator_conflict"},
		{"ownership conflict", leagues.ErrLeagueOwnershipTransferConflict, "league.ownership_transfer_conflict"},
		{"cancellation conflict", leagues.ErrLeagueCancellationConflict, "league.cancellation_conflict"},
		{"completion conflict", leagues.ErrLeagueCompletionConflict, "league.completion_conflict"},
		{"invalid input", leagues.ErrInvalidLeagueInput, "validation.rejected"},
		{"cancelled", context.Canceled, "request.cancelled"},
		{"database fallback", errors.New("postgres: account 019abcde-2222-7222-8222-222222222222 failed"), "database.query_failed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			ctx, span := provider.Tracer("test").Start(context.Background(), "GET /v1/leagues/{leagueId}")

			recordLeagueFailure(ctx, test.err)
			span.End()

			spans := recorder.Ended()
			if len(spans) != 1 {
				t.Fatalf("ended spans = %d, want 1", len(spans))
			}
			if got := leagueFailureReason(spans[0].Attributes()); got != test.want {
				t.Fatalf("failure reason = %q, want %q", got, test.want)
			}
			for _, attribute := range spans[0].Attributes() {
				if attribute.Key == "tournaments_manager.failure.reason" && attribute.Value.AsString() != test.want {
					t.Fatalf("unexpected failure attribute %s", attribute.Value.AsString())
				}
			}
		})
	}
}

func TestLeagueHandlersRecordValidationAndBusinessFailuresOnRootSpan(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		request *http.Request
		want    string
	}{
		{
			name:    "public lookup validation",
			handler: getPublicLeague(leagues.NewCreationService(testCreationRepository{})),
			request: httptest.NewRequest(http.MethodGet, "/v1/leagues/not-a-uuid", nil),
			want:    "validation.rejected",
		},
		{
			name:    "follow invisible league",
			handler: followLeague(leagues.NewService(testLeagueRepository{})),
			request: leaguePathRequest(http.MethodPut, "/v1/me/leagues/019abcde-2222-7222-8222-222222222222/follow", "019abcde-2222-7222-8222-222222222222"),
			want:    "league.not_found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			ctx, span := provider.Tracer("test").Start(test.request.Context(), "HTTP root")
			if test.name == "follow invisible league" {
				ctx = context.WithValue(ctx, accountContextKey{}, "019abcde-1111-7111-8111-111111111111")
			}
			test.handler.ServeHTTP(httptest.NewRecorder(), test.request.WithContext(ctx))
			span.End()

			spans := recorder.Ended()
			if got := leagueFailureReason(spans[0].Attributes()); got != test.want {
				t.Fatalf("failure reason = %q, want %q", got, test.want)
			}
		})
	}
}

func leaguePathRequest(method, target, leagueID string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	request.SetPathValue("leagueId", leagueID)
	return request
}

func leagueFailureReason(attributes []attribute.KeyValue) string {
	for _, attribute := range attributes {
		if attribute.Key == "tournaments_manager.failure.reason" {
			return attribute.Value.AsString()
		}
	}
	return ""
}
