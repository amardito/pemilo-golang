-- +migrate Up
-- Create login_attempts table for rate limiting
CREATE TABLE IF NOT EXISTS login_attempts (
    id VARCHAR(36) PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,  -- username or IP
    attempt_at TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_login_attempts_identifier ON login_attempts(identifier);
CREATE INDEX idx_login_attempts_attempt_at ON login_attempts(attempt_at);
CREATE INDEX idx_login_attempts_identifier_success ON login_attempts(identifier, success);

-- +migrate Down
DROP TABLE IF EXISTS login_attempts;
