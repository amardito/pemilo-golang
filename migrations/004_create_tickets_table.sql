-- +migrate Up
-- Create tickets table
CREATE TABLE IF NOT EXISTS tickets (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    code VARCHAR(255) NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    UNIQUE (room_id, code)
);

CREATE INDEX idx_tickets_room_id ON tickets(room_id);
CREATE INDEX idx_tickets_code ON tickets(code);
CREATE INDEX idx_tickets_is_used ON tickets(is_used);

-- +migrate Down
DROP TABLE IF EXISTS tickets;
