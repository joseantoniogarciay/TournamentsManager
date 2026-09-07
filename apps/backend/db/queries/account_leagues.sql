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

-- name: ConsumeReauthenticationTicketAndRemovePassword :execrows
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
DELETE FROM local_credentials
WHERE account_id IN (SELECT account_id FROM consumed)
  AND EXISTS (
      SELECT 1 FROM external_identities
      WHERE external_identities.account_id = local_credentials.account_id
        AND external_identities.provider = 'google'
  );

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

-- name: ListAdministeredTournaments :many
SELECT
    tournaments.id,
    tournaments.name,
    tournaments.state,
    tournaments.created_at,
    tournaments.last_activity_at,
    CASE
        WHEN tournaments.organizer_account_id = sqlc.arg(account_id)::uuid THEN 'organizer'
        ELSE 'delegated'
    END AS relationship
FROM tournaments
LEFT JOIN tournament_administrators
    ON tournament_administrators.tournament_id = tournaments.id
    AND tournament_administrators.account_id = sqlc.arg(account_id)::uuid
WHERE (
    tournaments.organizer_account_id = sqlc.arg(account_id)::uuid
    OR tournament_administrators.account_id = sqlc.arg(account_id)::uuid
)
AND (sqlc.narg(cursor_id)::uuid IS NULL OR tournaments.id < sqlc.narg(cursor_id)::uuid)
ORDER BY tournaments.id DESC
LIMIT sqlc.arg(page_size);

-- name: ListFollowedTournaments :many
SELECT
    tournaments.id,
    tournaments.name,
    tournaments.state,
    tournaments.created_at,
    tournaments.last_activity_at,
    'follower' AS relationship
FROM tournaments
JOIN tournament_followers ON tournament_followers.tournament_id = tournaments.id
WHERE tournament_followers.account_id = sqlc.arg(account_id)::uuid
  AND tournaments.organizer_account_id <> sqlc.arg(account_id)::uuid
  AND NOT EXISTS (
      SELECT 1
      FROM tournament_administrators
      WHERE tournament_administrators.tournament_id = tournaments.id
        AND tournament_administrators.account_id = sqlc.arg(account_id)::uuid
  )
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR tournaments.id < sqlc.narg(cursor_id)::uuid)
ORDER BY tournaments.id DESC
LIMIT sqlc.arg(page_size);

-- name: ListRecentAccountTournaments :many
WITH related_tournaments AS (
    SELECT
        tournaments.id,
        tournaments.name,
        tournaments.state,
        tournaments.created_at,
        tournaments.last_activity_at,
        CASE
            WHEN tournaments.organizer_account_id = sqlc.arg(account_id)::uuid THEN 'organizer'
            ELSE 'delegated'
        END AS relationship
    FROM tournaments
    LEFT JOIN tournament_administrators
        ON tournament_administrators.tournament_id = tournaments.id
        AND tournament_administrators.account_id = sqlc.arg(account_id)::uuid
    WHERE tournaments.organizer_account_id = sqlc.arg(account_id)::uuid
       OR tournament_administrators.account_id = sqlc.arg(account_id)::uuid

    UNION ALL

    SELECT
        tournaments.id,
        tournaments.name,
        tournaments.state,
        tournaments.created_at,
        tournaments.last_activity_at,
        'follower' AS relationship
    FROM tournaments
    JOIN tournament_followers ON tournament_followers.tournament_id = tournaments.id
    WHERE tournament_followers.account_id = sqlc.arg(account_id)::uuid
      AND tournaments.organizer_account_id <> sqlc.arg(account_id)::uuid
      AND NOT EXISTS (
          SELECT 1
          FROM tournament_administrators
          WHERE tournament_administrators.tournament_id = tournaments.id
            AND tournament_administrators.account_id = sqlc.arg(account_id)::uuid
      )
)
SELECT id, name, state, created_at, last_activity_at, relationship
FROM related_tournaments
ORDER BY last_activity_at DESC, id DESC
LIMIT 5;

-- name: FollowVisibleTournament :one
WITH visible_league AS (
    SELECT id
    FROM tournaments
    WHERE id = sqlc.arg(tournament_id)::uuid
      AND state IN ('published', 'in_progress', 'completed', 'cancelled')
), created_follow AS (
    INSERT INTO tournament_followers (tournament_id, account_id)
    SELECT id, sqlc.arg(account_id)::uuid
    FROM visible_league
    ON CONFLICT DO NOTHING
)
SELECT EXISTS (SELECT 1 FROM visible_league) AS visible;

-- name: UnfollowTournament :exec
DELETE FROM tournament_followers
WHERE tournament_id = sqlc.arg(tournament_id)::uuid
  AND account_id = sqlc.arg(account_id)::uuid;
