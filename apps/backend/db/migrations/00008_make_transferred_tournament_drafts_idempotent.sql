-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments ADD COLUMN source_draft_id uuid;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_organizer_source_draft_unique
    UNIQUE (organizer_account_id, source_draft_id);

-- +goose Down
-- No rollback: retirar la identidad de origen permitiría duplicar torneos al
-- reintentar una transferencia cuya respuesta se perdió.
