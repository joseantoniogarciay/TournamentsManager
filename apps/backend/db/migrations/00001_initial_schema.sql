-- +goose Up

CREATE TABLE accounts (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    email text NOT NULL,
    locale text NOT NULL CHECK (locale IN ('es', 'en', 'it', 'fr')),
    state text NOT NULL CHECK (state IN ('pending_verification', 'verified')),
    username text NOT NULL UNIQUE CHECK (username = lower(username)),
    created_at timestamptz NOT NULL DEFAULT now(),
    verified_at timestamptz,
    expires_at timestamptz,
    CONSTRAINT accounts_state_lifecycle CHECK (
        (state = 'pending_verification'
            AND verified_at IS NULL
            AND expires_at IS NOT NULL)
        OR
        (state = 'verified'
            AND verified_at IS NOT NULL
            AND expires_at IS NULL)
    )
);

CREATE UNIQUE INDEX accounts_email_lookup_unique_idx ON accounts (lower(email));

CREATE INDEX accounts_pending_expiration_idx
    ON accounts (expires_at)
    WHERE state = 'pending_verification';

CREATE TABLE local_credentials (
    account_id uuid PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE email_verification_tokens (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CONSTRAINT email_verification_tokens_expiration CHECK (expires_at > created_at),
    CONSTRAINT email_verification_tokens_consumption CHECK (
        consumed_at IS NULL OR consumed_at >= created_at
    ),
    CONSTRAINT email_verification_tokens_invalidation CHECK (
        invalidated_at IS NULL OR invalidated_at >= created_at
    ),
    CONSTRAINT email_verification_tokens_terminal_state CHECK (
        consumed_at IS NULL OR invalidated_at IS NULL
    )
);

CREATE UNIQUE INDEX email_verification_tokens_one_active_per_account_idx
    ON email_verification_tokens (account_id)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE INDEX email_verification_tokens_purge_idx
    ON email_verification_tokens (expires_at)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE TABLE password_reset_tokens (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CONSTRAINT password_reset_tokens_expiration CHECK (expires_at > created_at),
    CONSTRAINT password_reset_tokens_terminal_state CHECK (
        consumed_at IS NULL OR invalidated_at IS NULL
    )
);

CREATE UNIQUE INDEX password_reset_tokens_one_active_per_account_idx
    ON password_reset_tokens (account_id)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT sessions_expiration CHECK (
        idle_expires_at > created_at
        AND absolute_expires_at > created_at
        AND idle_expires_at <= absolute_expires_at
    ),
    CONSTRAINT sessions_last_seen CHECK (last_seen_at >= created_at),
    CONSTRAINT sessions_revocation CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX sessions_account_idx ON sessions (account_id);

CREATE TABLE session_refresh_tokens (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    session_id uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    CONSTRAINT session_refresh_tokens_expiration CHECK (expires_at > created_at),
    CONSTRAINT session_refresh_tokens_consumption CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CONSTRAINT session_refresh_tokens_revocation CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX session_refresh_tokens_session_idx ON session_refresh_tokens (session_id);

CREATE TABLE league_drafts (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL UNIQUE REFERENCES accounts (id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(btrim(name)) > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    CONSTRAINT league_drafts_expiration CHECK (expires_at > created_at),
    CONSTRAINT league_drafts_updated_at CHECK (updated_at >= created_at)
);

CREATE INDEX league_drafts_purge_idx ON league_drafts (expires_at);

CREATE TABLE draft_teams (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    draft_id uuid NOT NULL REFERENCES league_drafts (id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(btrim(name)) > 0),
    name_normalized text NOT NULL CHECK (length(btrim(name_normalized)) > 0),
    position integer NOT NULL CHECK (position > 0),
    CONSTRAINT draft_teams_name_unique UNIQUE (draft_id, name_normalized),
    CONSTRAINT draft_teams_position_unique UNIQUE (draft_id, position)
);

CREATE TABLE leagues (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    organizer_account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (length(btrim(name)) > 0),
    sport text NOT NULL DEFAULT 'football' CHECK (sport = 'football'),
    format text NOT NULL DEFAULT 'league' CHECK (format = 'league'),
    state text NOT NULL DEFAULT 'draft' CHECK (
        state IN ('draft', 'published', 'in_progress', 'completed', 'cancelled')
    ),
    round_robin_legs smallint NOT NULL DEFAULT 1 CHECK (round_robin_legs = 1),
    points_for_win smallint NOT NULL DEFAULT 3 CHECK (points_for_win = 3),
    points_for_draw smallint NOT NULL DEFAULT 1 CHECK (points_for_draw = 1),
    points_for_loss smallint NOT NULL DEFAULT 0 CHECK (points_for_loss = 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    CONSTRAINT leagues_published_at CHECK (
        (state = 'draft' AND published_at IS NULL)
        OR (state <> 'draft' AND published_at IS NOT NULL)
    )
);

CREATE INDEX leagues_organizer_idx ON leagues (organizer_account_id, created_at DESC);

CREATE TABLE league_administrators (
    league_id uuid NOT NULL REFERENCES leagues (id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    assigned_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (league_id, account_id)
);

CREATE INDEX league_administrators_account_idx
    ON league_administrators (account_id, league_id DESC);

CREATE TABLE league_followers (
    league_id uuid NOT NULL REFERENCES leagues (id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    followed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (league_id, account_id)
);

CREATE INDEX league_followers_account_idx
    ON league_followers (account_id, league_id DESC);

CREATE TABLE league_teams (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    league_id uuid NOT NULL REFERENCES leagues (id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(btrim(name)) > 0),
    name_normalized text NOT NULL CHECK (length(btrim(name_normalized)) > 0),
    position integer NOT NULL CHECK (position > 0),
    CONSTRAINT league_teams_name_unique UNIQUE (league_id, name_normalized),
    CONSTRAINT league_teams_position_unique UNIQUE (league_id, position),
    CONSTRAINT league_teams_league_id_id_unique UNIQUE (league_id, id)
);

CREATE TABLE matches (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    league_id uuid NOT NULL REFERENCES leagues (id) ON DELETE CASCADE,
    round_number integer NOT NULL CHECK (round_number > 0),
    sequence integer NOT NULL CHECK (sequence > 0),
    home_team_id uuid NOT NULL,
    away_team_id uuid NOT NULL,
    state text NOT NULL DEFAULT 'pending' CHECK (state = 'pending'),
    CONSTRAINT matches_distinct_teams CHECK (home_team_id <> away_team_id),
    CONSTRAINT matches_round_sequence_unique UNIQUE (league_id, round_number, sequence),
    CONSTRAINT matches_home_team_fk
        FOREIGN KEY (league_id, home_team_id)
        REFERENCES league_teams (league_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT matches_away_team_fk
        FOREIGN KEY (league_id, away_team_id)
        REFERENCES league_teams (league_id, id)
        ON DELETE RESTRICT
);

CREATE UNIQUE INDEX matches_one_pair_per_league_idx
    ON matches (league_id, LEAST(home_team_id, away_team_id), GREATEST(home_team_id, away_team_id));

CREATE TABLE external_identities (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider = 'google'),
    issuer text NOT NULL CHECK (issuer = 'https://accounts.google.com'),
    subject text NOT NULL CHECK (length(subject) BETWEEN 1 AND 255),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT external_identities_issuer_subject_unique UNIQUE (issuer, subject),
    CONSTRAINT external_identities_account_provider_unique UNIQUE (account_id, provider)
);

CREATE TABLE federated_login_challenges (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    provider text NOT NULL CHECK (provider = 'google'),
    nonce_hash bytea NOT NULL UNIQUE CHECK (octet_length(nonce_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT federated_login_challenges_expiration CHECK (expires_at > created_at),
    CONSTRAINT federated_login_challenges_consumption CHECK (
        consumed_at IS NULL OR consumed_at >= created_at
    )
);

CREATE INDEX federated_login_challenges_purge_idx
    ON federated_login_challenges (expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE identity_link_attempts (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    candidate_account_id uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider = 'google'),
    issuer text NOT NULL CHECK (issuer = 'https://accounts.google.com'),
    subject text NOT NULL CHECK (length(subject) BETWEEN 1 AND 255),
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CONSTRAINT identity_link_attempts_expiration CHECK (expires_at > created_at),
    CONSTRAINT identity_link_attempts_consumption CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CONSTRAINT identity_link_attempts_invalidation CHECK (invalidated_at IS NULL OR invalidated_at >= created_at),
    CONSTRAINT identity_link_attempts_terminal_state CHECK (consumed_at IS NULL OR invalidated_at IS NULL)
);

CREATE UNIQUE INDEX identity_link_attempts_one_active_subject_idx
    ON identity_link_attempts (issuer, subject)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

-- +goose Down

DROP INDEX identity_link_attempts_one_active_subject_idx;
DROP TABLE identity_link_attempts;
DROP INDEX federated_login_challenges_purge_idx;
DROP TABLE federated_login_challenges;
DROP TABLE external_identities;
DROP INDEX matches_one_pair_per_league_idx;
DROP TABLE matches;
DROP TABLE league_teams;
DROP INDEX league_followers_account_idx;
DROP TABLE league_followers;
DROP INDEX league_administrators_account_idx;
DROP TABLE league_administrators;
DROP TABLE leagues;
DROP TABLE draft_teams;
DROP TABLE league_drafts;
DROP INDEX session_refresh_tokens_session_idx;
DROP TABLE session_refresh_tokens;
DROP TABLE sessions;
DROP INDEX password_reset_tokens_one_active_per_account_idx;
DROP TABLE password_reset_tokens;
DROP TABLE email_verification_tokens;
DROP TABLE local_credentials;
DROP INDEX accounts_email_lookup_unique_idx;
DROP TABLE accounts;
