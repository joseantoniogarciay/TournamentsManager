-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE tournaments DROP CONSTRAINT tournaments_sport_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_sport_check
    CHECK (sport IN ('football', 'basketball', 'handball'));

-- +goose Down
-- No rollback: sport is immutable tournament data and handball is part of the
-- public contract introduced by ADR-0134.
