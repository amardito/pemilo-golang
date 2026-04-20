CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'event_status') THEN
        CREATE TYPE event_status AS ENUM ('DRAFT', 'SCHEDULED', 'OPEN', 'CLOSED', 'LOCKED');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ballot_type') THEN
        CREATE TYPE ballot_type AS ENUM ('SECRET');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'voter_status') THEN
        CREATE TYPE voter_status AS ENUM ('ELIGIBLE', 'DISABLED');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'token_status') THEN
        CREATE TYPE token_status AS ENUM ('ACTIVE', 'USED', 'REVOKED');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    status event_status NOT NULL DEFAULT 'DRAFT',
    opens_at TIMESTAMPTZ,
    closes_at TIMESTAMPTZ,
    ballot_type ballot_type NOT NULL DEFAULT 'SECRET',
    max_candidates INT NOT NULL DEFAULT 12,
    max_voters INT NOT NULL DEFAULT 1500,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS slates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    number INT NOT NULL,
    name TEXT NOT NULL,
    vision TEXT,
    mission TEXT,
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, number)
);

CREATE TABLE IF NOT EXISTS slate_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slate_id UUID NOT NULL REFERENCES slates(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    full_name TEXT NOT NULL,
    photo_url TEXT,
    bio TEXT,
    sort_order INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS voters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    full_name TEXT NOT NULL,
    nim VARCHAR(50) NOT NULL,
    class_name TEXT,
    status voter_status NOT NULL DEFAULT 'ELIGIBLE',
    has_voted BOOLEAN NOT NULL DEFAULT FALSE,
    voted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, nim)
);

CREATE TABLE IF NOT EXISTS voter_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    token CHAR(8) NOT NULL,
    status token_status NOT NULL DEFAULT 'ACTIVE',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    UNIQUE (voter_id),
    UNIQUE (event_id, token)
);

CREATE TABLE IF NOT EXISTS ballots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    slate_id UUID NOT NULL REFERENCES slates(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    actor_user_id UUID,
    action TEXT NOT NULL,
    meta JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    package_name TEXT NOT NULL,
    amount BIGINT NOT NULL,
    status TEXT NOT NULL,
    payment_ref TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ballots_event_slate ON ballots (event_id, slate_id);
CREATE INDEX IF NOT EXISTS idx_voters_event_has_voted ON voters (event_id, has_voted);
CREATE INDEX IF NOT EXISTS idx_voters_event_voted_at_desc ON voters (event_id, voted_at DESC);
CREATE INDEX IF NOT EXISTS idx_voter_tokens_event_token ON voter_tokens (event_id, token);
