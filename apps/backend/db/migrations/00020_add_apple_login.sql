-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE external_identities DROP CONSTRAINT external_identities_provider_check;
ALTER TABLE external_identities ADD CONSTRAINT external_identities_provider_check CHECK (provider IN ('google', 'apple'));
ALTER TABLE external_identities DROP CONSTRAINT external_identities_issuer_check;
ALTER TABLE external_identities ADD CONSTRAINT external_identities_issuer_check CHECK (
    (provider = 'google' AND issuer = 'https://accounts.google.com') OR
    (provider = 'apple' AND issuer = 'https://appleid.apple.com')
);
ALTER TABLE legal_account_acceptances DROP CONSTRAINT legal_account_acceptances_source_check;
ALTER TABLE legal_account_acceptances ADD CONSTRAINT legal_account_acceptances_source_check CHECK (source IN ('password_registration', 'google_registration', 'apple_registration'));

CREATE TABLE apple_login_challenges (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    state_hash bytea NOT NULL UNIQUE CHECK (octet_length(state_hash) = 32),
    nonce_hash bytea NOT NULL CHECK (octet_length(nonce_hash) = 32),
    proof_hash bytea NOT NULL CHECK (octet_length(proof_hash) = 32),
    platform text NOT NULL CHECK (platform IN ('web', 'ios', 'android')),
    subject text,
    email text,
    email_verified boolean NOT NULL DEFAULT false,
    callback_claimed_at timestamptz,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (subject IS NULL OR callback_claimed_at IS NOT NULL),
    CHECK (consumed_at IS NULL OR subject IS NOT NULL)
);
CREATE INDEX apple_login_challenges_expiry_idx ON apple_login_challenges (expires_at);
GRANT SELECT, INSERT, UPDATE, DELETE ON apple_login_challenges TO tournaments_manager_dev_app;

-- +goose Down
-- No destructive rollback: Apple identities and legal evidence must be preserved.
