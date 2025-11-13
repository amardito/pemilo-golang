-- +migrate Up
-- Create candidates table
CREATE TABLE IF NOT EXISTS candidates (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    photo_url VARCHAR(512) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

CREATE INDEX idx_candidates_room_id ON candidates(room_id);

-- +migrate Down
DROP TABLE IF EXISTS candidates;
