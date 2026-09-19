-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

CREATE TABLE tournament_team_invitations (
    tournament_id uuid PRIMARY KEY REFERENCES tournaments (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tournament_team_accounts (
    tournament_id uuid NOT NULL,
    team_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    joined_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tournament_id, account_id),
    CONSTRAINT tournament_team_accounts_team_unique UNIQUE (tournament_id, team_id),
    CONSTRAINT tournament_team_accounts_team_fk
        FOREIGN KEY (tournament_id, team_id)
        REFERENCES tournament_teams (tournament_id, id)
        ON DELETE CASCADE
);

CREATE INDEX tournament_team_accounts_account_idx
    ON tournament_team_accounts (account_id, tournament_id DESC);

GRANT SELECT, INSERT, UPDATE, DELETE ON tournament_team_invitations TO tournaments_manager_dev_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON tournament_team_accounts TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: eliminar estas relaciones descartaría invitaciones activas y la
-- vinculación aceptada entre cuentas y equipos (ADR-0130).
