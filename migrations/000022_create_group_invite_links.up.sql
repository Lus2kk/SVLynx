CREATE TABLE IF NOT EXISTS group_invite_links (
    id         UUID        PRIMARY KEY,
    group_id   UUID        NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    created_by UUID        NOT NULL,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ,
    max_uses   INTEGER,
    use_count  INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active     BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_group_invite_links_token     ON group_invite_links (token);
CREATE INDEX IF NOT EXISTS idx_group_invite_links_group_id  ON group_invite_links (group_id);