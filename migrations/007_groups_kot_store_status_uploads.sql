-- Description: User groups, KOT, store status snapshots, uploads, outlet franchise_type column
-- Run order: 7


ALTER TABLE outlets
  ADD COLUMN IF NOT EXISTS franchise_type VARCHAR(10) DEFAULT NULL;

--

CREATE TABLE IF NOT EXISTS user_groups (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(150) NOT NULL,
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('admin','biller')),
    outlet_id  UUID         NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_user_groups_outlet  ON user_groups (outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_groups_type    ON user_groups (type)      WHERE deleted_at IS NULL;

--

CREATE TABLE IF NOT EXISTS user_group_members (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID        NOT NULL REFERENCES user_groups (id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    biller_type VARCHAR(30),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (group_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_ugm_group_id ON user_group_members (group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ugm_user_id  ON user_group_members (user_id)  WHERE deleted_at IS NULL;

-- offline_until on menu_items 

ALTER TABLE menu_items
  ADD COLUMN IF NOT EXISTS offline_until TIMESTAMPTZ DEFAULT NULL;

-- kots 

CREATE TABLE IF NOT EXISTS kots (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID        NOT NULL REFERENCES orders  (id) ON DELETE CASCADE,
    outlet_id  UUID        NOT NULL REFERENCES outlets (id),
    kot_number VARCHAR(60) NOT NULL UNIQUE,
    status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    notes      TEXT,
    printed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_kots_order_id  ON kots (order_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kots_outlet_id ON kots (outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kots_status    ON kots (status)    WHERE deleted_at IS NULL;

-- kot_items 

CREATE TABLE IF NOT EXISTS kot_items (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    kot_id       UUID        NOT NULL REFERENCES kots (id) ON DELETE CASCADE,
    menu_item_id UUID        REFERENCES menu_items (id) ON DELETE SET NULL,
    name         VARCHAR(200) NOT NULL,
    quantity     SMALLINT    NOT NULL CHECK (quantity > 0),
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_kot_items_kot_id ON kot_items (kot_id) WHERE deleted_at IS NULL;

--

CREATE TABLE IF NOT EXISTS store_status_snapshots (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id     UUID        NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    platform      VARCHAR(50) NOT NULL,
    is_online     BOOLEAN     NOT NULL DEFAULT TRUE,
    offline_since TIMESTAMPTZ,
    last_checked  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    details       JSONB       NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    UNIQUE (outlet_id, platform)
);
CREATE INDEX IF NOT EXISTS idx_store_status_outlet   ON store_status_snapshots (outlet_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_store_status_platform ON store_status_snapshots (platform);
CREATE INDEX IF NOT EXISTS idx_store_status_online   ON store_status_snapshots (is_online)  WHERE deleted_at IS NULL;

-- uploads
CREATE TABLE IF NOT EXISTS uploads (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID,
    owner_type VARCHAR(30),
    file_name  VARCHAR(255) NOT NULL,
    mime_type  VARCHAR(100) NOT NULL,
    size_bytes BIGINT,
    url        TEXT        NOT NULL,
    store_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_uploads_owner ON uploads (owner_id, owner_type) WHERE deleted_at IS NULL;

-- fcm_sent column on notifications 

ALTER TABLE notifications
  ADD COLUMN IF NOT EXISTS fcm_sent BOOLEAN NOT NULL DEFAULT FALSE;
