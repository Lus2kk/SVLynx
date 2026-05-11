DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'channel_type') THEN
        CREATE TYPE channel_type AS ENUM ('public', 'private');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'channel_role') THEN
        CREATE TYPE channel_role AS ENUM ('owner', 'admin', 'editor', 'member');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS channels (
    id           UUID         NOT NULL,
    name         TEXT         NOT NULL,
    handle       TEXT         NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    avatar_url   TEXT         NOT NULL DEFAULT '',
    avatar_color TEXT         NOT NULL DEFAULT '',
    type         channel_type NOT NULL DEFAULT 'public',
    owner_id     UUID         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT channels_handle_unique UNIQUE (handle)
);

CREATE INDEX IF NOT EXISTS idx_channels_owner_id ON channels (owner_id);
CREATE INDEX IF NOT EXISTS idx_channels_type     ON channels (type);
CREATE INDEX IF NOT EXISTS idx_channels_name     ON channels (name);
CREATE INDEX IF NOT EXISTS idx_channels_handle   ON channels (handle);

CREATE TABLE IF NOT EXISTS channel_members (
    channel_id  UUID         NOT NULL,
    user_id     UUID         NOT NULL,
    role        channel_role NOT NULL DEFAULT 'member',
    joined_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    invited_by  UUID,
    custom_name TEXT         NOT NULL DEFAULT '',
    PRIMARY KEY (channel_id, user_id),
    CONSTRAINT fk_channel_members_channel
        FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_channel_members_user_id    ON channel_members (user_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_channel_id ON channel_members (channel_id);