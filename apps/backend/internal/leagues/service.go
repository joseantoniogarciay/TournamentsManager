// Package leagues contains related league collection queries.
package leagues

import (
	"context"
	"errors"
)

const (
	defaultLimit = 20
	maximumLimit = 50
)

var (
	// ErrInvalidRelationship indicates a disallowed relationship filter.
	ErrInvalidRelationship = errors.New("relación de liga inválida")
	// ErrInvalidPage indicates invalid pagination parameters.
	ErrInvalidPage = errors.New("página de ligas inválida")
	// ErrUnauthenticated indicates that no valid session exists.
	ErrUnauthenticated = errors.New("sesión no autenticada")
	// ErrLeagueNotFound indicates a nonexistent or invisible league.
	ErrLeagueNotFound = errors.New("liga no disponible")
)

// Relationship identifies an account's relationship with a league.
type Relationship string

// Relationships supported by the account collection.
const (
	Administered Relationship = "administered"
	Followed     Relationship = "followed"
)

// Item is the compact projection of a related league.
type Item struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	State          string `json:"state"`
	CreatedAt      string `json:"createdAt"`
	LastActivityAt string `json:"lastActivityAt"`
	Relationship   string `json:"relationship"`
}

// Page contains an item page and its optional cursor.
type Page struct {
	Items      []Item `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Repository persists and queries league relationships.
type Repository interface {
	List(context.Context, string, Relationship, string, int) ([]Item, error)
	ListRecent(context.Context, string) ([]Item, error)
	Follow(context.Context, string, string) (bool, error)
	Unfollow(context.Context, string, string) error
}

// ListRecent returns up to five related leagues with the most recent activity.
func (s Service) ListRecent(ctx context.Context, accountID string) ([]Item, error) {
	return s.repository.ListRecent(ctx, accountID)
}

// Follow saves a visible league for an account.
func (s Service) Follow(ctx context.Context, accountID, leagueID string) error {
	visible, err := s.repository.Follow(ctx, accountID, leagueID)
	if err != nil {
		return err
	}
	if !visible {
		return ErrLeagueNotFound
	}
	return nil
}

// Unfollow removes a saved league from an account.
func (s Service) Unfollow(ctx context.Context, accountID, leagueID string) error {
	return s.repository.Unfollow(ctx, accountID, leagueID)
}

// Service coordinates league collections and relationships.
type Service struct{ repository Repository }

// NewService builds the service with its persistence port.
func NewService(repository Repository) Service { return Service{repository: repository} }

// List returns a league page for the specified relationship.
func (s Service) List(ctx context.Context, accountID string, relationship Relationship, cursor string, limit int) (Page, error) {
	if relationship != Administered && relationship != Followed {
		return Page{}, ErrInvalidRelationship
	}
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > maximumLimit {
		return Page{}, ErrInvalidPage
	}

	items, err := s.repository.List(ctx, accountID, relationship, cursor, limit+1)
	if err != nil {
		return Page{}, err
	}
	page := Page{Items: items}
	if len(page.Items) > limit {
		page.NextCursor = page.Items[limit].ID
		page.Items = page.Items[:limit]
	}
	return page, nil
}
