-- name: CreatePendingRegistration :one
WITH created_account AS (
    INSERT INTO accounts (email, state, username, expires_at)
    VALUES ($1, 'pending_verification', $2, now() + interval '7 days')
    ON CONFLICT DO NOTHING
    RETURNING id, email
), created_credentials AS (
    INSERT INTO local_credentials (account_id, password_hash)
    SELECT id, $3 FROM created_account
)
INSERT INTO email_verification_tokens (account_id, token_hash, expires_at)
SELECT id, $4, now() + interval '24 hours'
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
), created_session AS (
    INSERT INTO sessions (account_id, token_hash, idle_expires_at, absolute_expires_at)
    SELECT id, $2, now() + interval '7 days', now() + interval '30 days'
    FROM verified_account
    RETURNING idle_expires_at
)
SELECT verified_account.id, verified_account.username, created_session.idle_expires_at
FROM verified_account
CROSS JOIN created_session;
