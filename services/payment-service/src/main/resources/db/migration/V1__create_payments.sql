-- Payment Service schema (payments_db). One row per order payment attempt;
-- order_id is UNIQUE so a duplicate Kafka delivery cannot double-charge.

CREATE TABLE payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID NOT NULL UNIQUE,
    customer_id      VARCHAR(64) NOT NULL,
    amount           NUMERIC(10, 2) NOT NULL,
    status           VARCHAR(32) NOT NULL,   -- SUCCESS | FAILED
    failure_reason   VARCHAR(64),            -- null on success
    transaction_id   VARCHAR(64),            -- null on failure
    processing_ms    INT,                    -- how long the payment took
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order    ON payments(order_id);
CREATE INDEX idx_payments_customer ON payments(customer_id);
CREATE INDEX idx_payments_status   ON payments(status);
