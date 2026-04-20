-- +goose Up
CREATE TABLE slates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    number INT NOT NULL,
    name TEXT NOT NULL,
    vision TEXT,
    mission TEXT,
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE slates ADD CONSTRAINT uq_slates_event_number UNIQUE (event_id, number);
CREATE INDEX idx_slates_event ON slates(event_id);

CREATE TABLE slate_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slate_id UUID NOT NULL REFERENCES slates(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    full_name TEXT NOT NULL,
    photo_url TEXT,
    bio TEXT,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_slate_members_slate ON slate_members(slate_id);

-- +goose Down
DROP TABLE IF EXISTS slate_members;
DROP TABLE IF EXISTS slates;
