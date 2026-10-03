-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT tournaments_sport_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_sport_check
    CHECK (sport IN ('football','basketball','handball','tennis','padel','table_tennis','volleyball'));
ALTER TABLE tournaments DROP CONSTRAINT tournaments_set_format_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_set_format_check CHECK (
    (sport = 'tennis' AND best_of_sets IS NOT NULL AND best_of_sets IN (3,5))
    OR (sport = 'padel' AND best_of_sets IS NOT NULL AND best_of_sets = 3)
    OR (sport = 'table_tennis' AND best_of_sets IS NOT NULL AND best_of_sets IN (3,5,7))
    OR (sport = 'volleyball' AND best_of_sets IS NOT NULL AND best_of_sets = 5)
    OR (sport IN ('football','basketball','handball') AND best_of_sets IS NULL)
);

-- +goose Down
-- No rollback: preserve volleyball tournaments and their public history.
