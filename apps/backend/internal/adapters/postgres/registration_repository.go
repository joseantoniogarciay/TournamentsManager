package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

// RegistrationRepository implementa la persistencia del alta local con sqlc.
type RegistrationRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// VerifyAndCreateSession consume una verificación y emite su sesión atómica.
func (r RegistrationRepository) VerifyAndCreateSession(ctx context.Context, verificationHash, sessionHash, refreshHash, previousSessionHash []byte) (registration.Session, error) {
	row, err := r.queries.VerifyRegistrationAndCreateSession(ctx, sqlc.VerifyRegistrationAndCreateSessionParams{TokenHash: verificationHash, TokenHash_2: sessionHash, TokenHash_3: refreshHash, PreviousSessionHash: previousSessionHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrVerificationInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: row.ID.String(), Username: row.Username, IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// RotateSessionTokens consume un refresh y crea de forma atómica sus sucesores.
func (r RegistrationRepository) RotateSessionTokens(ctx context.Context, refreshHash, sessionHash, nextRefreshHash []byte) (registration.Session, error) {
	row, err := r.queries.RotateSessionTokens(ctx, sqlc.RotateSessionTokensParams{TokenHash: refreshHash, TokenHash_2: sessionHash, TokenHash_3: nextRefreshHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrRefreshInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: row.ID.String(), Username: row.Username, IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// NewRegistrationRepository conecta el puerto del caso de uso al pool PostgreSQL.
func NewRegistrationRepository(pool *pgxpool.Pool) RegistrationRepository {
	return RegistrationRepository{pool: pool, queries: sqlc.New(pool)}
}

// IsUsernameAvailable consulta la restricción de unicidad sin crear una reserva.
func (r RegistrationRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	return r.queries.IsUsernameAvailable(ctx, username)
}

// FindLocalAccountForLogin obtiene la credencial local y el estado de verificación por correo.
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

// CreateLocalLoginSession persiste los tokens hasheados de una nueva sesión local.
func (r RegistrationRepository) CreateLocalLoginSession(ctx context.Context, accountID string, sessionHash, refreshHash []byte) (registration.Session, error) {
	id, err := parseUUID(accountID)
	if err != nil {
		return registration.Session{}, err
	}
	row, err := r.queries.CreateLocalLoginSession(ctx, sqlc.CreateLocalLoginSessionParams{ID: id, TokenHash: sessionHash, TokenHash_2: refreshHash})
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: accountID, Username: row.Username, IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// RenewLoginVerification rota el token de verificación pendiente y devuelve su destinatario.
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

// CreatePending crea los tres registros de identidad en una sola sentencia.
func (r RegistrationRepository) CreatePending(ctx context.Context, input registration.Input, passwordHash string, tokenHash []byte) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var accountID string
	err = tx.QueryRow(ctx, `INSERT INTO accounts (email, locale, state, username, expires_at)
		VALUES ($1, $2, 'pending_verification', $3, now() + interval '7 days')
		ON CONFLICT DO NOTHING
		RETURNING id::text`, input.Email, string(input.Locale), input.Username).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO local_credentials (account_id, password_hash) VALUES ($1, $2)`, accountID, passwordHash); err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO email_verification_tokens (account_id, token_hash, expires_at) VALUES ($1, $2, now() + interval '24 hours')`, accountID, tokenHash); err != nil {
		return false, err
	}
	if input.Draft != nil {
		var leagueID string
		if err := tx.QueryRow(ctx, `INSERT INTO leagues (organizer_account_id, name, state, published_at)
			VALUES ($1, $2, 'published', now()) RETURNING id::text`, accountID, input.Draft.Name).Scan(&leagueID); err != nil {
			return false, err
		}
		for position, name := range input.Draft.Teams {
			if _, err := tx.Exec(ctx, `INSERT INTO league_teams (league_id, name, name_normalized, position)
				VALUES ($1, $2, lower($2), $3)`, leagueID, name, position+1); err != nil {
				return false, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// CreatePasswordReset persiste un token solo para una cuenta local verificada.
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

// InspectPasswordReset devuelve el email de un token de restablecimiento vigente.
func (r RegistrationRepository) InspectPasswordReset(ctx context.Context, hash []byte) (string, error) {
	email, err := r.queries.InspectPasswordReset(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", registration.ErrPasswordResetInvalid
	}
	return email, err
}

// ConsumePasswordReset cambia la credencial, revoca sesiones y crea la nueva atómicamente.
func (r RegistrationRepository) ConsumePasswordReset(ctx context.Context, tokenHash []byte, passwordHash string, sessionHash, refreshHash []byte) (registration.Session, error) {
	row, err := r.queries.ConsumePasswordReset(ctx, sqlc.ConsumePasswordResetParams{TokenHash: tokenHash, PasswordHash: passwordHash, TokenHash_2: sessionHash, TokenHash_3: refreshHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrPasswordResetInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	session := registration.Session{AccountID: row.ID.String(), Username: row.Username}
	session.IdleExpiresAt, session.RefreshExpiresAt = row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), row.ExpiresAt.Time.UTC().Format(time.RFC3339Nano)
	return session, nil
}
