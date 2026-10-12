package observability

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDatabaseBudgetErrorsUseSafeTimeoutCategory(t *testing.T) {
	for _, code := range []string{"57014", "55P03"} {
		if reason := databaseFailureReason(&pgconn.PgError{Code: code, Message: "private SQL and input"}); reason != "request.timeout" {
			t.Fatalf("code %s: %s", code, reason)
		}
	}
}
