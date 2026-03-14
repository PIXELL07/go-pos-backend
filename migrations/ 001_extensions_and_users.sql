-- Description: Enable UUID extension and create users + auth tables
-- Run order: 1

-- Extensions 
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- ILIKE trigram indexes

-- ENUM types 
DO $$ BEGIN
  CREATE TYPE user_role AS ENUM ('admin', 'owner', 'biller');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- users 
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(255) UNIQUE,
    mobile        VARCHAR(20)  UNIQUE,
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    role          user_role    NOT NULL DEFAULT 'biller',
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    google_id     VARCHAR(255) UNIQUE,
    avatar_url    TEXT,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email      ON users (email)     WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_mobile     ON users (mobile)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_role       ON users (role)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- refresh_tokens 
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id   ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token     ON refresh_tokens (token)     WHERE is_revoked = FALSE;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires   ON refresh_tokens (expires_at);
