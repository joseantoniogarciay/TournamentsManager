-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE legal_account_acceptances
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

CREATE INDEX legal_account_acceptances_updated_at_idx
    ON legal_account_acceptances (updated_at, id);

-- +goose Down
-- No rollback: the column is required to preserve backup retention evidence.
