package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// ListAdministrators returns the usernames that administer a tournament.
func (r AccountTournamentRepository) ListAdministrators(ctx context.Context, accountID, leagueID string) ([]string, error) {
	account, err := uuidValue(accountID)
	if err != nil {
		return nil, err
	}
	var organizer string
	if err := r.pool.QueryRow(ctx, `SELECT organizer_account_id::text FROM tournaments WHERE id = $1`, leagueID).Scan(&organizer); errors.Is(err, pgx.ErrNoRows) {
		return nil, tournaments.ErrTournamentNotFound
	} else if err != nil {
		return nil, err
	}
	if organizer != account.String() {
		return nil, tournaments.ErrTournamentForbidden
	}
	rows, err := r.pool.Query(ctx, `SELECT accounts.username FROM tournament_administrators JOIN accounts ON accounts.id = tournament_administrators.account_id WHERE tournament_administrators.tournament_id = $1 ORDER BY accounts.username`, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	administrators := []string{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		administrators = append(administrators, username)
	}
	return administrators, rows.Err()
}

// RemoveAdministrator removes a delegated administrator exclusively for the league owner.
func (r AccountTournamentRepository) RemoveAdministrator(ctx context.Context, accountID, leagueID, username string) error {
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
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE tournament_id = $1 AND account_id = (SELECT id FROM accounts WHERE username = $2)`, leagueID, username); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// TransferOwnership atomically replaces the organizer while preserving the competition.
func (r AccountTournamentRepository) TransferOwnership(ctx context.Context, accountID, leagueID, username string) error {
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
	var recipient string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM accounts WHERE username = $1 AND state = 'verified'`, username).Scan(&recipient); errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrTournamentNotFound
	} else if err != nil {
		return err
	}
	if recipient == organizer {
		return tournaments.ErrTournamentOwnershipTransferConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE tournaments SET organizer_account_id = $2, last_activity_at = now() WHERE id = $1`, leagueID, recipient); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE tournament_id = $1 AND account_id = $2`, leagueID, recipient); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO account_notifications (account_id, kind, tournament_id) VALUES ($1, 'tournament_ownership_transferred', $2)`, recipient, leagueID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
