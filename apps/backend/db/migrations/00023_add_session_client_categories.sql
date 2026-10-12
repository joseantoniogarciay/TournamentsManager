-- +goose Up
ALTER TABLE sessions
    ADD COLUMN client_platform text NOT NULL DEFAULT 'unknown'
        CHECK (client_platform IN ('unknown', 'ios', 'android', 'windows', 'macos', 'linux')),
    ADD COLUMN client_kind text NOT NULL DEFAULT 'unknown'
        CHECK (client_kind IN ('unknown', 'app', 'chrome', 'safari', 'firefox', 'edge', 'web'));

-- +goose Down
ALTER TABLE sessions DROP COLUMN client_kind, DROP COLUMN client_platform;
