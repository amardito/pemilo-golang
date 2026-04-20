-- +goose Up
CREATE TABLE ballots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    slate_id UUID NOT NULL REFERENCES slates(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- NOTE: NO voter_id column — secret ballot by design
CREATE INDEX idx_ballots_event ON ballots(event_id);
CREATE INDEX idx_ballots_event_slate ON ballots(event_id, slate_id);

-- +goose Down
DROP TABLE IF EXISTS ballots;
