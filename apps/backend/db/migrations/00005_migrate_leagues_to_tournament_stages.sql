-- +goose Up
-- La migración preserva las ligas existentes como torneos con una fase league.
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE leagues RENAME TO tournaments;
ALTER TABLE tournaments DROP CONSTRAINT leagues_format_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_format_check CHECK (format IN ('league', 'single_elimination'));
ALTER TABLE league_administrators RENAME TO tournament_administrators;
ALTER TABLE league_followers RENAME TO tournament_followers;
ALTER TABLE league_teams RENAME TO tournament_teams;
ALTER TABLE league_champions RENAME TO tournament_champions;

ALTER TABLE tournament_administrators RENAME COLUMN league_id TO tournament_id;
ALTER TABLE tournament_followers RENAME COLUMN league_id TO tournament_id;
ALTER TABLE tournament_teams RENAME COLUMN league_id TO tournament_id;
ALTER TABLE tournament_champions RENAME COLUMN league_id TO tournament_id;
ALTER TABLE account_notifications RENAME COLUMN league_id TO tournament_id;
ALTER TABLE account_notifications DROP CONSTRAINT account_notifications_kind_check;
UPDATE account_notifications SET kind=replace(kind,'league_','tournament_');
ALTER TABLE account_notifications ADD CONSTRAINT account_notifications_kind_check
    CHECK (kind IN ('tournament_administrator_assigned','tournament_ownership_transferred'));
ALTER TABLE matches RENAME COLUMN league_id TO tournament_id;

CREATE TABLE tournament_stages (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    tournament_id uuid NOT NULL REFERENCES tournaments (id) ON DELETE CASCADE,
    position integer NOT NULL CHECK (position > 0),
    type text NOT NULL CHECK (type IN ('league', 'single_elimination')),
    state text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'in_progress', 'completed', 'cancelled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    round_robin_legs smallint CHECK (round_robin_legs IN (1, 2)),
    CONSTRAINT tournament_stages_rules CHECK (
        (type = 'league' AND round_robin_legs IS NOT NULL)
        OR (type = 'single_elimination' AND round_robin_legs IS NULL)
    ),
    CONSTRAINT tournament_stages_position_unique UNIQUE (tournament_id, position),
    CONSTRAINT tournament_stages_tournament_id_unique UNIQUE (tournament_id, id)
);

-- Every existing league is a one-stage tournament. The stage lifecycle is
-- derived from the already accepted tournament lifecycle.
INSERT INTO tournament_stages (tournament_id, position, type, state, round_robin_legs)
SELECT id, 1, 'league',
    CASE state
        WHEN 'published' THEN 'pending'
        WHEN 'in_progress' THEN 'in_progress'
        WHEN 'completed' THEN 'completed'
        ELSE 'cancelled'
    END, round_robin_legs
FROM tournaments;

ALTER TABLE matches ADD COLUMN stage_id uuid;
UPDATE matches
SET stage_id = tournament_stages.id
FROM tournament_stages
WHERE tournament_stages.tournament_id = matches.tournament_id
  AND tournament_stages.position = 1;
ALTER TABLE matches ALTER COLUMN stage_id SET NOT NULL;
ALTER TABLE matches
    ADD CONSTRAINT matches_stage_fk
    FOREIGN KEY (tournament_id, stage_id) REFERENCES tournament_stages (tournament_id, id) ON DELETE CASCADE;

ALTER TABLE matches DROP CONSTRAINT matches_round_sequence_unique;
ALTER TABLE matches
    ADD CONSTRAINT matches_stage_round_sequence_unique UNIQUE (stage_id, round_number, sequence),
    ADD CONSTRAINT matches_stage_id_unique UNIQUE (stage_id, id);

ALTER TABLE matches DROP CONSTRAINT matches_distinct_teams;
ALTER TABLE matches ALTER COLUMN home_team_id DROP NOT NULL;
ALTER TABLE matches ALTER COLUMN away_team_id DROP NOT NULL;
ALTER TABLE matches
    ADD CONSTRAINT matches_distinct_known_teams CHECK (
        home_team_id IS NULL OR away_team_id IS NULL OR home_team_id <> away_team_id
    );

