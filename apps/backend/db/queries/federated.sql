-- name: CreateGoogleLoginChallenge :one
INSERT INTO federated_login_challenges (provider, nonce_hash, expires_at)
VALUES ('google', $1, $2)
RETURNING id;

-- name: GetActiveGoogleChallengeForUpdate :one
SELECT nonce_hash
FROM federated_login_challenges
WHERE id = $1
  AND provider = 'google'
  AND consumed_at IS NULL
  AND expires_at > now()
FOR UPDATE;

-- name: ConsumeGoogleLoginChallenge :exec
UPDATE federated_login_challenges SET consumed_at = now() WHERE id = $1;

-- name: FindGoogleIdentityAccount :one
SELECT accounts.id, accounts.username
FROM external_identities
JOIN accounts ON accounts.id = external_identities.account_id
WHERE external_identities.issuer = $1 AND external_identities.subject = $2;

-- name: AccountEmailExists :one
SELECT EXISTS (SELECT 1 FROM accounts WHERE lower(email) = lower($1)) AS exists;

-- name: CreateGoogleAccount :one
INSERT INTO accounts (email, locale, state, username, verified_at)
VALUES ($1, $2, 'verified', $3, now())
RETURNING id;

-- name: CreateGoogleExternalIdentity :exec
INSERT INTO external_identities (account_id, provider, issuer, subject)
VALUES ($1, 'google', $2, $3);

-- name: CreateFederatedSession :one
INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at)
VALUES ($1, $2, now() + interval '7 days', now() + interval '7 days')
RETURNING id, idle_expires_at;

-- name: CreateFederatedRefreshToken :one
INSERT INTO session_refresh_tokens (session_id, token_hash, expires_at)
VALUES ($1, $2, now() + interval '30 days')
RETURNING expires_at;

-- name: FindGoogleIdentityOwner :one
SELECT account_id FROM external_identities WHERE issuer = $1 AND subject = $2;

-- name: ConsumeReauthenticationTicketAndRemoveGoogle :execrows
WITH consumed AS (
    UPDATE reauthentication_tickets AS tickets
    SET consumed_at = now()
    FROM sessions
    WHERE tickets.token_hash = $2
      AND tickets.session_id = sessions.id
      AND sessions.token_hash = $1
      AND sessions.revoked_at IS NULL
      AND sessions.idle_expires_at > now()
      AND sessions.absolute_expires_at > now()
      AND tickets.consumed_at IS NULL
      AND tickets.expires_at > now()
    RETURNING tickets.account_id
)
DELETE FROM external_identities
WHERE account_id IN (SELECT account_id FROM consumed)
  AND provider = 'google'
  AND EXISTS (
      SELECT 1 FROM local_credentials
      WHERE local_credentials.account_id = external_identities.account_id
  );
