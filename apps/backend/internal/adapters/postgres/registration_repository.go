package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/legal"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
)

// RegistrationRepository implements local registration persistence with sqlc.
type RegistrationRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// VerifyAndCreateSession consumes a verification and issues its session atomically.
func (r RegistrationRepository) VerifyAndCreateSession(ctx context.Context, verificationHash, sessionHash, refreshHash, previousSessionHash []byte) (registration.Session, error) {
	row, err := r.queries.VerifyRegistrationAndCreateSession(ctx, sqlc.VerifyRegistrationAndCreateSessionParams{TokenHash: verificationHash, TokenHash_2: sessionHash, TokenHash_3: refreshHash, PreviousSessionHash: previousSessionHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrVerificationInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: row.ID.String(), Username: row.Username, LastTeamName: nullableText(row.LastTeamName), IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// RotateSessionTokens consumes a refresh token and atomically creates its successors.
func (r RegistrationRepository) RotateSessionTokens(ctx context.Context, refreshHash, sessionHash, nextRefreshHash []byte) (registration.Session, error) {
	row, err := r.queries.RotateSessionTokens(ctx, sqlc.RotateSessionTokensParams{TokenHash: refreshHash, TokenHash_2: sessionHash, TokenHash_3: nextRefreshHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrRefreshInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: row.ID.String(), Username: row.Username, LastTeamName: nullableText(row.LastTeamName), IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// NewRegistrationRepository connects the use-case port to the PostgreSQL pool.
func NewRegistrationRepository(pool *pgxpool.Pool) RegistrationRepository {
	return RegistrationRepository{pool: pool, queries: sqlc.New(pool)}
}

// IsUsernameAvailable checks the uniqueness constraint without creating a reservation.
func (r RegistrationRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	return r.queries.IsUsernameAvailable(ctx, username)
}

// SearchUsernames searches public matches without exposing other account data.
func (r RegistrationRepository) SearchUsernames(ctx context.Context, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `-- name: SearchPublicUsernames :many
		SELECT username FROM accounts WHERE state = 'verified' AND username LIKE '%' || $1 || '%' ORDER BY username LIMIT 20`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	usernames := []string{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, rows.Err()
}

// FindLocalAccountForLogin gets the local credential and email-verification state.
func (r RegistrationRepository) FindLocalAccountForLogin(ctx context.Context, email string) (registration.LocalAccount, error) {
	row, err := r.queries.FindLocalAccountForLogin(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.LocalAccount{}, registration.ErrLoginInvalid
	}
	if err != nil {
		return registration.LocalAccount{}, err
	}
	return registration.LocalAccount{ID: row.ID.String(), Email: row.Email, Locale: registration.Locale(row.Locale), Username: row.Username, PasswordHash: row.PasswordHash, Verified: row.State == "verified"}, nil
}

// CreateLocalLoginSession persists an optional tournament and the hashed session tokens atomically.
func (r RegistrationRepository) CreateLocalLoginSession(ctx context.Context, accountID string, sessionHash, refreshHash []byte, draft *registration.Draft) (registration.Session, error) {
	id, err := parseUUID(accountID)
	if err != nil {
		return registration.Session{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return registration.Session{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	row, err := r.queries.WithTx(tx).CreateLocalLoginSession(ctx, sqlc.CreateLocalLoginSessionParams{ID: id, TokenHash: sessionHash, TokenHash_2: refreshHash})
	if err != nil {
		return registration.Session{}, err
	}
	if draft != nil {
		if err := createTournamentFromDraft(ctx, tx, accountID, draft.ID, draft.Name, draft.Sport, draft.Teams); err != nil {
			return registration.Session{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return registration.Session{}, err
	}
	lastTeamName := nullableText(row.LastTeamName)
	if draft != nil && len(draft.Teams) > 0 {
		lastTeamName = draft.Teams[0]
	}
	return registration.Session{AccountID: accountID, Username: row.Username, LastTeamName: lastTeamName, IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

func createTournamentFromDraft(ctx context.Context, tx pgx.Tx, accountID, draftID, name string, sport tournaments.Sport, teams []string) error {
	var tournamentID string
	err := tx.QueryRow(ctx, `INSERT INTO tournaments (organizer_account_id, source_draft_id, name, sport, state, published_at) VALUES ($1, $2, $3, $4, 'published', now()) ON CONFLICT (organizer_account_id, source_draft_id) DO NOTHING RETURNING id::text`, accountID, draftID, name, sport).Scan(&tournamentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var firstTeamID string
	for position, team := range teams {
		if position == 0 {
			if err := tx.QueryRow(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3) RETURNING id::text`, tournamentID, team, position+1).Scan(&firstTeamID); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tournament_teams (tournament_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3)`, tournamentID, team, position+1); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO tournament_team_accounts (tournament_id, team_id, account_id) VALUES ($1, $2, $3)`, tournamentID, firstTeamID, accountID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET last_team_name = $2 WHERE id = $1`, accountID, teams[0]); err != nil {
		return err
	}
	return nil
}

// RenewLoginVerification rotates the pending verification token and returns its recipient.
func (r RegistrationRepository) RenewLoginVerification(ctx context.Context, accountID string, tokenHash []byte) (string, registration.Locale, error) {
	id, err := parseUUID(accountID)
	if err != nil {
		return "", "", err
	}
	row, err := r.queries.RenewLoginVerification(ctx, sqlc.RenewLoginVerificationParams{ID: id, TokenHash: tokenHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", registration.ErrLoginInvalid
	}
	if err != nil {
		return "", "", err
	}
	return row.Email, registration.Locale(row.Locale), nil
}

// CreatePending creates the three identity records in a single statement.
func (r RegistrationRepository) CreatePending(ctx context.Context, input registration.Input, passwordHash string, tokenHash []byte) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var accountID string
	err = tx.QueryRow(ctx, `-- name: CreatePendingAccount :one
		INSERT INTO accounts (email, locale, state, username, expires_at)
		VALUES ($1, $2, 'pending_verification', $3, now() + interval '7 days')
		ON CONFLICT DO NOTHING
		RETURNING id::text`, input.Email, string(input.Locale), input.Username).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `-- name: CreatePendingCredential :exec
		INSERT INTO local_credentials (account_id, password_hash) VALUES ($1, $2)`, accountID, passwordHash); err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `-- name: CreatePendingVerification :exec
		INSERT INTO email_verification_tokens (account_id, token_hash, expires_at) VALUES ($1, $2, now() + interval '24 hours')`, accountID, tokenHash); err != nil {
		return false, err
	}
	emailHash := sha256.Sum256([]byte("legal-account-email:" + strings.ToLower(input.Email)))
	if _, err := tx.Exec(ctx, `INSERT INTO legal_account_acceptances (account_id, email_hash, terms_version, terms_content_hash, source) VALUES ($1, $2, $3, $4, 'password_registration')`, accountID, emailHash[:], input.TermsVersion, legal.CurrentTermsContentHash()); err != nil {
		return false, err
	}
	if input.Draft != nil {
		if err := createTournamentFromDraft(ctx, tx, accountID, input.Draft.ID, input.Draft.Name, input.Draft.Sport, input.Draft.Teams); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// CreatePasswordReset persists a token only for a verified local account.
func (r RegistrationRepository) CreatePasswordReset(ctx context.Context, email string, tokenHash []byte) (string, registration.Locale, bool, error) {
	row, err := r.queries.CreatePasswordReset(ctx, sqlc.CreatePasswordResetParams{Lower: email, TokenHash: tokenHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}
	return row.Email, registration.Locale(row.Locale), true, nil
}

// InspectPasswordReset returns the email of a current password-reset token.
func (r RegistrationRepository) InspectPasswordReset(ctx context.Context, hash []byte) (string, error) {
	email, err := r.queries.InspectPasswordReset(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", registration.ErrPasswordResetInvalid
	}
	return email, err
}

// ConsumePasswordReset changes the credential, revokes sessions, and creates the new one atomically.
func (r RegistrationRepository) ConsumePasswordReset(ctx context.Context, tokenHash []byte, passwordHash string, sessionHash, refreshHash []byte) (registration.Session, error) {
	row, err := r.queries.ConsumePasswordReset(ctx, sqlc.ConsumePasswordResetParams{TokenHash: tokenHash, PasswordHash: passwordHash, TokenHash_2: sessionHash, TokenHash_3: refreshHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrPasswordResetInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	session := registration.Session{AccountID: row.ID.String(), Username: row.Username, LastTeamName: nullableText(row.LastTeamName)}
	session.IdleExpiresAt, session.RefreshExpiresAt = row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)
	return session, nil
}
