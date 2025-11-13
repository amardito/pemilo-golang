-- +migrate Up
-- Create votes table
CREATE TABLE IF NOT EXISTS votes (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    candidate_id VARCHAR(36) NOT NULL,
    sub_candidate_id VARCHAR(36),
    voter_identifier VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE,
    FOREIGN KEY (sub_candidate_id) REFERENCES sub_candidates(id) ON DELETE SET NULL,
    UNIQUE (room_id, voter_identifier)
);

CREATE INDEX idx_votes_room_id ON votes(room_id);
CREATE INDEX idx_votes_candidate_id ON votes(candidate_id);
CREATE INDEX idx_votes_voter_identifier ON votes(voter_identifier);
CREATE INDEX idx_votes_created_at ON votes(created_at);

-- +migrate Down
DROP TABLE IF EXISTS votes;
