package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
)

const maxRequestBodyBytes = 64 * 1024
const requestBudget = 15 * time.Second

// passwordWorkAdmission bounds complete requests, including hashes that cannot
// be interrupted once running. Saturated work never waits in an unbounded queue.
type passwordWorkAdmission struct{ slots chan struct{} }

func (g passwordWorkAdmission) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Err() != nil {
			return
		}
		select {
		case g.slots <- struct{}{}:
			defer func() { <-g.slots }()
			next(w, r)
		default:
			rejectResourceLimit(w, r, 1)
		}
	}
}

func rejectResourceLimit(w http.ResponseWriter, r *http.Request, retry int) {
	observability.RecordEndpointFailure(r.Context(), "rate_limit.exceeded")
	w.Header().Set("Retry-After", strconv.Itoa(retry))
	writeProblem(w, http.StatusTooManyRequests, "Too many requests")
}

// Account keys come only from the authenticated context, never client inputs.
func resourceRateLimit(next http.HandlerFunc, limiter *requestLimiter, key func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Err() != nil {
			return
		}
		if allowed, retry := limiter.allow(key(r)); !allowed {
			rejectResourceLimit(w, r, retry)
			return
		}
		next(w, r)
	}
}

func accountResourceKey(r *http.Request) string {
	account, _ := currentAccountID(r.Context())
	return account
}

func withResourceBudget(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), requestBudget)
	return r.WithContext(ctx), cancel
}
