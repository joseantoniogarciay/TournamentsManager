// Package notifications contiene el buzón interno duradero de cada cuenta.
package notifications

import "context"

// Item es una notificación mostrable en el buzón de una cuenta.
type Item struct {
	ID, Kind, LeagueID, LeagueName, CreatedAt string
	ReadAt                                    *string
}

// Repository persiste el buzón interno por cuenta.
type Repository interface {
	ListNotifications(context.Context, string) ([]Item, error)
	UnreadCount(context.Context, string) (int, error)
	MarkAllRead(context.Context, string) error
	Delete(context.Context, string, string) error
	DeleteAll(context.Context, string) error
}

// Service coordina las operaciones del buzón interno.
type Service struct{ repository Repository }

// NewService construye el servicio sobre su puerto de persistencia.
func NewService(repository Repository) Service { return Service{repository: repository} }

// List devuelve el buzón de una cuenta.
func (s Service) List(ctx context.Context, accountID string) ([]Item, error) {
	return s.repository.ListNotifications(ctx, accountID)
}

// UnreadCount devuelve el número de avisos pendientes de lectura.
func (s Service) UnreadCount(ctx context.Context, accountID string) (int, error) {
	return s.repository.UnreadCount(ctx, accountID)
}

// MarkAllRead marca el buzón completo como leído.
func (s Service) MarkAllRead(ctx context.Context, accountID string) error {
	return s.repository.MarkAllRead(ctx, accountID)
}

// Delete elimina una notificación que pertenezca a la cuenta.
func (s Service) Delete(ctx context.Context, accountID, notificationID string) error {
	return s.repository.Delete(ctx, accountID, notificationID)
}

// DeleteAll elimina todo el buzón de una cuenta.
func (s Service) DeleteAll(ctx context.Context, accountID string) error {
	return s.repository.DeleteAll(ctx, accountID)
}
