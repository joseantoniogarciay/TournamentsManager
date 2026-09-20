-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT tournaments_format_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_format_check
    CHECK (format IN ('league', 'single_elimination', 'league_then_single_elimination'));

ALTER TABLE tournament_stages
    ADD COLUMN league_structure text CHECK (league_structure IN ('single_table', 'groups')),
    ADD COLUMN qualifier_count smallint CHECK (qualifier_count BETWEEN 2 AND 64),
    ADD COLUMN group_count smallint CHECK (group_count BETWEEN 2 AND 64),
    ADD COLUMN qualifiers_per_group smallint CHECK (qualifiers_per_group BETWEEN 1 AND 63),
    ADD CONSTRAINT tournament_stages_mixed_configuration CHECK (
        (league_structure IS NULL AND qualifier_count IS NULL AND group_count IS NULL AND qualifiers_per_group IS NULL)
        OR (type = 'league' AND league_structure = 'single_table' AND qualifier_count IS NOT NULL
            AND group_count IS NULL AND qualifiers_per_group IS NULL)
        OR (type = 'league' AND league_structure = 'groups' AND qualifier_count IS NULL
            AND group_count IS NOT NULL AND qualifiers_per_group IS NOT NULL)
    );

CREATE TABLE tournament_stage_teams (
    tournament_id uuid NOT NULL,
    stage_id uuid NOT NULL,
    team_id uuid NOT NULL,
    seed_position integer NOT NULL CHECK (seed_position > 0),
    group_number integer CHECK (group_number > 0),
    PRIMARY KEY (stage_id, team_id),
    UNIQUE (stage_id, seed_position),
    FOREIGN KEY (tournament_id, stage_id)
        REFERENCES tournament_stages (tournament_id, id) ON DELETE CASCADE,
    FOREIGN KEY (tournament_id, team_id)
        REFERENCES tournament_teams (tournament_id, id) ON DELETE RESTRICT
);

ALTER TABLE matches ADD COLUMN group_number integer CHECK (group_number > 0);

GRANT SELECT, INSERT, UPDATE, DELETE ON tournament_stage_teams TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: persisted stage composition is part of the accepted tournament history.
