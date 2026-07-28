// Package leagues contiene la consulta de colecciones de ligas relacionadas.
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
	// ErrInvalidRelationship indica un filtro de relación no permitido.
	ErrInvalidRelationship = errors.New("relación de liga inválida")
	// ErrInvalidPage indica parámetros de paginación no permitidos.
	ErrInvalidPage = errors.New("página de ligas inválida")
	// ErrUnauthenticated indica que no hay una sesión válida.
	ErrUnauthenticated = errors.New("sesión no autenticada")
	// ErrLeagueNotFound indica una liga inexistente o no visible.
	ErrLeagueNotFound = errors.New("liga no disponible")
)

// Relationship identifica la perspectiva de una cuenta sobre una liga.
type Relationship string

// Relaciones admitidas por la colección de cuenta.
const (
	Administered Relationship = "administered"
	Followed     Relationship = "followed"
)

// Item es la proyección compacta de una liga relacionada.
type Item struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	State        string `json:"state"`
	CreatedAt    string `json:"createdAt"`
	Relationship string `json:"relationship"`
}

// Page contiene una página de elementos y su cursor opcional.
type Page struct {
	Items      []Item `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Repository persiste y consulta relaciones de ligas.
type Repository interface {
	List(context.Context, string, Relationship, string, int) ([]Item, error)
	Follow(context.Context, string, string) (bool, error)
	Unfollow(context.Context, string, string) error
}

// Follow guarda una liga visible para una cuenta.
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

// Unfollow retira una liga guardada para una cuenta.
func (s Service) Unfollow(ctx context.Context, accountID, leagueID string) error {
	return s.repository.Unfollow(ctx, accountID, leagueID)
}

// Service coordina las colecciones y relaciones de ligas.
type Service struct{ repository Repository }

// NewService construye el servicio con su puerto de persistencia.
func NewService(repository Repository) Service { return Service{repository: repository} }

// List devuelve una página de ligas de la relación indicada.
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
