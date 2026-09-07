package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRecordTournamentFailureUsesClosedSafeReasons(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"not found", tournaments.ErrTournamentNotFound, "tournament.not_found"},
		{"forbidden", tournaments.ErrTournamentForbidden, "tournament.forbidden"},
		{"start conflict", tournaments.ErrTournamentConflict, "tournament.start_conflict"},
		{"team conflict", tournaments.ErrTournamentTeamConflict, "tournament.team_conflict"},
		{"withdrawal conflict", tournaments.ErrTournamentWithdrawalConflict, "tournament.withdrawal_conflict"},
		{"result conflict", tournaments.ErrMatchResultConflict, "tournament.result_conflict"},
		{"administrator conflict", tournaments.ErrTournamentAdministratorConflict, "tournament.administrator_conflict"},
		{"ownership conflict", tournaments.ErrTournamentOwnershipTransferConflict, "tournament.ownership_transfer_conflict"},
		{"cancellation conflict", tournaments.ErrTournamentCancellationConflict, "tournament.cancellation_conflict"},
		{"completion conflict", tournaments.ErrTournamentCompletionConflict, "tournament.completion_conflict"},
		{"invalid input", tournaments.ErrInvalidTournamentInput, "validation.rejected"},
		{"invalid bracket result", tournaments.ErrInvalidBracketResult, "validation.rejected"},
		{"unresolved bracket", tournaments.ErrBracketMatchNotReady, "bracket.match_not_ready"},
		{"dependent bracket result", tournaments.ErrBracketResultDependency, "bracket.result_dependency"},
		{"cancelled", context.Canceled, "request.cancelled"},
		{"database fallback", errors.New("postgres: account 019abcde-2222-7222-8222-222222222222 failed"), "database.query_failed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			ctx, span := provider.Tracer("test").Start(context.Background(), "GET /v1/tournaments/{tournamentId}")

			recordTournamentFailure(ctx, test.err)
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

func TestTournamentHandlersRecordValidationAndBusinessFailuresOnRootSpan(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		request *http.Request
		want    string
	}{
		{
			name:    "public lookup validation",
			handler: getPublicTournament(tournaments.NewCreationService(testCreationRepository{})),
			request: httptest.NewRequest(http.MethodGet, "/v1/tournaments/not-a-uuid", nil),
			want:    "validation.rejected",
		},
		{
			name:    "follow invisible league",
			handler: followTournament(tournaments.NewService(testTournamentRepository{})),
			request: leaguePathRequest(http.MethodPut, "/v1/me/tournaments/019abcde-2222-7222-8222-222222222222/follow", "019abcde-2222-7222-8222-222222222222"),
			want:    "tournament.not_found",
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
	request.SetPathValue("tournamentId", leagueID)
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
