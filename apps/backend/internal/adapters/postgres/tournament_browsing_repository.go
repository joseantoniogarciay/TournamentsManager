package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// PersonalMatches projects linked team encounters in a single database statement.
func (r AccountTournamentRepository) PersonalMatches(ctx context.Context, account, view, cursor string, limit int) (tournaments.PersonalMatchPage, error) {
	if limit < 1 || limit > 51 {
		return tournaments.PersonalMatchPage{}, fmt.Errorf("invalid page limit")
	}
	accountID, err := uuidValue(account)
	if err != nil {
		return tournaments.PersonalMatchPage{}, err
	}
	cursorID, err := optionalUUID(cursor)
	if err != nil {
		return tournaments.PersonalMatchPage{}, err
	}
	payload, err := r.queries.ListPersonalMatches(ctx, sqlc.ListPersonalMatchesParams{AccountID: accountID, View: view, CursorID: cursorID, PageSize: int32(limit)})
	if err != nil {
		return tournaments.PersonalMatchPage{}, err
	}
	var page tournaments.PersonalMatchPage
	err = json.Unmarshal(payload, &page)
	return page, err
}

// Editions resolves ancestors and descendants without guessing relationships by name.
func (r AccountTournamentRepository) Editions(ctx context.Context, id, cursor string, limit int) ([]tournaments.Edition, error) {
	if limit < 1 || limit > 51 {
		return nil, fmt.Errorf("invalid page limit")
	}
	tournamentID, err := uuidValue(id)
	if err != nil {
		return nil, err
	}
	cursorID, err := optionalUUID(cursor)
	if err != nil {
		return nil, err
	}
	payload, err := r.queries.ListTournamentEditions(ctx, sqlc.ListTournamentEditionsParams{TournamentID: tournamentID, CursorID: cursorID, PageSize: int32(limit)})
	if err != nil {
		return nil, err
	}
	var result struct {
		Exists bool                  `json:"exists"`
		Items  []tournaments.Edition `json:"items"`
	}
	if err = json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	if !result.Exists {
		return nil, tournaments.ErrTournamentNotFound
	}
	return result.Items, nil
}
