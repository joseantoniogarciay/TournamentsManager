-- name: FindAuthenticatedAccountID :one
SELECT sessions.account_id
FROM sessions
JOIN accounts ON accounts.id = sessions.account_id
WHERE sessions.token_hash = $1
  AND sessions.revoked_at IS NULL
  AND sessions.idle_expires_at > now()
  AND sessions.absolute_expires_at > now()
  AND accounts.state = 'verified';

-- name: ListAdministeredLeagues :many
SELECT
    leagues.id,
    leagues.name,
    leagues.state,
    leagues.created_at,
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

-- name: FollowVisibleLeague :one
WITH visible_league AS (
    SELECT id
    FROM leagues
    WHERE id = sqlc.arg(league_id)::uuid
      AND state <> 'draft'
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
