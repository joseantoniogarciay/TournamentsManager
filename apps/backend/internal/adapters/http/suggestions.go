package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
)

type suggestionRequest struct {
	Body string `json:"body"`
}

func submitSuggestion(service suggestions.Service, limiter *requestLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountID(r.Context())
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		var body suggestionRequest
		if err := decodeBody(r, &body); err != nil {
			recordSuggestionFailure(r.Context(), suggestions.ErrInvalidInput)
			writeValidationProblem(w)
			return
		}
		normalized, valid := suggestions.NormalizeBody(body.Body)
		if !valid {
			recordSuggestionFailure(r.Context(), suggestions.ErrInvalidInput)
			writeValidationProblem(w)
			return
		}
		if allowed, retryAfter := limiter.allow(accountID); !allowed {
			observability.RecordEndpointFailure(r.Context(), "rate_limit.exceeded")
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(w, http.StatusTooManyRequests, "Too many suggestions")
			return
		}
		result, err := service.Submit(r.Context(), accountID, normalized)
		recordSuggestionFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not save suggestion")
			return
		}
		if result.NotificationFailed {
			slog.Warn("no se pudo enviar el aviso de una sugerencia ya guardada")
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// recordSuggestionFailure never exports the account, body or generated identifier.
func recordSuggestionFailure(ctx context.Context, err error) {
	switch {
	case err == nil:
	case errors.Is(err, suggestions.ErrInvalidInput):
		observability.RecordEndpointFailure(ctx, "validation.rejected")
	default:
		observability.RecordDatabaseEndpointFailure(ctx, err)
	}
}
