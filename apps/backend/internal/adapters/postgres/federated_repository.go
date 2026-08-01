package postgres

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

// FederatedRepository usa consultas sqlc y una transacción para cada cambio de
// challenge, identidad y sesión.
type FederatedRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewFederatedRepository construye el adaptador PostgreSQL de identidad federada.
func NewFederatedRepository(pool *pgxpool.Pool) FederatedRepository {
	return FederatedRepository{pool: pool, queries: sqlc.New(pool)}
}

// CreateChallenge persiste el hash y la caducidad de una prueba Google.
func (r FederatedRepository) CreateChallenge(ctx context.Context, nonceHash []byte, expiresAt time.Time) (string, error) {
	id, err := r.queries.CreateGoogleLoginChallenge(ctx, sqlc.CreateGoogleLoginChallengeParams{NonceHash: nonceHash, ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true}})
	return id.String(), err
}

// AuthenticateGoogle consume el challenge y resuelve la identidad en una transacción.
func (r FederatedRepository) AuthenticateGoogle(ctx context.Context, challengeID string, nonceHash []byte, identity federated.Identity, registration *federated.Registration, accessHash, refreshHash []byte) (federated.Session, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return federated.Session{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if err := consumeChallenge(ctx, queries, challengeID, nonceHash); err != nil {
		return federated.Session{}, err
	}

	account, err := queries.FindGoogleIdentityAccount(ctx, sqlc.FindGoogleIdentityAccountParams{Issuer: identity.Issuer, Subject: identity.Subject})
	if err == nil {
		return createSession(ctx, tx, queries, account.ID, account.Username, accessHash, refreshHash)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return federated.Session{}, err
	}
	if exists, err := queries.AccountEmailExists(ctx, identity.Email); err != nil {
		return federated.Session{}, err
	} else if exists {
		return federated.Session{}, federated.ErrEmailConflict
	}
	if registration == nil {
		return federated.Session{}, federated.ErrRegistration
	}

	accountID, err := queries.CreateGoogleAccount(ctx, sqlc.CreateGoogleAccountParams{Email: identity.Email, Locale: registration.Locale, Username: registration.Username})
	if err != nil {
		return federated.Session{}, err
	}
	if err := queries.CreateGoogleExternalIdentity(ctx, sqlc.CreateGoogleExternalIdentityParams{AccountID: accountID, Issuer: identity.Issuer, Subject: identity.Subject}); err != nil {
		return federated.Session{}, err
	}
	return createSession(ctx, tx, queries, accountID, registration.Username, accessHash, refreshHash)
}

// AddGoogleIdentity agrega una identidad libre o confirma idempotentemente la propia.
func (r FederatedRepository) AddGoogleIdentity(ctx context.Context, accountID, challengeID string, nonceHash []byte, identity federated.Identity) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if err := consumeChallenge(ctx, queries, challengeID, nonceHash); err != nil {
		return err
	}
	owner, err := queries.FindGoogleIdentityOwner(ctx, sqlc.FindGoogleIdentityOwnerParams{Issuer: identity.Issuer, Subject: identity.Subject})
	if err == nil {
		if owner.String() != accountID {
			return federated.ErrIdentityConflict
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	parsedAccountID, err := parseUUID(accountID)
	if err != nil {
		return err
	}
	if err := queries.CreateGoogleExternalIdentity(ctx, sqlc.CreateGoogleExternalIdentityParams{AccountID: parsedAccountID, Issuer: identity.Issuer, Subject: identity.Subject}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReauthenticateGoogle valida una identidad de Google y emite un ticket asociado a la sesión.
func (r FederatedRepository) ReauthenticateGoogle(ctx context.Context, accountID, sessionToken, challengeID string, nonceHash []byte, identity federated.Identity, ticketHash []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if err := consumeChallenge(ctx, queries, challengeID, nonceHash); err != nil {
		return err
	}
	owner, err := queries.FindGoogleIdentityOwner(ctx, sqlc.FindGoogleIdentityOwnerParams{Issuer: identity.Issuer, Subject: identity.Subject})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner.String() != accountID) {
		return federated.ErrIdentityConflict
	}
	if err != nil {
		return err
	}
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	if _, err := queries.CreateReauthenticationTicket(ctx, sqlc.CreateReauthenticationTicketParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// AddGoogleIdentityWithTicket consume el ticket y vincula una identidad de Google a la cuenta.
func (r FederatedRepository) AddGoogleIdentityWithTicket(ctx context.Context, sessionToken, challengeID string, nonceHash []byte, identity federated.Identity, ticketHash []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sessionHash := sha256.Sum256([]byte("session:" + sessionToken))
	accountID, err := queries.ConsumeReauthenticationTicket(ctx, sqlc.ConsumeReauthenticationTicketParams{TokenHash: sessionHash[:], TokenHash_2: ticketHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return federated.ErrChallengeInvalid
	}
	if err != nil {
		return err
	}
	if err := consumeChallenge(ctx, queries, challengeID, nonceHash); err != nil {
		return err
	}
	owner, err := queries.FindGoogleIdentityOwner(ctx, sqlc.FindGoogleIdentityOwnerParams{Issuer: identity.Issuer, Subject: identity.Subject})
	if err == nil {
		if owner != accountID {
			return federated.ErrIdentityConflict
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err := queries.CreateGoogleExternalIdentity(ctx, sqlc.CreateGoogleExternalIdentityParams{AccountID: accountID, Issuer: identity.Issuer, Subject: identity.Subject}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func consumeChallenge(ctx context.Context, queries *sqlc.Queries, challengeID string, nonceHash []byte) error {
	id, err := parseUUID(challengeID)
	if err != nil {
		return federated.ErrChallengeInvalid
	}
	stored, err := queries.GetActiveGoogleChallengeForUpdate(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return federated.ErrChallengeInvalid
	}
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare(stored, nonceHash) != 1 {
		return federated.ErrChallengeInvalid
	}
	return queries.ConsumeGoogleLoginChallenge(ctx, id)
}

func createSession(ctx context.Context, tx pgx.Tx, queries *sqlc.Queries, accountID pgtype.UUID, username string, accessHash, refreshHash []byte) (federated.Session, error) {
	created, err := queries.CreateFederatedSession(ctx, sqlc.CreateFederatedSessionParams{AccountID: accountID, TokenHash: accessHash})
	if err != nil {
		return federated.Session{}, err
	}
	refreshExpires, err := queries.CreateFederatedRefreshToken(ctx, sqlc.CreateFederatedRefreshTokenParams{SessionID: created.ID, TokenHash: refreshHash})
	if err != nil {
		return federated.Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return federated.Session{}, err
	}
	return federated.Session{AccountID: accountID.String(), Username: username, IdleExpiresAt: created.IdleExpiresAt.Time.UTC().Format(time.RFC3339Nano), RefreshExpiresAt: refreshExpires.Time.UTC().Format(time.RFC3339Nano)}, nil
}

func parseUUID(value string) (pgtype.UUID, error) { var id pgtype.UUID; return id, id.Scan(value) }

var _ federated.Repository = FederatedRepository{}
