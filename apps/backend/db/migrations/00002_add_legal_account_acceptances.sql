-- +goose Up
-- Esta migración se aplica con la identidad de migración, nunca por la API.
SET ROLE tournaments_manager_dev_schema_owner;

CREATE TABLE legal_account_acceptances (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid REFERENCES accounts (id) ON DELETE SET NULL,
    email_hash bytea NOT NULL CHECK (octet_length(email_hash) = 32),
    terms_version text NOT NULL,
    terms_content_hash bytea NOT NULL CHECK (octet_length(terms_content_hash) = 32),
    source text NOT NULL CHECK (source IN ('password_registration', 'google_registration')),
    accepted_at timestamptz NOT NULL DEFAULT now(),
    retention_until timestamptz
);

CREATE INDEX legal_account_acceptances_retention_until_idx
    ON legal_account_acceptances (retention_until)
    WHERE retention_until IS NOT NULL;

GRANT SELECT, INSERT, UPDATE, DELETE ON legal_account_acceptances TO tournaments_manager_dev_app;

-- +goose Down
-- No se proporciona rollback: eliminar evidencia legal podría incumplir el plazo de conservación.
