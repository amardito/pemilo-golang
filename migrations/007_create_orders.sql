-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    package TEXT NOT NULL CHECK (package IN ('STARTER','PRO')),
    amount INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING','PAID','EXPIRED','FAILED')),
    ipaymu_reference TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_event ON orders(event_id);
CREATE INDEX idx_orders_ipaymu_ref ON orders(ipaymu_reference);

-- +goose Down
DROP TABLE IF EXISTS orders;
