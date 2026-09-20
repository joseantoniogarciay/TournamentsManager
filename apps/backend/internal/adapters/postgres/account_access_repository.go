package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// Authenticate resolves an active opaque session token.
func (r AccountTournamentRepository) Authenticate(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	accountID, err := r.queries.FindAuthenticatedAccountID(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", tournaments.ErrUnauthenticated
		}
		return "", fmt.Errorf("buscar sesión: %w", err)
	}
	return uuidString(accountID), nil
}

// GetCurrentSession returns the identity and validity of the presented session.
func (r AccountTournamentRepository) GetCurrentSession(ctx context.Context, token string) (tournaments.CurrentSession, error) {
	hash := sha256.Sum256([]byte("session:" + token))
	row, err := r.queries.GetCurrentSession(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tournaments.CurrentSession{}, tournaments.ErrUnauthenticated
		}
		return tournaments.CurrentSession{}, fmt.Errorf("consultar sesión actual: %w", err)
	}
	return tournaments.CurrentSession{
		AccountID:         uuidString(row.ID),
		Username:          row.Username,
		LastTeamName:      nullableText(row.LastTeamName),
		IdleExpiresAt:     timestamp(row.IdleExpiresAt.Time),
		AbsoluteExpiresAt: timestamp(row.AbsoluteExpiresAt.Time),
	}, nil
}

// RevokeSession idempotently revokes the presented session and its refresh tokens.
func (r AccountTournamentRepository) RevokeSession(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte("session:" + token))
	_, err := r.queries.RevokeSession(ctx, hash[:])
	return err
}

// GetAccessMethods returns the access methods configured for an account.
func (r AccountTournamentRepository) GetAccessMethods(ctx context.Context, accountID string) (tournaments.AccessMethods, error) {
	id, err := uuidValue(accountID)
	if err != nil {
		return tournaments.AccessMethods{}, err
	}
	row, err := r.queries.GetAccessMethods(ctx, id)
	if err != nil {
		return tournaments.AccessMethods{}, err
	}
	return tournaments.AccessMethods{Email: row.Email, Username: row.Username, HasPassword: row.HasPassword, HasGoogle: row.HasGoogle}, nil
}

// CurrentPasswordHash gets the verifier associated with an active session.
func (r AccountTournamentRepository) CurrentPasswordHash(ctx context.Context, sessionToken string) (string, error) {
	hash := sha256.Sum256([]byte("session:" + sessionToken))
	return r.queries.GetCurrentPasswordHash(ctx, hash[:])
}

// CreateReauthenticationTicket stores a reauthentication ticket for a session.
func (r AccountTournamentRepository) CreateReauthenticationTicket(ctx context.Context, sessionToken string, ticketHash []byte) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	_, err := r.queries.CreateReauthenticationTicket(ctx, sqlc.CreateReauthenticationTicketParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	return err
}

// ConsumeReauthenticationTicketAndSetPassword consumes the ticket and changes the password.
func (r AccountTournamentRepository) ConsumeReauthenticationTicketAndSetPassword(ctx context.Context, sessionToken string, ticketHash []byte, passwordHash string) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	rows, err := r.queries.ConsumeReauthenticationTicketAndSetPassword(ctx, sqlc.ConsumeReauthenticationTicketAndSetPasswordParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash, PasswordHash: passwordHash})
	if err != nil {
		return err
	}
	if rows != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

// ConsumeReauthenticationTicketAndRemovePassword removes the local method only
// when Google remains available for the same account.
func (r AccountTournamentRepository) ConsumeReauthenticationTicketAndRemovePassword(ctx context.Context, sessionToken string, ticketHash []byte) error {
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	rows, err := r.queries.ConsumeReauthenticationTicketAndRemovePassword(ctx, sqlc.ConsumeReauthenticationTicketAndRemovePasswordParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	if err != nil {
		return err
	}
	if rows != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

// List returns the requested page of league relationships.
