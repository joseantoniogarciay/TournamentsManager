// Package suggestions stores private product feedback from authenticated accounts.
package suggestions

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	// MinBodyLength is the minimum number of Unicode characters accepted by the public contract.
	MinBodyLength = 8
	// MaxBodyLength is the maximum number of Unicode characters accepted by the public contract.
	MaxBodyLength = 1000
)

// ErrInvalidInput means the suggestion body does not satisfy the public contract.
var ErrInvalidInput = errors.New("invalid suggestion input")

// Item is the persisted suggestion passed to the notification boundary.
type Item struct {
	ID        string
	Username  string
	Body      string
	CreatedAt string
}

// Repository persists product suggestions independently from their delivery channel.
type Repository interface {
	CreateSuggestion(context.Context, string, string) (Item, error)
}

// Notifier alerts the product owner after the suggestion has been persisted.
type Notifier interface {
	NotifySuggestion(context.Context, Item) error
}

// SubmitResult separates the durable outcome from a secondary notification failure.
type SubmitResult struct{ NotificationFailed bool }

// Service coordinates durable submission and best-effort notification.
type Service struct {
	repository Repository
	notifier   Notifier
}

// NewService builds the suggestion use case.
func NewService(repository Repository, notifier Notifier) Service {
	return Service{repository: repository, notifier: notifier}
}

// NormalizeBody trims outer whitespace and checks the Unicode character limits.
func NormalizeBody(body string) (string, bool) {
	body = strings.TrimSpace(body)
	length := utf8.RuneCountInString(body)
	return body, length >= MinBodyLength && length <= MaxBodyLength
}

// Submit persists before notifying so SMTP can never become the source of truth.
func (s Service) Submit(ctx context.Context, accountID, body string) (SubmitResult, error) {
	body, valid := NormalizeBody(body)
	if !valid {
		return SubmitResult{}, ErrInvalidInput
	}
	item, err := s.repository.CreateSuggestion(ctx, accountID, body)
	if err != nil {
		return SubmitResult{}, err
	}
	if s.notifier == nil {
		return SubmitResult{}, nil
	}
	if err := s.notifier.NotifySuggestion(ctx, item); err != nil {
		return SubmitResult{NotificationFailed: true}, nil //nolint:nilerr // Notification is explicitly best-effort after durable persistence.
	}
	return SubmitResult{}, nil
}
