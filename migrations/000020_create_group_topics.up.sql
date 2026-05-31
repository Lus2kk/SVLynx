CREATE TABLE IF NOT EXISTS group_topics (
    id          UUID        NOT NULL,
    group_id    UUID        NOT NULL,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    is_closed   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_by  UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    PRIMARY KEY (id),
    CONSTRAINT fk_group_topics_group
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_topics_group_id ON group_topics (group_id);