-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT leagues_sport_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_sport_check
    CHECK (sport IN ('football', 'basketball'));

ALTER TABLE matches ADD COLUMN result_type text;
UPDATE matches SET result_type = 'played' WHERE state = 'completed';
ALTER TABLE matches ADD CONSTRAINT matches_result_type_check CHECK (
    (state = 'completed' AND result_type IN ('played', 'administrative'))
    OR (state IN ('pending', 'bye') AND result_type IS NULL)
);

ALTER TABLE match_result_changes ADD COLUMN result_type text;
UPDATE match_result_changes SET result_type = 'played';
ALTER TABLE match_result_changes ALTER COLUMN result_type SET NOT NULL;
ALTER TABLE match_result_changes ADD CONSTRAINT match_result_changes_result_type_check
    CHECK (result_type IN ('played', 'administrative'));

-- +goose Down
-- No rollback: sport is part of the public contract and administrative results
-- retain sporting audit meaning introduced by ADR-0126.
