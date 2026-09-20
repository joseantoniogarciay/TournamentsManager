-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE accounts
    ADD COLUMN last_team_name text,
    ADD CONSTRAINT accounts_last_team_name_valid CHECK (
        last_team_name IS NULL
        OR (
            last_team_name = btrim(last_team_name)
            AND char_length(last_team_name) BETWEEN 1 AND 100
        )
    );

-- +goose Down
-- No rollback: eliminar la columna descartaría una preferencia sincronizada de
-- la cuenta (ADR-0131).
