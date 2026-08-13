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

// FederatedRepository uses sqlc queries and a transaction for each challenge,
// identity, and session change.
type FederatedRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewFederatedRepository builds the federated-identity PostgreSQL adapter.
func NewFederatedRepository(pool *pgxpool.Pool) FederatedRepository {
	return FederatedRepository{pool: pool, queries: sqlc.New(pool)}
}

// CreateChallenge persists a Google challenge hash and expiry.
func (r FederatedRepository) CreateChallenge(ctx context.Context, nonceHash []byte, expiresAt time.Time) (string, error) {
	id, err := r.queries.CreateGoogleLoginChallenge(ctx, sqlc.CreateGoogleLoginChallengeParams{NonceHash: nonceHash, ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true}})
	return id.String(), err
}

// AuthenticateGoogle consumes the challenge and resolves the identity in a transaction.
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
	if registration.Draft != nil {
		var leagueID string
		if err := tx.QueryRow(ctx, `INSERT INTO leagues (organizer_account_id, name, state, published_at) VALUES ($1, $2, 'published', now()) RETURNING id::text`, accountID, registration.Draft.Name).Scan(&leagueID); err != nil {
			return federated.Session{}, err
		}
		for position, name := range registration.Draft.Teams {
			if _, err := tx.Exec(ctx, `INSERT INTO league_teams (league_id, name, name_normalized, position) VALUES ($1, $2, lower($2), $3)`, leagueID, name, position+1); err != nil {
				return federated.Session{}, err
			}
		}
	}
	return createSession(ctx, tx, queries, accountID, registration.Username, accessHash, refreshHash)
}

// AddGoogleIdentity adds an unlinked identity or idempotently confirms the account's own identity.
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

// ReauthenticateGoogle validates a Google identity and issues a session-bound ticket.
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

// AddGoogleIdentityWithTicket consumes the ticket and links a Google identity to the account.
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
