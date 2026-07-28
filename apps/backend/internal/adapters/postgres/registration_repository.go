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
func (r RegistrationRepository) VerifyAndCreateSession(ctx context.Context, verificationHash, sessionHash []byte) (registration.Session, error) {
	row, err := r.queries.VerifyRegistrationAndCreateSession(ctx, sqlc.VerifyRegistrationAndCreateSessionParams{TokenHash: verificationHash, TokenHash_2: sessionHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Session{}, registration.ErrVerificationInvalid
	}
	if err != nil {
		return registration.Session{}, err
	}
	return registration.Session{AccountID: row.ID.String(), Username: row.Username, IdleExpiresAt: row.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano)}, nil
}

// NewRegistrationRepository conecta el puerto del caso de uso al pool PostgreSQL.
func NewRegistrationRepository(pool *pgxpool.Pool) RegistrationRepository {
	return RegistrationRepository{queries: sqlc.New(pool)}
}

// CreatePending crea los tres registros de identidad en una sola sentencia.
func (r RegistrationRepository) CreatePending(ctx context.Context, input registration.Input, passwordHash string, tokenHash []byte) (bool, error) {
	_, err := r.queries.CreatePendingRegistration(ctx, sqlc.CreatePendingRegistrationParams{
		Email:        input.Email,
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
