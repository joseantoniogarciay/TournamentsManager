-- name: FindAuthenticatedAccountID :one
SELECT sessions.account_id
FROM sessions
JOIN accounts ON accounts.id = sessions.account_id
WHERE sessions.token_hash = $1
  AND sessions.revoked_at IS NULL
  AND sessions.idle_expires_at > now()
  AND sessions.absolute_expires_at > now()
  AND accounts.state = 'verified';

-- name: GetCurrentSession :one
SELECT accounts.id, accounts.username, sessions.idle_expires_at, sessions.absolute_expires_at
FROM sessions
JOIN accounts ON accounts.id = sessions.account_id
WHERE sessions.token_hash = $1
  AND sessions.revoked_at IS NULL
  AND sessions.idle_expires_at > now()
  AND sessions.absolute_expires_at > now()
  AND accounts.state = 'verified';

-- name: GetAccessMethods :one
SELECT accounts.email, accounts.username,
  EXISTS (SELECT 1 FROM local_credentials WHERE account_id = accounts.id) AS has_password,
  EXISTS (SELECT 1 FROM external_identities WHERE account_id = accounts.id AND provider = 'google') AS has_google
FROM accounts
WHERE accounts.id = $1 AND accounts.state = 'verified';

-- name: GetCurrentPasswordHash :one
SELECT local_credentials.password_hash
FROM sessions
JOIN local_credentials ON local_credentials.account_id = sessions.account_id
WHERE sessions.token_hash = $1
  AND sessions.revoked_at IS NULL
  AND sessions.idle_expires_at > now()
  AND sessions.absolute_expires_at > now();

-- name: CreateReauthenticationTicket :one
INSERT INTO reauthentication_tickets (account_id, session_id, token_hash, expires_at)
SELECT sessions.account_id, sessions.id, $2, now() + interval '5 minutes'
FROM sessions
WHERE sessions.token_hash = $1
  AND sessions.revoked_at IS NULL
  AND sessions.idle_expires_at > now()
  AND sessions.absolute_expires_at > now()
RETURNING id;

-- name: ConsumeReauthenticationTicketAndSetPassword :execrows
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
INSERT INTO local_credentials (account_id, password_hash)
SELECT account_id, $3 FROM consumed
ON CONFLICT (account_id) DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = now();

-- name: ConsumeReauthenticationTicket :one
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
RETURNING tickets.account_id;

-- name: RevokeSession :execrows
WITH revoked_session AS (
    UPDATE sessions
    SET revoked_at = now()
    WHERE sessions.token_hash = $1
      AND sessions.revoked_at IS NULL
      AND sessions.idle_expires_at > now()
      AND sessions.absolute_expires_at > now()
    RETURNING id
)
UPDATE session_refresh_tokens
SET revoked_at = now()
WHERE session_id IN (SELECT id FROM revoked_session)
  AND revoked_at IS NULL;

-- name: ListAdministeredLeagues :many
SELECT
    leagues.id,
    leagues.name,
    leagues.state,
    leagues.created_at,
    leagues.last_activity_at,
    CASE
        WHEN leagues.organizer_account_id = sqlc.arg(account_id)::uuid THEN 'organizer'
        ELSE 'delegated'
    END AS relationship
FROM leagues
LEFT JOIN league_administrators
    ON league_administrators.league_id = leagues.id
    AND league_administrators.account_id = sqlc.arg(account_id)::uuid
WHERE (
    leagues.organizer_account_id = sqlc.arg(account_id)::uuid
    OR league_administrators.account_id = sqlc.arg(account_id)::uuid
)
AND (sqlc.narg(cursor_id)::uuid IS NULL OR leagues.id < sqlc.narg(cursor_id)::uuid)
ORDER BY leagues.id DESC
LIMIT sqlc.arg(page_size);

-- name: ListFollowedLeagues :many
SELECT
    leagues.id,
    leagues.name,
    leagues.state,
    leagues.created_at,
    leagues.last_activity_at,
    'follower' AS relationship
FROM leagues
JOIN league_followers ON league_followers.league_id = leagues.id
WHERE league_followers.account_id = sqlc.arg(account_id)::uuid
  AND leagues.organizer_account_id <> sqlc.arg(account_id)::uuid
  AND NOT EXISTS (
      SELECT 1
      FROM league_administrators
      WHERE league_administrators.league_id = leagues.id
        AND league_administrators.account_id = sqlc.arg(account_id)::uuid
  )
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR leagues.id < sqlc.narg(cursor_id)::uuid)
ORDER BY leagues.id DESC
LIMIT sqlc.arg(page_size);

-- name: ListRecentAccountLeagues :many
WITH related_leagues AS (
    SELECT
        leagues.id,
        leagues.name,
        leagues.state,
        leagues.created_at,
        leagues.last_activity_at,
        CASE
            WHEN leagues.organizer_account_id = sqlc.arg(account_id)::uuid THEN 'organizer'
            ELSE 'delegated'
        END AS relationship
    FROM leagues
    LEFT JOIN league_administrators
        ON league_administrators.league_id = leagues.id
        AND league_administrators.account_id = sqlc.arg(account_id)::uuid
    WHERE leagues.organizer_account_id = sqlc.arg(account_id)::uuid
       OR league_administrators.account_id = sqlc.arg(account_id)::uuid

    UNION ALL

    SELECT
        leagues.id,
        leagues.name,
        leagues.state,
        leagues.created_at,
        leagues.last_activity_at,
        'follower' AS relationship
    FROM leagues
    JOIN league_followers ON league_followers.league_id = leagues.id
    WHERE league_followers.account_id = sqlc.arg(account_id)::uuid
      AND leagues.organizer_account_id <> sqlc.arg(account_id)::uuid
      AND NOT EXISTS (
          SELECT 1
          FROM league_administrators
          WHERE league_administrators.league_id = leagues.id
            AND league_administrators.account_id = sqlc.arg(account_id)::uuid
      )
)
SELECT id, name, state, created_at, last_activity_at, relationship
FROM related_leagues
ORDER BY last_activity_at DESC, id DESC
LIMIT 5;

-- name: FollowVisibleLeague :one
WITH visible_league AS (
    SELECT id
    FROM leagues
    WHERE id = sqlc.arg(league_id)::uuid
      AND state IN ('published', 'in_progress', 'completed', 'cancelled')
), created_follow AS (
    INSERT INTO league_followers (league_id, account_id)
    SELECT id, sqlc.arg(account_id)::uuid
    FROM visible_league
    ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM visible_league) AS visible;

-- name: UnfollowLeague :exec
DELETE FROM league_followers
WHERE league_id = sqlc.arg(league_id)::uuid
  AND account_id = sqlc.arg(account_id)::uuid;
