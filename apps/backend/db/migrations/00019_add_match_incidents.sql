-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

ALTER TABLE matches ADD COLUMN incident jsonb;
ALTER TABLE match_result_changes ADD COLUMN incident jsonb;
ALTER TABLE match_result_changes ADD COLUMN previous_incident jsonb;
ALTER TABLE match_result_changes ADD COLUMN previous_result_type text;

ALTER TABLE matches DROP CONSTRAINT matches_result_type_check;
ALTER TABLE matches ADD CONSTRAINT matches_result_type_check CHECK (
    (state = 'completed' AND result_type IS NOT NULL AND result_type IN ('played','administrative','no_show','retirement'))
    OR (state IN ('pending','bye') AND result_type IS NULL)
);
ALTER TABLE matches ADD CONSTRAINT matches_incident_shape CHECK (
    (result_type IN ('no_show','retirement') AND incident IS NOT NULL AND
        COALESCE(jsonb_typeof(incident)='object' AND incident->>'type'=result_type
            AND incident->>'side' IN ('home','away'),false))
    OR (COALESCE(result_type,'') NOT IN ('no_show','retirement') AND incident IS NULL)
);
ALTER TABLE match_result_changes DROP CONSTRAINT match_result_changes_result_type_check;
ALTER TABLE match_result_changes ADD CONSTRAINT match_result_changes_result_type_check
    CHECK (result_type IN ('played','administrative','no_show','retirement'));
ALTER TABLE match_result_changes ADD CONSTRAINT match_result_changes_incident_shape CHECK (
    (result_type IN ('no_show','retirement') AND incident IS NOT NULL AND
        COALESCE(jsonb_typeof(incident)='object' AND incident->>'type'=result_type
            AND incident->>'side' IN ('home','away'),false))
    OR (result_type NOT IN ('no_show','retirement') AND incident IS NULL)
);
ALTER TABLE match_result_changes ADD CONSTRAINT match_result_changes_previous_result_type_check
    CHECK (previous_result_type IS NULL OR previous_result_type IN ('played','administrative','no_show','retirement'));

-- +goose Down
-- No rollback: preserve incident metadata and immutable public result history.
