-- +goose Up
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SCHEDULED','OPEN','CLOSED','LOCKED')),
    opens_at TIMESTAMPTZ,
    closes_at TIMESTAMPTZ,
    max_slates INT NOT NULL DEFAULT 2,
    max_voters INT NOT NULL DEFAULT 30,
    package TEXT NOT NULL DEFAULT 'FREE'
        CHECK (package IN ('FREE','STARTER','PRO')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_owner ON events(owner_user_id);
CREATE INDEX idx_events_status ON events(status);

-- +goose Down
DROP TABLE IF EXISTS events;
