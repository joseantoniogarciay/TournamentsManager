-- name: AuthenticateSessionWithClient :one
WITH active AS MATERIALIZED (
    SELECT sessions.id, sessions.account_id
    FROM sessions JOIN accounts ON accounts.id = sessions.account_id
    WHERE sessions.token_hash = sqlc.arg(token_hash)
      AND sessions.revoked_at IS NULL
      AND sessions.idle_expires_at > now() AND sessions.absolute_expires_at > now()
      AND accounts.state = 'verified'
), touched AS (
    UPDATE sessions SET last_seen_at = now(),
      client_platform = CASE WHEN client_platform = 'unknown' THEN sqlc.arg(client_platform)::text ELSE client_platform END,
      client_kind = CASE WHEN client_kind = 'unknown' THEN sqlc.arg(client_kind)::text ELSE client_kind END
    WHERE id IN (SELECT id FROM active)
      AND revoked_at IS NULL
      AND (last_seen_at < now() - interval '5 minutes'
        OR (client_platform = 'unknown' AND sqlc.arg(client_platform)::text <> 'unknown')
        OR (client_kind = 'unknown' AND sqlc.arg(client_kind)::text <> 'unknown'))
)
SELECT account_id FROM active;

-- name: ListSessions :many
SELECT sessions.id, sessions.created_at, sessions.last_seen_at,
       sessions.client_platform, sessions.client_kind,
       sessions.token_hash = sqlc.arg(token_hash) AS current
FROM sessions
WHERE sessions.account_id = (
    SELECT current_session.account_id FROM sessions AS current_session
    JOIN accounts ON accounts.id = current_session.account_id
    WHERE current_session.token_hash = sqlc.arg(token_hash)
      AND current_session.revoked_at IS NULL
      AND current_session.idle_expires_at > now() AND current_session.absolute_expires_at > now()
      AND accounts.state = 'verified'
)
  AND sessions.revoked_at IS NULL
  AND ((sessions.idle_expires_at > now() AND sessions.absolute_expires_at > now()) OR EXISTS (
    SELECT 1 FROM session_refresh_tokens WHERE session_id = sessions.id
      AND revoked_at IS NULL AND consumed_at IS NULL AND expires_at > now()
  ))
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR sessions.id < sqlc.narg(cursor_id)::uuid)
ORDER BY sessions.id DESC LIMIT 51;

-- name: RevokeOtherSessions :one
WITH current_session AS MATERIALIZED (
    SELECT sessions.id, sessions.account_id
    FROM sessions JOIN accounts ON accounts.id = sessions.account_id
    WHERE sessions.token_hash = sqlc.arg(token_hash)
      AND sessions.revoked_at IS NULL
      AND sessions.idle_expires_at > now() AND sessions.absolute_expires_at > now()
      AND accounts.state = 'verified'
    FOR UPDATE OF sessions
), revoked AS (
    UPDATE sessions SET revoked_at = now()
    FROM current_session
    WHERE sessions.account_id = current_session.account_id
      AND sessions.id <> current_session.id
      AND (sqlc.narg(target_id)::uuid IS NULL OR sessions.id = sqlc.narg(target_id)::uuid)
      AND sessions.revoked_at IS NULL
    RETURNING sessions.id
), refresh_revoked AS (
    UPDATE session_refresh_tokens SET revoked_at = now()
    WHERE session_id IN (SELECT id FROM revoked) AND revoked_at IS NULL
)
SELECT id FROM current_session;
