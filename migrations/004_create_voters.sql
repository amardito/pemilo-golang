-- +goose Up
CREATE TABLE voters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    full_name TEXT NOT NULL,
    nim_raw VARCHAR(50) NOT NULL,
    nim_normalized VARCHAR(50) NOT NULL,
    class_name TEXT,
    status TEXT NOT NULL DEFAULT 'ELIGIBLE'
        CHECK (status IN ('ELIGIBLE','DISABLED')),
    has_voted BOOLEAN NOT NULL DEFAULT false,
    voted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE voters ADD CONSTRAINT uq_voters_event_nim UNIQUE (event_id, nim_normalized);
CREATE INDEX idx_voters_event_voted ON voters(event_id, has_voted);
CREATE INDEX idx_voters_event_voted_at ON voters(event_id, voted_at DESC);

CREATE TABLE voter_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES voters(id) ON DELETE CASCADE UNIQUE,
    token CHAR(8) NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE','USED','REVOKED')),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at TIMESTAMPTZ
);

ALTER TABLE voter_tokens ADD CONSTRAINT uq_voter_tokens_event_token UNIQUE (event_id, token);
CREATE INDEX idx_voter_tokens_event_token ON voter_tokens(event_id, token);
CREATE INDEX idx_voter_tokens_voter ON voter_tokens(voter_id);

-- +goose Down
DROP TABLE IF EXISTS voter_tokens;
DROP TABLE IF EXISTS voters;
