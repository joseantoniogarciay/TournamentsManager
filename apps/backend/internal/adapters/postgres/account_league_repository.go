package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
)

// AccountLeagueRepository persiste sesiones y relaciones de ligas.
type AccountLeagueRepository struct{ queries *sqlc.Queries }

// NewAccountLeagueRepository construye el adaptador de relaciones de cuenta.
func NewAccountLeagueRepository(pool *pgxpool.Pool) AccountLeagueRepository {
	return AccountLeagueRepository{queries: sqlc.New(pool)}
}

// Authenticate resuelve una sesión opaca válida en su cuenta.
func (r AccountLeagueRepository) Authenticate(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	accountID, err := r.queries.FindAuthenticatedAccountID(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", leagues.ErrUnauthenticated
		}
		return "", fmt.Errorf("buscar sesión: %w", err)
	}
	return uuidString(accountID), nil
}

// List devuelve la página solicitada de relaciones con ligas.
func (r AccountLeagueRepository) List(ctx context.Context, accountID string, relationship leagues.Relationship, cursor string, limit int) ([]leagues.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return nil, fmt.Errorf("convertir cuenta: %w", err)
	}
	cursorUUID, err := optionalUUID(cursor)
	if err != nil {
		return nil, fmt.Errorf("convertir cursor: %w", err)
	}
	if limit < 1 || limit > 51 {
		return nil, fmt.Errorf("límite inválido")
	}
	pageSize := int32(limit)
	params := sqlc.ListAdministeredLeaguesParams{AccountID: accountUUID, CursorID: cursorUUID, PageSize: pageSize}
	if relationship == leagues.Followed {
		rows, queryErr := r.queries.ListFollowedLeagues(ctx, sqlc.ListFollowedLeaguesParams(params))
		return followedItems(rows), queryErr
	}
	rows, err := r.queries.ListAdministeredLeagues(ctx, params)
	return administeredItems(rows), err
}

// Follow crea la relación de seguimiento cuando la liga es visible.
func (r AccountLeagueRepository) Follow(ctx context.Context, accountID, leagueID string) (bool, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return false, fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return false, fmt.Errorf("convertir liga: %w", err)
	}
	visible, err := r.queries.FollowVisibleLeague(ctx, sqlc.FollowVisibleLeagueParams{AccountID: accountUUID, LeagueID: leagueUUID})
	if err != nil {
		return false, fmt.Errorf("seguir liga: %w", err)
	}
	return visible, nil
}

// Unfollow elimina la relación de seguimiento de forma idempotente.
func (r AccountLeagueRepository) Unfollow(ctx context.Context, accountID, leagueID string) error {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return fmt.Errorf("convertir liga: %w", err)
	}
	if err := r.queries.UnfollowLeague(ctx, sqlc.UnfollowLeagueParams{AccountID: accountUUID, LeagueID: leagueUUID}); err != nil {
		return fmt.Errorf("dejar de seguir liga: %w", err)
	}
	return nil
}

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

func timestamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func administeredItems(rows []sqlc.ListAdministeredLeaguesRow) []leagues.Item {
	items := make([]leagues.Item, len(rows))
	for index, row := range rows {
		items[index] = leagues.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), Relationship: row.Relationship}
	}
	return items
}

func followedItems(rows []sqlc.ListFollowedLeaguesRow) []leagues.Item {
	items := make([]leagues.Item, len(rows))
	for index, row := range rows {
		items[index] = leagues.Item{ID: uuidString(row.ID), Name: row.Name, State: row.State, CreatedAt: timestamp(row.CreatedAt.Time), Relationship: row.Relationship}
	}
	return items
}
