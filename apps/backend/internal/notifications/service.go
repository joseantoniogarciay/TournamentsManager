// Package notifications contains each account's durable internal inbox.
package notifications

import "context"

// Item is a notification that can be displayed in an account inbox.
type Item struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	LeagueID   string  `json:"leagueId"`
	LeagueName string  `json:"leagueName"`
	CreatedAt  string  `json:"createdAt"`
	ReadAt     *string `json:"readAt"`
}

// Repository persists each account's internal inbox.
type Repository interface {
	ListNotifications(context.Context, string) ([]Item, error)
	UnreadCount(context.Context, string) (int, error)
	MarkAllRead(context.Context, string) error
	Delete(context.Context, string, string) error
	DeleteAll(context.Context, string) error
}

// Service coordinates internal inbox operations.
type Service struct{ repository Repository }

// NewService builds the service over its persistence port.
func NewService(repository Repository) Service { return Service{repository: repository} }

// List returns an account inbox.
func (s Service) List(ctx context.Context, accountID string) ([]Item, error) {
	return s.repository.ListNotifications(ctx, accountID)
}

// UnreadCount returns the number of unread notifications.
func (s Service) UnreadCount(ctx context.Context, accountID string) (int, error) {
	return s.repository.UnreadCount(ctx, accountID)
}

// MarkAllRead marks the whole inbox as read.
func (s Service) MarkAllRead(ctx context.Context, accountID string) error {
	return s.repository.MarkAllRead(ctx, accountID)
}

// Delete removes a notification owned by the account.
func (s Service) Delete(ctx context.Context, accountID, notificationID string) error {
	return s.repository.Delete(ctx, accountID, notificationID)
}

// DeleteAll removes every notification from an account inbox.
func (s Service) DeleteAll(ctx context.Context, accountID string) error {
	return s.repository.DeleteAll(ctx, accountID)
}
