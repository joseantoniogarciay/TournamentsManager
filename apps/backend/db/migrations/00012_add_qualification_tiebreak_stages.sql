-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournament_stages DROP CONSTRAINT tournament_stages_type_check;
ALTER TABLE tournament_stages ADD CONSTRAINT tournament_stages_type_check
    CHECK (type IN ('league', 'qualification_tiebreak', 'single_elimination'));

ALTER TABLE tournament_stages DROP CONSTRAINT tournament_stages_rules;
ALTER TABLE tournament_stages ADD CONSTRAINT tournament_stages_rules CHECK (
    (type = 'league' AND round_robin_legs IS NOT NULL)
    OR (type IN ('qualification_tiebreak', 'single_elimination') AND round_robin_legs IS NULL)
);

-- Mixed tournaments created before this rule already have league at position 1
-- and elimination at position 2. Reserve position 2 for the conditional
-- tiebreak phase without changing a bracket that has already started.
UPDATE tournament_stages AS stage
SET position = 3
FROM tournaments AS tournament
WHERE stage.tournament_id = tournament.id
  AND tournament.format = 'league_then_single_elimination'
  AND stage.type = 'single_elimination'
  AND stage.position = 2;

INSERT INTO tournament_stages (tournament_id, position, type, state)
SELECT tournament.id, 2, 'qualification_tiebreak',
    CASE
        WHEN tournament.state = 'cancelled' THEN 'cancelled'
        WHEN elimination.state = 'pending' THEN 'pending'
        ELSE 'completed'
    END
FROM tournaments AS tournament
JOIN tournament_stages AS elimination
  ON elimination.tournament_id = tournament.id
 AND elimination.type = 'single_elimination'
 AND elimination.position = 3
WHERE tournament.format = 'league_then_single_elimination';

CREATE TABLE tournament_tiebreak_pools (
    tournament_id uuid NOT NULL,
    stage_id uuid NOT NULL,
    pool_number integer NOT NULL CHECK (pool_number > 0),
    source_group_number integer NOT NULL DEFAULT 0 CHECK (source_group_number >= 0),
    qualifier_count smallint NOT NULL CHECK (qualifier_count > 0),
    current_cycle integer NOT NULL DEFAULT 1 CHECK (current_cycle > 0),
    state text NOT NULL DEFAULT 'in_progress' CHECK (state IN ('in_progress', 'completed')),
    PRIMARY KEY (stage_id, pool_number),
    FOREIGN KEY (tournament_id, stage_id)
        REFERENCES tournament_stages (tournament_id, id) ON DELETE CASCADE
);

GRANT SELECT, INSERT, UPDATE, DELETE ON tournament_tiebreak_pools TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: a persisted qualification tiebreak is sporting history.
