package postgres

import (
	"context"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
)

func TestIntegrationSuggestionPersistsAccountBodyAndTimestamp(t *testing.T) {
	ctx := context.Background()
	pool := integrationPool(t)
	accountID := createVerifiedLocalAccount(t, ctx, pool, "suggestion@example.test", "suggestion_person", "correct password")

	result, err := suggestions.NewService(NewSuggestionRepository(pool), nil).Submit(ctx, accountID, "  Añadir torneos por parejas  ")
	if err != nil || result.NotificationFailed {
		t.Fatalf("Submit() = %#v, %v", result, err)
	}

	var username, body string
	var hasCreatedAt bool
	if err := pool.QueryRow(ctx, `SELECT accounts.username, product_suggestions.body, product_suggestions.created_at IS NOT NULL FROM product_suggestions JOIN accounts ON accounts.id = product_suggestions.account_id WHERE product_suggestions.account_id = $1`, accountID).Scan(&username, &body, &hasCreatedAt); err != nil {
		t.Fatalf("leer sugerencia: %v", err)
	}
	if username != "suggestion_person" || body != "Añadir torneos por parejas" || !hasCreatedAt {
		t.Fatalf("suggestion = username %q, body %q, created %v", username, body, hasCreatedAt)
	}
}
