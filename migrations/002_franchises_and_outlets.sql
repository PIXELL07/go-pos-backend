-- Description: Franchise, Outlet, OutletAccess, Zone, Table
-- Run order: 2

-- ENUM types
DO $$ BEGIN
  CREATE TYPE outlet_type AS ENUM ('dine_in', 'takeaway', 'delivery', 'cloud');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- franchises 
CREATE TABLE IF NOT EXISTS franchises (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- outlets 
CREATE TABLE IF NOT EXISTS outlets (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(150) NOT NULL,
    ref_id       VARCHAR(50)  NOT NULL UNIQUE,
    type         outlet_type  NOT NULL DEFAULT 'dine_in',
    address      TEXT,
    city         VARCHAR(100),
    state        VARCHAR(100),
    pin_code     VARCHAR(20),
    phone        VARCHAR(20),
    gst_number   VARCHAR(30),
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    is_locked    BOOLEAN      NOT NULL DEFAULT FALSE,
    franchise_id UUID         REFERENCES franchises (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outlets_ref_id      ON outlets (ref_id)       WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlets_franchise    ON outlets (franchise_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlets_is_active    ON outlets (is_active)    WHERE deleted_at IS NULL;

-- outlet_accesses 
CREATE TABLE IF NOT EXISTS outlet_accesses (
    id         UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID      NOT NULL REFERENCES users   (id) ON DELETE CASCADE,
    outlet_id  UUID      NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    role       user_role NOT NULL DEFAULT 'biller',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, outlet_id)
);

CREATE INDEX IF NOT EXISTS idx_outlet_accesses_user   ON outlet_accesses (user_id)   WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlet_accesses_outlet ON outlet_accesses (outlet_id) WHERE deleted_at IS NULL;

-- ─── zones ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS zones (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    outlet_id  UUID         NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_zones_outlet_id ON zones (outlet_id) WHERE deleted_at IS NULL;

-- tables 
CREATE TABLE IF NOT EXISTS tables (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50)  NOT NULL,
    zone_id     UUID         NOT NULL REFERENCES zones (id) ON DELETE CASCADE,
    capacity    SMALLINT     NOT NULL DEFAULT 4,
    is_occupied BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tables_zone_id ON tables (zone_id) WHERE deleted_at IS NULL;
