-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

-- CHECK accepts SQL NULL. Require a configuration explicitly for set sports.
ALTER TABLE tournaments DROP CONSTRAINT tournaments_set_format_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_set_format_check CHECK (
    (sport = 'tennis' AND best_of_sets IS NOT NULL AND best_of_sets IN (3, 5))
    OR (sport = 'padel' AND best_of_sets IS NOT NULL AND best_of_sets = 3)
    OR (sport IN ('football', 'basketball', 'handball') AND best_of_sets IS NULL)
);

-- +goose Down
-- No rollback: preserve the validated configuration of public tournaments.
