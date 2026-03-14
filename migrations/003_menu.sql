-- Description: Categories and Menu Items
-- Run order: 3

-- categories
CREATE TABLE IF NOT EXISTS categories (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(150) NOT NULL,
    description TEXT,
    outlet_id   UUID         NOT NULL REFERENCES outlets (id) ON DELETE CASCADE,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order  SMALLINT     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_categories_outlet_id  ON categories (outlet_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_is_active  ON categories (is_active)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories (sort_order);


CREATE TABLE IF NOT EXISTS menu_items (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(200)   NOT NULL,
    description     TEXT,
    category_id     UUID           NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    outlet_id       UUID           NOT NULL REFERENCES outlets   (id) ON DELETE CASCADE,
    price           NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    tax_rate        NUMERIC(5, 2)  NOT NULL DEFAULT 5.00,
    is_veg          BOOLEAN        NOT NULL DEFAULT TRUE,
    is_available    BOOLEAN        NOT NULL DEFAULT TRUE,
    is_online_active BOOLEAN       NOT NULL DEFAULT FALSE,
    image_url       TEXT,
    sort_order      SMALLINT       NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_menu_items_outlet_id      ON menu_items (outlet_id)       WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_menu_items_category_id    ON menu_items (category_id)     WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_menu_items_is_available   ON menu_items (is_available)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_menu_items_is_online      ON menu_items (is_online_active) WHERE deleted_at IS NULL;

-- Trigram index for fast ILIKE search on item names
CREATE INDEX IF NOT EXISTS idx_menu_items_name_trgm      ON menu_items USING gin (name gin_trgm_ops);
