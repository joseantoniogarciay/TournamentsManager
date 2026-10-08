-- name: CreateAppleLoginChallenge :one
INSERT INTO apple_login_challenges (state_hash, nonce_hash, proof_hash, platform, expires_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: ClaimAppleCallback :one
UPDATE apple_login_challenges SET callback_claimed_at = now()
WHERE state_hash = $1 AND callback_claimed_at IS NULL AND expires_at > now()
RETURNING id, nonce_hash, platform;

-- name: CompleteAppleCallback :execrows
UPDATE apple_login_challenges SET subject = $2, email = $3, email_verified = $4
WHERE id = $1 AND callback_claimed_at IS NOT NULL AND subject IS NULL AND expires_at > now();

-- name: GetAppleLoginForUpdate :one
SELECT proof_hash, subject, email, email_verified FROM apple_login_challenges
WHERE id = $1 AND consumed_at IS NULL AND subject IS NOT NULL AND expires_at > now()
FOR UPDATE;

-- name: ConsumeAppleLogin :exec
DELETE FROM apple_login_challenges WHERE id = $1;

-- name: CreateAppleExternalIdentity :exec
INSERT INTO external_identities (account_id, provider, issuer, subject)
VALUES ($1, 'apple', $2, $3);
