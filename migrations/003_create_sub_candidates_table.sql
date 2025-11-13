-- +migrate Up
-- Create sub_candidates table
CREATE TABLE IF NOT EXISTS sub_candidates (
    id VARCHAR(36) PRIMARY KEY,
    candidate_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    photo_url VARCHAR(512) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);

CREATE INDEX idx_sub_candidates_candidate_id ON sub_candidates(candidate_id);

-- +migrate Down
DROP TABLE IF EXISTS sub_candidates;
