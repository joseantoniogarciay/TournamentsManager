package postgres

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/adapters/postgres/sqlc"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

// CreateAppleChallenge persists digests, never a provider token or client secret.
func (r FederatedRepository) CreateAppleChallenge(ctx context.Context, stateHash, nonceHash, proofHash []byte, platform string, expires time.Time) (string, error) {
	id, err := r.queries.CreateAppleLoginChallenge(ctx, sqlc.CreateAppleLoginChallengeParams{StateHash: stateHash, NonceHash: nonceHash, ProofHash: proofHash, Platform: platform, ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true}})
	return id.String(), err
}

// ClaimAppleCallback atomically prevents duplicate exchanges across API replicas.
func (r FederatedRepository) ClaimAppleCallback(ctx context.Context, hash []byte) (federated.AppleCallback, error) {
	row, err := r.queries.ClaimAppleCallback(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return federated.AppleCallback{}, federated.ErrChallengeInvalid
	}
	if err != nil {
		return federated.AppleCallback{}, err
	}
	return federated.AppleCallback{ID: row.ID.String(), Platform: row.Platform, NonceHash: row.NonceHash}, nil
}

// CompleteAppleCallback stores only the verified, short-lived identity claims.
func (r FederatedRepository) CompleteAppleCallback(ctx context.Context, id string, identity federated.Identity) error {
	parsed, err := parseUUID(id)
	if err != nil {
		return federated.ErrChallengeInvalid
	}
	count, err := r.queries.CompleteAppleCallback(ctx, sqlc.CompleteAppleCallbackParams{ID: parsed, Subject: pgtype.Text{String: identity.Subject, Valid: true}, Email: pgtype.Text{String: identity.Email, Valid: identity.Email != ""}, EmailVerified: identity.EmailVerified})
	if err != nil {
		return err
	}
	if count != 1 {
		return federated.ErrChallengeInvalid
	}
	return nil
}

// AuthenticateApple consumes the initiating app's proof and creates account,
// legal evidence, draft and session in the same transaction. 202 rolls it back.
func (r FederatedRepository) AuthenticateApple(ctx context.Context, id string, proofHash []byte, registration *federated.Registration, draft *federated.Draft, accessHash, refreshHash []byte) (federated.Session, error) {
	parsed, err := parseUUID(id)
	if err != nil {
		return federated.Session{}, federated.ErrChallengeInvalid
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return federated.Session{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	row, err := queries.GetAppleLoginForUpdate(ctx, parsed)
	if errors.Is(err, pgx.ErrNoRows) {
		return federated.Session{}, federated.ErrChallengeInvalid
	}
	if err != nil {
		return federated.Session{}, err
	}
	if subtle.ConstantTimeCompare(proofHash, row.ProofHash) != 1 {
		return federated.Session{}, federated.ErrChallengeInvalid
	}
	if err := queries.ConsumeAppleLogin(ctx, parsed); err != nil {
		return federated.Session{}, err
	}
	identity := federated.Identity{Issuer: federated.AppleIssuer, Subject: row.Subject.String, Email: row.Email.String, EmailVerified: row.EmailVerified}
	return resolveFederatedSession(ctx, tx, queries, "apple", identity, registration, draft, accessHash, refreshHash)
}

var _ federated.AppleRepository = FederatedRepository{}
