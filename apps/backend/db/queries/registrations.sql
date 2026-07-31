-- name: CreatePendingRegistration :one
WITH created_account AS (
    INSERT INTO accounts (email, locale, state, username, expires_at)
    VALUES ($1, $2, 'pending_verification', $3, now() + interval '7 days')
    ON CONFLICT DO NOTHING
    RETURNING id, email
), created_credentials AS (
    INSERT INTO local_credentials (account_id, password_hash)
    SELECT id, $4 FROM created_account
)
INSERT INTO email_verification_tokens (account_id, token_hash, expires_at)
SELECT id, $5, now() + interval '24 hours'
FROM created_account
RETURNING (SELECT email FROM created_account) AS email;

-- name: IsUsernameAvailable :one
SELECT NOT EXISTS (
    SELECT 1
    FROM accounts
    WHERE username = $1
) AS available;

-- name: VerifyRegistrationAndCreateSession :one
WITH consumed_token AS (
    UPDATE email_verification_tokens
    SET consumed_at = now()
    WHERE email_verification_tokens.token_hash = $1
      AND email_verification_tokens.consumed_at IS NULL
      AND email_verification_tokens.invalidated_at IS NULL
      AND email_verification_tokens.expires_at > now()
    RETURNING account_id
), verified_account AS (
    UPDATE accounts
    SET state = 'verified', verified_at = now(), expires_at = NULL
    WHERE id = (SELECT account_id FROM consumed_token)
      AND state = 'pending_verification'
    RETURNING id, username
), revoked_presented_session AS (
    UPDATE sessions
    SET revoked_at = now()
    WHERE sessions.token_hash = sqlc.narg(previous_session_hash)::bytea
      AND sessions.revoked_at IS NULL
    RETURNING id
), created_session AS (
    INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at)
    SELECT id, $2, now() + interval '7 days', now() + interval '7 days'
    FROM verified_account
    RETURNING id, idle_expires_at
), created_refresh AS (
    INSERT INTO session_refresh_tokens (session_id, token_hash, expires_at)
    SELECT id, $3, now() + interval '30 days'
    FROM created_session
    RETURNING expires_at
)
SELECT verified_account.id, verified_account.username, created_session.idle_expires_at, created_refresh.expires_at
FROM verified_account
CROSS JOIN created_session
CROSS JOIN created_refresh;

-- name: RotateSessionTokens :one
WITH consumed_refresh AS (
    UPDATE session_refresh_tokens
    SET consumed_at = now()
    WHERE session_refresh_tokens.token_hash = $1
      AND session_refresh_tokens.consumed_at IS NULL
      AND session_refresh_tokens.revoked_at IS NULL
      AND session_refresh_tokens.expires_at > now()
    RETURNING session_id
), rotated_session AS (
    UPDATE sessions
    SET token_hash = $2,
        last_seen_at = now(),
        idle_expires_at = now() + interval '7 days',
        absolute_expires_at = now() + interval '7 days'
    WHERE id = (SELECT session_id FROM consumed_refresh)
      AND revoked_at IS NULL
    RETURNING id, account_id, idle_expires_at
), created_refresh AS (
    INSERT INTO session_refresh_tokens (session_id, token_hash, expires_at)
    SELECT id, $3, now() + interval '30 days'
    FROM rotated_session
    RETURNING expires_at
)
SELECT accounts.id, accounts.username, rotated_session.idle_expires_at, created_refresh.expires_at
FROM rotated_session
JOIN accounts ON accounts.id = rotated_session.account_id
CROSS JOIN created_refresh;
