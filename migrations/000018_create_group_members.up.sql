CREATE TABLE IF NOT EXISTS group_members (
    group_id    UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    role        group_role  NOT NULL DEFAULT 'member',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    invited_by  UUID,
    custom_name TEXT        NOT NULL DEFAULT '',
    is_banned   BOOLEAN     NOT NULL DEFAULT FALSE,
    banned_until TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id),
    CONSTRAINT fk_group_members_group
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_members_user_id    ON group_members (user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_group_id   ON group_members (group_id);