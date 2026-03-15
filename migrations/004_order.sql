-- Description: Orders, Order Items, Payments
-- Run order: 4

-- ENUM types

DO $$ BEGIN
  CREATE TYPE order_source AS ENUM (
    'pos', 'zomato', 'swiggy', 'foodpanda', 'uber_eats', 'dunzo', 'website'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE order_type AS ENUM ('dine_in', 'takeaway', 'delivery');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE order_status AS ENUM (
    'pending', 'accepted', 'preparing', 'ready',
    'dispatched', 'delivered', 'completed',
    'cancelled', 'complimentary', 'sales_return'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE payment_method AS ENUM ('cash', 'card', 'online', 'wallet', 'due', 'other');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- orders

CREATE TABLE IF NOT EXISTS orders (
    id                UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number    VARCHAR(60)    NOT NULL UNIQUE,
    outlet_id         UUID           NOT NULL REFERENCES outlets (id),
    table_id          UUID           REFERENCES tables (id) ON DELETE SET NULL,
    cashier_id        UUID           NOT NULL REFERENCES users  (id),
    source            order_source   NOT NULL DEFAULT 'pos',
    type              order_type     NOT NULL DEFAULT 'dine_in',
    status            order_status   NOT NULL DEFAULT 'pending',
    customer_name     VARCHAR(150),
    customer_phone    VARCHAR(20),
    customer_otp      VARCHAR(10),
    rider_details     TEXT,
    pax               SMALLINT       NOT NULL DEFAULT 1,

    -- Financials

    sub_total         NUMERIC(12, 2) NOT NULL DEFAULT 0,
    discount_amount   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    discount_percent  NUMERIC(5, 2)  NOT NULL DEFAULT 0,
    tax_amount        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    delivery_charge   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    container_charge  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    service_charge    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    additional_charge NUMERIC(10, 2) NOT NULL DEFAULT 0,
    round_off         NUMERIC(6, 2)  NOT NULL DEFAULT 0,
    waived_off        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    total_amount      NUMERIC(12, 2) NOT NULL DEFAULT 0,
    net_sales         NUMERIC(12, 2) NOT NULL DEFAULT 0,
    online_tax_calc   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    gst_by_merchant   NUMERIC(12, 2) NOT NULL DEFAULT 0,
    gst_by_ecommerce  NUMERIC(12, 2) NOT NULL DEFAULT 0,

    -- Meta

    is_modified       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_printed        BOOLEAN        NOT NULL DEFAULT FALSE,
    print_count       SMALLINT       NOT NULL DEFAULT 0,
    external_order_id VARCHAR(100),
    notes             TEXT,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

-- Covering indexes for the most common query patterns
CREATE INDEX IF NOT EXISTS idx_orders_outlet_created  ON orders (outlet_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_status          ON orders (status)                      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_source          ON orders (source)                      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_cashier          ON orders (cashier_id)                  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_invoice          ON orders (invoice_number);
CREATE INDEX IF NOT EXISTS idx_orders_external_id      ON orders (external_order_id)           WHERE external_order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_created_at       ON orders (created_at DESC);

-- initialized partial index for running orders
CREATE INDEX IF NOT EXISTS idx_orders_running ON orders (outlet_id, created_at ASC)
    WHERE status IN ('pending','accepted','preparing','ready','dispatched') AND deleted_at IS NULL;

-- order_items

CREATE TABLE IF NOT EXISTS order_items (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID           NOT NULL REFERENCES orders     (id) ON DELETE CASCADE,
    menu_item_id UUID           NOT NULL REFERENCES menu_items (id),
    name         VARCHAR(200)   NOT NULL,
    quantity     SMALLINT       NOT NULL CHECK (quantity > 0),
    unit_price   NUMERIC(10, 2) NOT NULL,
    tax_rate     NUMERIC(5, 2)  NOT NULL DEFAULT 0,
    tax_amount   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    total_price  NUMERIC(12, 2) NOT NULL,
    notes        TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id     ON order_items (order_id)     WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_order_items_menu_item_id ON order_items (menu_item_id);

-- payments 
CREATE TABLE IF NOT EXISTS payments (
    id         UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID           NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    method     payment_method NOT NULL,
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    ref_no     VARCHAR(100),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_method   ON payments (method);
