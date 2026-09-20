package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// AssignAdministrator grants tournament administration to an account.
func (r AccountTournamentRepository) AssignAdministrator(ctx context.Context, accountID, leagueID, username string) error {
	account, err := uuidValue(accountID)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var organizer string
	if err := tx.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if organizer != account.String() {
		return tournaments.ErrTournamentForbidden
	}
	var administrator string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM accounts WHERE username = $1 AND state = 'verified'`, username).Scan(&administrator); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if administrator == organizer {
		return tournaments.ErrTournamentAdministratorConflict
	}
	result, err := tx.Exec(ctx, `INSERT INTO tournament_administrators (tournament_id, account_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, leagueID, administrator)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `INSERT INTO account_notifications (account_id, kind, tournament_id) VALUES ($1, 'tournament_administrator_assigned', $2)`, administrator, leagueID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListNotifications returns the account's internal inbox.
