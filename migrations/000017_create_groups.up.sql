DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_type') THEN
        CREATE TYPE group_type AS ENUM ('public', 'private');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_role') THEN
        CREATE TYPE group_role AS ENUM ('creator', 'admin', 'member');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS groups (
    id           UUID        NOT NULL,
    name         TEXT         NOT NULL,
    handle       TEXT         NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    avatar_url   TEXT         NOT NULL DEFAULT '',
    avatar_color TEXT         NOT NULL DEFAULT '',
    type         group_type   NOT NULL DEFAULT 'public',
    creator_id   UUID         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT groups_handle_unique UNIQUE (handle)
);

CREATE INDEX IF NOT EXISTS idx_groups_creator_id ON groups (creator_id);
CREATE INDEX IF NOT EXISTS idx_groups_type       ON groups (type);
CREATE INDEX IF NOT EXISTS idx_groups_name       ON groups (name);
CREATE INDEX IF NOT EXISTS idx_groups_handle     ON groups (handle);