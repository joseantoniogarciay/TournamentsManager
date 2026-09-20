package postgres

import (
	"context"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
)

// ListNotifications returns the durable inbox for an account.
func (r AccountTournamentRepository) ListNotifications(ctx context.Context, accountID string) ([]notifications.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT account_notifications.id::text, account_notifications.kind, tournaments.id::text, tournaments.name, account_notifications.created_at, account_notifications.read_at FROM account_notifications JOIN tournaments ON tournaments.id = account_notifications.tournament_id WHERE account_notifications.account_id = $1 ORDER BY account_notifications.created_at DESC, account_notifications.id DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []notifications.Item{}
	for rows.Next() {
		var item notifications.Item
		var createdAt time.Time
		var readAt *time.Time
		if err := rows.Scan(&item.ID, &item.Kind, &item.TournamentID, &item.TournamentName, &createdAt, &readAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(time.RFC3339Nano)
		if readAt != nil {
			formattedReadAt := readAt.Format(time.RFC3339Nano)
			item.ReadAt = &formattedReadAt
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UnreadCount counts the account's unread notifications.
func (r AccountTournamentRepository) UnreadCount(ctx context.Context, accountID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM account_notifications WHERE account_id = $1 AND read_at IS NULL`, accountID).Scan(&count)
	return count, err
}

// MarkAllRead marks the account's pending notifications as read.
func (r AccountTournamentRepository) MarkAllRead(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE account_notifications SET read_at = now() WHERE account_id = $1 AND read_at IS NULL`, accountID)
	return err
}

// Delete removes a notification owned by the account.
func (r AccountTournamentRepository) Delete(ctx context.Context, accountID, notificationID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE id = $1 AND account_id = $2`, notificationID, accountID)
	return err
}

// DeleteAll removes all account notifications.
func (r AccountTournamentRepository) DeleteAll(ctx context.Context, accountID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM account_notifications WHERE account_id = $1`, accountID)
	return err
}

// ListAdministrators returns delegated administrators exclusively to the league owner.