-- Sources make a future bracket slot explainable. Existing league matches have
-- direct team sources; bracket matches can instead point at an earlier winner.
ALTER TABLE matches
    ADD COLUMN home_source_kind text NOT NULL DEFAULT 'seeded_team'
        CHECK (home_source_kind IN ('seeded_team', 'winner', 'bye')),
    ADD COLUMN away_source_kind text NOT NULL DEFAULT 'seeded_team'
        CHECK (away_source_kind IN ('seeded_team', 'winner', 'bye')),
    ADD COLUMN home_source_match_id uuid,
    ADD COLUMN away_source_match_id uuid,
    ADD COLUMN winner_team_id uuid,
    ADD COLUMN home_penalties integer,
    ADD COLUMN away_penalties integer;

ALTER TABLE matches
    ADD CONSTRAINT matches_home_source_fk FOREIGN KEY (stage_id, home_source_match_id)
        REFERENCES matches (stage_id, id),
    ADD CONSTRAINT matches_away_source_fk FOREIGN KEY (stage_id, away_source_match_id)
        REFERENCES matches (stage_id, id);

ALTER TABLE matches
    ADD CONSTRAINT matches_source_shape CHECK (
        (home_source_kind = 'winner') = (home_source_match_id IS NOT NULL)
        AND (away_source_kind = 'winner') = (away_source_match_id IS NOT NULL)
        AND (home_source_kind <> 'seeded_team' OR home_team_id IS NOT NULL)
        AND (away_source_kind <> 'seeded_team' OR away_team_id IS NOT NULL)
        AND (home_source_kind <> 'bye' OR home_team_id IS NULL)
        AND (away_source_kind <> 'bye' OR away_team_id IS NULL)
        AND NOT (home_source_kind = 'bye' AND away_source_kind = 'bye')
        AND (home_source_match_id IS NULL OR home_source_match_id <> id)
        AND (away_source_match_id IS NULL OR away_source_match_id <> id)
        AND (home_source_match_id IS NULL OR away_source_match_id IS NULL OR home_source_match_id <> away_source_match_id)
    );

ALTER TABLE matches DROP CONSTRAINT matches_state_check;
ALTER TABLE matches ADD CONSTRAINT matches_state_check CHECK (state IN ('pending', 'completed', 'bye'));
ALTER TABLE matches DROP CONSTRAINT matches_score_lifecycle;
ALTER TABLE matches ADD CONSTRAINT matches_score_lifecycle CHECK (
    (state = 'pending' AND home_score IS NULL AND away_score IS NULL AND winner_team_id IS NULL)
    OR (state = 'completed' AND home_team_id IS NOT NULL AND away_team_id IS NOT NULL
        AND home_score IS NOT NULL AND away_score IS NOT NULL AND home_score >= 0 AND away_score >= 0)
    OR (state = 'bye' AND home_score IS NULL AND away_score IS NULL AND winner_team_id IS NOT NULL
        AND (home_source_kind = 'bye' OR away_source_kind = 'bye'))
);
ALTER TABLE matches ADD CONSTRAINT matches_winner_participates CHECK (
    winner_team_id IS NULL OR COALESCE(winner_team_id = home_team_id OR winner_team_id = away_team_id, false)
);
ALTER TABLE matches ADD CONSTRAINT matches_penalties_shape CHECK (
    (home_penalties IS NULL AND away_penalties IS NULL)
    OR (home_penalties IS NOT NULL AND away_penalties IS NOT NULL AND home_penalties >= 0
        AND away_penalties >= 0 AND home_penalties <> away_penalties AND state = 'completed'
        AND home_score = away_score AND winner_team_id IS NOT NULL)
);

ALTER TABLE match_result_changes
    ADD COLUMN previous_home_penalties integer,
    ADD COLUMN previous_away_penalties integer,
    ADD COLUMN home_penalties integer,
    ADD COLUMN away_penalties integer;

ALTER TABLE matches
    ADD CONSTRAINT matches_winner_team_fk
    FOREIGN KEY (tournament_id, winner_team_id)
    REFERENCES tournament_teams (tournament_id, id) ON DELETE RESTRICT;

GRANT SELECT, INSERT, UPDATE, DELETE ON tournament_stages TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: public resource names and new stage references are intentionally
-- incompatible with the former league-only contract (ADR-0122 and ADR-0123).
