package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
)

// SuggestionRepository persists private product suggestions.
type SuggestionRepository struct{ queries *sqlc.Queries }

// NewSuggestionRepository builds the PostgreSQL adapter.
func NewSuggestionRepository(pool *pgxpool.Pool) SuggestionRepository {
	return SuggestionRepository{queries: sqlc.New(pool)}
}

// CreateSuggestion stores the trimmed body and returns only the data needed by the notifier.
func (r SuggestionRepository) CreateSuggestion(ctx context.Context, accountID, body string) (suggestions.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return suggestions.Item{}, err
	}
	row, err := r.queries.CreateProductSuggestion(ctx, sqlc.CreateProductSuggestionParams{AccountID: accountUUID, Body: body})
	if err != nil {
		return suggestions.Item{}, err
	}
	return suggestions.Item{ID: row.ID, Username: row.Username, Body: row.Body, CreatedAt: row.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")}, nil
}
