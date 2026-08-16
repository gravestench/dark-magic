-- Mutable pre-production Realm schema.
--
-- Until Dark Magic declares its persistence format production-stable, schema
-- changes replace this baseline and development installations are recreated
-- with `make realm-fresh-install`. Immutable forward migrations begin only at
-- that production compatibility boundary.

CREATE TABLE IF NOT EXISTS realm_accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    normalized_name TEXT NOT NULL UNIQUE,
    email TEXT,
    normalized_email TEXT UNIQUE,
    email_verified_at TIMESTAMPTZ,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS realm_sessions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES realm_accounts(id) ON DELETE CASCADE,
    token_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    selected_character_id TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS realm_sessions_expiry_idx ON realm_sessions (expires_at);

CREATE TABLE IF NOT EXISTS realm_characters (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES realm_accounts(id) ON DELETE CASCADE,
    revision BIGINT NOT NULL CHECK (revision > 0),
    character JSONB NOT NULL,
    compatibility JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS realm_characters_account_idx ON realm_characters (account_id, id);

CREATE TABLE IF NOT EXISTS realm_character_leases (
    character_id TEXT PRIMARY KEY REFERENCES realm_characters(id) ON DELETE CASCADE,
    token_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    revision BIGINT NOT NULL CHECK (revision > 0),
    game_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS realm_character_leases_expiry_idx ON realm_character_leases (expires_at);

CREATE TABLE IF NOT EXISTS realm_games (
    id TEXT PRIMARY KEY,
    normalized_name TEXT NOT NULL,
    owner_account_id TEXT NOT NULL REFERENCES realm_accounts(id) ON DELETE CASCADE,
    created_by TEXT NOT NULL,
    state TEXT NOT NULL,
    specification JSONB NOT NULL,
    password_hash BYTEA,
    revision BIGINT NOT NULL CHECK (revision > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS realm_games_active_name_idx
    ON realm_games (normalized_name)
    WHERE state <> 'completed';
CREATE INDEX IF NOT EXISTS realm_games_active_created_idx
    ON realm_games (created_at, id)
    WHERE state <> 'completed';

CREATE TABLE IF NOT EXISTS realm_allocations (
    game_id TEXT PRIMARY KEY REFERENCES realm_games(id) ON DELETE CASCADE,
    allocator_id TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL,
    endpoint JSONB,
    runtime_identity JSONB,
    last_healthy_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS realm_memberships (
    game_id TEXT NOT NULL REFERENCES realm_games(id) ON DELETE CASCADE,
    player_id TEXT NOT NULL,
    account_id TEXT NOT NULL REFERENCES realm_accounts(id) ON DELETE CASCADE,
    character_id TEXT NOT NULL REFERENCES realm_characters(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    baseline JSONB NOT NULL,
    departure_receipt JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, player_id),
    UNIQUE (game_id, account_id)
);

CREATE TABLE IF NOT EXISTS realm_game_reservations (
    token_digest BYTEA PRIMARY KEY CHECK (octet_length(token_digest) = 32),
    game_id TEXT NOT NULL REFERENCES realm_games(id) ON DELETE CASCADE,
    character_id TEXT NOT NULL REFERENCES realm_characters(id) ON DELETE CASCADE,
    player JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (game_id, character_id)
);
CREATE INDEX IF NOT EXISTS realm_game_reservations_game_idx
    ON realm_game_reservations (game_id, created_at);

CREATE TABLE IF NOT EXISTS realm_game_players (
    game_id TEXT NOT NULL REFERENCES realm_games(id) ON DELETE CASCADE,
    character_id TEXT NOT NULL REFERENCES realm_characters(id) ON DELETE CASCADE,
    player JSONB NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, character_id)
);

CREATE TABLE IF NOT EXISTS realm_game_checkpoints (
    game_id TEXT PRIMARY KEY REFERENCES realm_games(id) ON DELETE CASCADE,
    allocation_id TEXT NOT NULL,
    identity_hash TEXT NOT NULL,
    tick BIGINT NOT NULL CHECK (tick >= 0),
    checksum TEXT NOT NULL,
    payload BYTEA NOT NULL CHECK (octet_length(payload) > 0 AND octet_length(payload) <= 33554432),
    payload_digest BYTEA NOT NULL CHECK (octet_length(payload_digest) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS realm_mail_outbox (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    recipient TEXT NOT NULL,
    payload JSONB NOT NULL,
    state TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL,
    locked_by TEXT,
    locked_until TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS realm_mail_outbox_ready_idx ON realm_mail_outbox (state, available_at);

CREATE TABLE IF NOT EXISTS realm_account_challenges (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES realm_accounts(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('verify_email', 'reset_password')),
    token_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS realm_account_challenges_account_idx
    ON realm_account_challenges (account_id, kind, created_at DESC);
CREATE INDEX IF NOT EXISTS realm_account_challenges_expiry_idx
    ON realm_account_challenges (expires_at) WHERE consumed_at IS NULL;

CREATE TABLE IF NOT EXISTS realm_audit_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event JSONB NOT NULL
);
