-- +goose Up
SET ROLE tournaments_manager_dev_schema_owner;

CREATE TABLE google_risc_events (
    id text PRIMARY KEY,
    received_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    CONSTRAINT google_risc_events_expiration CHECK (expires_at > received_at)
);

CREATE INDEX google_risc_events_purge_idx ON google_risc_events (expires_at);

GRANT SELECT, INSERT, DELETE ON google_risc_events TO tournaments_manager_dev_app;

-- +goose Down
-- No rollback: removing deduplication after enabling RISC would permit replay.
