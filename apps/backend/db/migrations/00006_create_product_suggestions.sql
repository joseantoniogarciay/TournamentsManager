-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

CREATE TABLE product_suggestions (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    body text NOT NULL CHECK (body = btrim(body)) CHECK (char_length(body) BETWEEN 8 AND 1000),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX product_suggestions_account_idx
    ON product_suggestions (account_id, created_at DESC, id DESC);

GRANT SELECT, INSERT, UPDATE, DELETE ON product_suggestions TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: eliminar la tabla descartaría sugerencias de producto ya recibidas.
