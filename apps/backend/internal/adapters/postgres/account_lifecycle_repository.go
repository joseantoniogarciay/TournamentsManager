package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/accounts"
)

// PurgeExpired removes a bounded batch of accounts whose deletion window expired.
func (r AccountTournamentRepository) PurgeExpired(ctx context.Context, limit int) (int64, error) {
	if limit < 1 {
		return 0, fmt.Errorf("límite de purga debe ser positivo")
	}
	if _, err := r.pool.Exec(ctx, `DELETE FROM google_risc_events WHERE expires_at <= now()`); err != nil {
		return 0, err
	}
	command, err := r.pool.Exec(ctx, `
		WITH candidates AS (
			SELECT id
			FROM accounts
			WHERE state = 'deletion_pending'
				AND deletion_requested_at <= now() - interval '30 days'
				AND NOT EXISTS (
					SELECT 1
					FROM tournaments
					WHERE organizer_account_id = accounts.id
				)
			ORDER BY deletion_requested_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		, blocked_legal_evidence AS (
			UPDATE legal_account_acceptances
			SET retention_until = now() + interval '5 years', updated_at = now()
			WHERE account_id IN (SELECT id FROM candidates)
		)
		, expired_legal_evidence AS (
			DELETE FROM legal_account_acceptances
			WHERE retention_until IS NOT NULL
				AND retention_until <= now()
		)
		DELETE FROM accounts
		USING candidates
		WHERE accounts.id = candidates.id`, limit)
	if err != nil {
		return 0, err
	}
	return command.RowsAffected(), nil
}

var _ accounts.PurgeRepository = AccountTournamentRepository{}

// ScheduleAccountDeletion revokes access and personal relationships without deleting the account.
func (r AccountTournamentRepository) ScheduleAccountDeletion(ctx context.Context, accountID string) (time.Time, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var hasOwned bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tournaments WHERE organizer_account_id = $1)`, accountID).Scan(&hasOwned); err != nil {
		return time.Time{}, err
	}
	if hasOwned {
		return time.Time{}, accounts.ErrAccountHasOwnedTournaments
	}
	var requested time.Time
	if err := tx.QueryRow(ctx, `UPDATE accounts SET state = 'deletion_pending', deletion_requested_at = now() WHERE id = $1 AND state = 'verified' RETURNING deletion_requested_at`, accountID).Scan(&requested); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_followers WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_administrators WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tournament_team_accounts WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE account_id = $1`, accountID); err != nil {
		return time.Time{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return time.Time{}, err
	}
	return requested.AddDate(0, 0, 30).UTC(), nil
}

// Create creates a published league without a schedule yet.
