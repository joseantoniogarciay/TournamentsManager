-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT tournaments_sport_check;
ALTER TABLE tournaments ADD COLUMN best_of_sets smallint;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_sport_check
    CHECK (sport IN ('football', 'basketball', 'handball', 'tennis', 'padel'));
ALTER TABLE tournaments ADD CONSTRAINT tournaments_set_format_check CHECK (
    (sport = 'tennis' AND best_of_sets IN (3, 5))
    OR (sport = 'padel' AND best_of_sets = 3)
    OR (sport IN ('football', 'basketball', 'handball') AND best_of_sets IS NULL)
);

CREATE TABLE match_sets (
    match_id uuid NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    set_number smallint NOT NULL CHECK (set_number BETWEEN 1 AND 5),
    home_score smallint NOT NULL CHECK (home_score >= 0),
    away_score smallint NOT NULL CHECK (away_score >= 0),
    PRIMARY KEY (match_id, set_number),
    CONSTRAINT match_sets_valid_tiebreak_set CHECK (
        home_score <> away_score AND (
            (GREATEST(home_score, away_score) = 6 AND LEAST(home_score, away_score) <= 4)
            OR (GREATEST(home_score, away_score) = 7 AND LEAST(home_score, away_score) IN (5, 6))
        )
    )
);

CREATE TABLE match_result_change_sets (
    change_id uuid NOT NULL REFERENCES match_result_changes (id) ON DELETE CASCADE,
    set_number smallint NOT NULL CHECK (set_number BETWEEN 1 AND 5),
    home_score smallint NOT NULL CHECK (home_score >= 0),
    away_score smallint NOT NULL CHECK (away_score >= 0),
    PRIMARY KEY (change_id, set_number),
    CONSTRAINT match_result_change_sets_valid_tiebreak_set CHECK (
        home_score <> away_score AND (
            (GREATEST(home_score, away_score) = 6 AND LEAST(home_score, away_score) <= 4)
            OR (GREATEST(home_score, away_score) = 7 AND LEAST(home_score, away_score) IN (5, 6))
        )
    )
);

GRANT SELECT, INSERT, UPDATE, DELETE ON match_sets TO tournaments_manager_dev_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON match_result_change_sets TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: racket sports and their set history are immutable public data
-- introduced by ADR-0135.
