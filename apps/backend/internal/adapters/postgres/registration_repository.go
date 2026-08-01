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
	return RegistrationRepository{queries: sqlc.New(pool)}
}

// IsUsernameAvailable consulta la restricción de unicidad sin crear una reserva.
func (r RegistrationRepository) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	return r.queries.IsUsernameAvailable(ctx, username)
}

// CreatePending crea los tres registros de identidad en una sola sentencia.
func (r RegistrationRepository) CreatePending(ctx context.Context, input registration.Input, passwordHash string, tokenHash []byte) (bool, error) {
	_, err := r.queries.CreatePendingRegistration(ctx, sqlc.CreatePendingRegistrationParams{
		Email:        input.Email,
		Locale:       string(input.Locale),
		Username:     input.Username,
		PasswordHash: passwordHash,
		TokenHash:    tokenHash,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
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
