package postgres

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/sessions"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// AuthenticateWithClient validates access and occasionally touches advisory metadata.
func (r AccountTournamentRepository) AuthenticateWithClient(ctx context.Context, token string, client sessions.Client) (string, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	id, err := r.queries.AuthenticateSessionWithClient(ctx, sqlc.AuthenticateSessionWithClientParams{TokenHash: hash[:], ClientPlatform: client.Platform, ClientKind: client.Kind})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", tournaments.ErrUnauthenticated
	}
	if err != nil {
		return "", err
	}
	return uuidString(id), nil
}

// ListSessions exposes only the safe projection of active families belonging to the caller.
func (r AccountTournamentRepository) ListSessions(ctx context.Context, token, cursor string) (sessions.Page, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	var cursorID pgtype.UUID
	if cursor != "" {
		var err error
		cursorID, err = uuidValue(cursor)
		if err != nil {
			return sessions.Page{}, err
		}
	}
	rows, err := r.queries.ListSessions(ctx, sqlc.ListSessionsParams{TokenHash: hash[:], CursorID: cursorID})
	if err != nil {
		return sessions.Page{}, err
	}
	page := sessions.Page{Items: make([]sessions.Item, 0, min(len(rows), 50))}
	if len(rows) > 50 {
		cursor := uuidString(rows[49].ID)
		page.NextCursor = &cursor
		rows = rows[:50]
	}
	for _, row := range rows {
		page.Items = append(page.Items, sessions.Item{ID: uuidString(row.ID), CreatedAt: timestamp(row.CreatedAt.Time), LastSeenAt: timestamp(row.LastSeenAt.Time), Platform: row.ClientPlatform, Client: row.ClientKind, Current: row.Current})
	}
	return page, nil
}

// RevokeOtherSessions atomically revokes one other family or all other families.
func (r AccountTournamentRepository) RevokeOtherSessions(ctx context.Context, token, target string) error {
	hash := sha256.Sum256([]byte("session:" + token))
	var targetID pgtype.UUID
	if target != "" {
		var err error
		targetID, err = uuidValue(target)
		if err != nil {
			return err
		}
	}
	_, err := r.queries.RevokeOtherSessions(ctx, sqlc.RevokeOtherSessionsParams{TokenHash: hash[:], TargetID: targetID})
	if errors.Is(err, pgx.ErrNoRows) {
		return tournaments.ErrUnauthenticated
	}
	return err
}
