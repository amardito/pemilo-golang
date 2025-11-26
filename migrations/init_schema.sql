-- ============================================
-- Pemilo Golang Database Schema - Complete
-- ============================================
-- This script creates all tables for the pemilo-golang voting system
-- Run this script on a fresh PostgreSQL database

-- ============================================
-- 1. ADMINS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS admins (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    max_room INTEGER NOT NULL DEFAULT 10,
    max_voters INTEGER NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_admins_username ON admins(username);
CREATE INDEX idx_admins_is_active ON admins(is_active);

-- ============================================
-- 2. LOGIN ATTEMPTS TABLE (for rate limiting)
-- ============================================
CREATE TABLE IF NOT EXISTS login_attempts (
    id VARCHAR(36) PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,
    attempt_at TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_login_attempts_identifier ON login_attempts(identifier);
CREATE INDEX idx_login_attempts_attempt_at ON login_attempts(attempt_at);
CREATE INDEX idx_login_attempts_identifier_success ON login_attempts(identifier, success);

-- ============================================
-- 3. ROOMS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(36) PRIMARY KEY,
    admin_id VARCHAR(36),
    name VARCHAR(255) NOT NULL,
    voters_type VARCHAR(50) NOT NULL CHECK (voters_type IN ('custom_tickets', 'wild_limited', 'wild_unlimited')),
    voters_limit INTEGER,
    session_start_time TIMESTAMP,
    session_end_time TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('enabled', 'disabled')),
    publish_state VARCHAR(20) NOT NULL CHECK (publish_state IN ('draft', 'published')),
    session_state VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (session_state IN ('open', 'closed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE
);

CREATE INDEX idx_rooms_admin_id ON rooms(admin_id);
CREATE INDEX idx_rooms_voters_type ON rooms(voters_type);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_publish_state ON rooms(publish_state);
CREATE INDEX idx_rooms_session_state ON rooms(session_state);

-- ============================================
-- 4. CANDIDATES TABLE
-- ============================================
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

-- ============================================
-- 5. SUB CANDIDATES TABLE
-- ============================================
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

-- ============================================
-- 6. TICKETS TABLE
-- ============================================
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

-- ============================================
-- 7. VOTES TABLE
-- ============================================
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

-- ============================================
-- SCHEMA CREATION COMPLETE
-- ============================================
