CREATE TABLE channel_invite_links (
    id         UUID        PRIMARY KEY,
    channel_id UUID        NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    created_by UUID        NOT NULL,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ,
    max_uses   INTEGER,
    use_count  INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active     BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_channel_invite_links_token      ON channel_invite_links (token);
CREATE INDEX idx_channel_invite_links_channel_id ON channel_invite_links (channel_id);