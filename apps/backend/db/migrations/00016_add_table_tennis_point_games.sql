-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT tournaments_sport_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_sport_check
    CHECK (sport IN ('football','basketball','handball','tennis','padel','table_tennis'));
ALTER TABLE tournaments DROP CONSTRAINT tournaments_set_format_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_set_format_check CHECK (
    (sport = 'tennis' AND best_of_sets IS NOT NULL AND best_of_sets IN (3,5))
    OR (sport = 'padel' AND best_of_sets IS NOT NULL AND best_of_sets = 3)
    OR (sport = 'table_tennis' AND best_of_sets IS NOT NULL AND best_of_sets IN (3,5,7))
    OR (sport IN ('football','basketball','handball') AND best_of_sets IS NULL)
);

-- Sport-specific final scores are validated by the domain; both point games
-- and tennis games share nonnegative smallint scores and ordered history.
ALTER TABLE match_sets DROP CONSTRAINT match_sets_valid_tiebreak_set;
ALTER TABLE match_sets DROP CONSTRAINT match_sets_set_number_check;
ALTER TABLE match_sets ADD CONSTRAINT match_sets_set_number_check CHECK (set_number BETWEEN 1 AND 7);
ALTER TABLE match_sets ADD CONSTRAINT match_sets_decisive_score CHECK (home_score <> away_score);
ALTER TABLE match_result_change_sets DROP CONSTRAINT match_result_change_sets_valid_tiebreak_set;
ALTER TABLE match_result_change_sets DROP CONSTRAINT match_result_change_sets_set_number_check;
ALTER TABLE match_result_change_sets ADD CONSTRAINT match_result_change_sets_set_number_check CHECK (set_number BETWEEN 1 AND 7);
ALTER TABLE match_result_change_sets ADD CONSTRAINT match_result_change_sets_decisive_score CHECK (home_score <> away_score);

-- +goose Down
-- No rollback: preserve table tennis tournaments and their public history.
