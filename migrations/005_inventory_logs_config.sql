-- Description: Pending purchases, audit logs, third-party config, notifications
-- Run order: 5

-- pending_purchases 

CREATE TABLE IF NOT EXISTS pending_purchases (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id    UUID           NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    item_name    VARCHAR(200)   NOT NULL,
    quantity     NUMERIC(10, 3) NOT NULL CHECK (quantity > 0),
    unit         VARCHAR(30)    NOT NULL,
    amount       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    status       VARCHAR(30)    NOT NULL DEFAULT 'pending',
    type         VARCHAR(20)    NOT NULL DEFAULT 'purchase',   -- purchase | transfer
    requested_by UUID           REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pending_purchases_outlet   ON pending_purchases (outlet_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pending_purchases_status   ON pending_purchases (status)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pending_purchases_created  ON pending_purchases (created_at DESC);

-- menu_trigger_logs

CREATE TABLE IF NOT EXISTS menu_trigger_logs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id    UUID        NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    item_id      UUID        REFERENCES menu_items (id) ON DELETE SET NULL,
    action       VARCHAR(80) NOT NULL,
    details      TEXT,
    triggered_by UUID        REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_menu_trigger_logs_outlet  ON menu_trigger_logs (outlet_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_menu_trigger_logs_created ON menu_trigger_logs (created_at DESC);

-- online_store_logs 

CREATE TABLE IF NOT EXISTS online_store_logs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id    UUID        NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    platform     VARCHAR(50) NOT NULL,
    action       VARCHAR(80) NOT NULL,
    status       VARCHAR(30) NOT NULL,
    details      TEXT,
    triggered_by UUID        REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_online_store_logs_outlet   ON online_store_logs (outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_online_store_logs_platform ON online_store_logs (platform);
CREATE INDEX IF NOT EXISTS idx_online_store_logs_created  ON online_store_logs (created_at DESC);

-- online_item_logs

CREATE TABLE IF NOT EXISTS online_item_logs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id    UUID        NOT NULL REFERENCES outlets    (id) ON DELETE CASCADE,
    item_id      UUID        REFERENCES menu_items (id) ON DELETE SET NULL,
    platform     VARCHAR(50) NOT NULL,
    action       VARCHAR(80) NOT NULL,
    old_status   VARCHAR(30),
    new_status   VARCHAR(30),
    triggered_by UUID        REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_online_item_logs_outlet   ON online_item_logs (outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_online_item_logs_platform ON online_item_logs (platform);
CREATE INDEX IF NOT EXISTS idx_online_item_logs_created  ON online_item_logs (created_at DESC);

-- third_party_configs

CREATE TABLE IF NOT EXISTS third_party_configs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id  UUID        NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    platform   VARCHAR(50) NOT NULL,
    api_key    TEXT,
    store_id   VARCHAR(100),
    is_active  BOOLEAN     NOT NULL DEFAULT FALSE,
    config     JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (outlet_id, platform)
);

CREATE INDEX IF NOT EXISTS idx_third_party_outlet   ON third_party_configs (outlet_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_third_party_platform ON third_party_configs (platform);

-- notifications 

CREATE TABLE IF NOT EXISTS notifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title      VARCHAR(200) NOT NULL,
    body       TEXT        NOT NULL,
    type       VARCHAR(50),
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    data       JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id  ON notifications (user_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_is_read  ON notifications (is_read)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_created  ON notifications (created_at DESC);
