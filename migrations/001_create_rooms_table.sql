-- +migrate Up
-- Create rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    voters_type VARCHAR(50) NOT NULL CHECK (voters_type IN ('custom_tickets', 'wild_limited', 'wild_unlimited')),
    voters_limit INTEGER,
    session_start_time TIMESTAMP,
    session_end_time TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('enabled', 'disabled')),
    publish_state VARCHAR(20) NOT NULL CHECK (publish_state IN ('draft', 'published')),
    session_state VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (session_state IN ('open', 'closed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rooms_voters_type ON rooms(voters_type);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_publish_state ON rooms(publish_state);
CREATE INDEX idx_rooms_session_state ON rooms(session_state);

-- +migrate Down
DROP TABLE IF EXISTS rooms;
