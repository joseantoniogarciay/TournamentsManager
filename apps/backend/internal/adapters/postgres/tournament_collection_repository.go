package postgres

import (
	"context"
	"fmt"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// List returns a page of tournaments related to an account.
func (r AccountTournamentRepository) List(ctx context.Context, accountID string, relationship tournaments.Relationship, cursor string, limit int) ([]tournaments.Item, error) {
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
	params := sqlc.ListAdministeredTournamentsParams{AccountID: accountUUID, CursorID: cursorUUID, PageSize: pageSize}
	if relationship == tournaments.Followed {
		rows, queryErr := r.queries.ListFollowedTournaments(ctx, sqlc.ListFollowedTournamentsParams(params))
		return followedItems(rows), queryErr
	}
	rows, err := r.queries.ListAdministeredTournaments(ctx, params)
	return administeredItems(rows), err
}

// ListRecent returns the fixed summary of relationships with recent activity.
func (r AccountTournamentRepository) ListRecent(ctx context.Context, accountID string) ([]tournaments.Item, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return nil, fmt.Errorf("convertir cuenta: %w", err)
	}
	rows, err := r.queries.ListRecentAccountTournaments(ctx, accountUUID)
	return recentItems(rows), err
}

// Follow creates the follow relationship when the league is visible.
func (r AccountTournamentRepository) Follow(ctx context.Context, accountID, leagueID string) (bool, error) {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return false, fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return false, fmt.Errorf("convertir liga: %w", err)
	}
	visible, err := r.queries.FollowVisibleTournament(ctx, sqlc.FollowVisibleTournamentParams{AccountID: accountUUID, TournamentID: leagueUUID})
	if err != nil {
		return false, fmt.Errorf("seguir liga: %w", err)
	}
	return visible, nil
}

// Unfollow idempotently removes the follow relationship.
func (r AccountTournamentRepository) Unfollow(ctx context.Context, accountID, leagueID string) error {
	accountUUID, err := uuidValue(accountID)
	if err != nil {
		return fmt.Errorf("convertir cuenta: %w", err)
	}
	leagueUUID, err := uuidValue(leagueID)
	if err != nil {
		return fmt.Errorf("convertir liga: %w", err)
	}
	if err := r.queries.UnfollowTournament(ctx, sqlc.UnfollowTournamentParams{AccountID: accountUUID, TournamentID: leagueUUID}); err != nil {
		return fmt.Errorf("dejar de seguir liga: %w", err)
	}
	return nil
}
