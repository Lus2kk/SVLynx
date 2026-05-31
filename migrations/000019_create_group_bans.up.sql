CREATE TABLE IF NOT EXISTS group_bans (
    group_id  UUID        NOT NULL,
    user_id   UUID        NOT NULL,
    banned_by UUID        NOT NULL,
    reason    TEXT        NOT NULL DEFAULT '',
    banned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    until     TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id),
    CONSTRAINT fk_group_bans_group
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_bans_group_id ON group_bans (group_id);
CREATE INDEX IF NOT EXISTS idx_group_bans_user_id  ON group_bans (user_id);