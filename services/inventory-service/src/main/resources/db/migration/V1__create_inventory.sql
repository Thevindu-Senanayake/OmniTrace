-- Inventory Service schema (inventory_db).
-- Stock reservation uses SELECT ... FOR UPDATE against the inventory row,
-- producing genuine PostgreSQL row-lock contention under concurrent load.

CREATE TABLE inventory (
    item_id       VARCHAR(64) PRIMARY KEY,
    stock         INT NOT NULL DEFAULT 0,
    reserved      INT NOT NULL DEFAULT 0,   -- in-flight reservations
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT stock_non_negative CHECK (stock >= 0)
);

-- Append-only audit log of every stock action. Also the idempotency guard:
-- a (order_id, action) pair is processed at most once.
CREATE TABLE inventory_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID NOT NULL,
    item_id       VARCHAR(64) NOT NULL,
    quantity      INT NOT NULL,
    action        VARCHAR(32) NOT NULL,   -- RESERVED | RELEASED | REPLENISHED
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_events_order ON inventory_events(order_id);

-- One RESERVED (and one RELEASED) row per order is allowed at most once,
-- enforcing consumer idempotency at the database level.
CREATE UNIQUE INDEX uq_inventory_events_order_action
    ON inventory_events(order_id, action);
