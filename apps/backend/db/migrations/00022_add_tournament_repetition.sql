-- +goose Up
ALTER TABLE tournaments
 ADD COLUMN repetition_source_id uuid REFERENCES tournaments(id),
 ADD COLUMN repetition_account_id uuid REFERENCES accounts(id) ON DELETE SET NULL,
 ADD COLUMN repetition_request_id uuid,
 ADD COLUMN suggested_start jsonb,
 ADD CONSTRAINT tournament_repetition_identity CHECK (
   (repetition_source_id IS NULL AND repetition_account_id IS NULL AND repetition_request_id IS NULL)
   OR (repetition_source_id IS NOT NULL AND repetition_request_id IS NOT NULL)
 ),
 ADD CONSTRAINT tournament_repetition_unique UNIQUE (repetition_source_id, repetition_account_id, repetition_request_id),
 ADD CONSTRAINT tournament_suggested_start CHECK (suggested_start IS NULL OR (state = 'published' AND jsonb_typeof(suggested_start) = 'object'));

-- +goose Down
ALTER TABLE tournaments DROP CONSTRAINT tournament_suggested_start,
 DROP CONSTRAINT tournament_repetition_unique, DROP CONSTRAINT tournament_repetition_identity,
 DROP COLUMN suggested_start, DROP COLUMN repetition_request_id,
 DROP COLUMN repetition_account_id, DROP COLUMN repetition_source_id;
