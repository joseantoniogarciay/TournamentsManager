package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/access"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// AccountTournamentRepository shares the PostgreSQL pool and generated queries
// used by the capability-focused adapter files in this package.
type AccountTournamentRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewAccountTournamentRepository builds the account-relationship adapter.
func NewAccountTournamentRepository(pool *pgxpool.Pool) AccountTournamentRepository {
	return AccountTournamentRepository{pool: pool, queries: sqlc.New(pool)}
}

// PurgeExpired removes a batch of accounts whose deletion window has expired.
// The CTE locks only selected accounts, and SKIP LOCKED avoids waiting for a
// concurrent transaction on one of them.

func optionalUUID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{}, nil
	}
	return uuidValue(value)
}

func uuidValue(value string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}
	return uuid, nil
}

func uuidString(value pgtype.UUID) string { return value.String() }

func nullableText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func timestamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func administeredItems(rows []sqlc.ListAdministeredTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func followedItems(rows []sqlc.ListFollowedTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

func recentItems(rows []sqlc.ListRecentAccountTournamentsRow) []tournaments.Item {
	items := make([]tournaments.Item, len(rows))
	for index, row := range rows {
		items[index] = tournaments.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), LastActivityAt: timestamp(row.LastActivityAt.Time), Relationship: row.Relationship}
	}
	return items
}

var _ access.Repository = AccountTournamentRepository{}
