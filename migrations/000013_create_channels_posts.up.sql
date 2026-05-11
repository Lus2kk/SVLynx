CREATE TABLE   channel_posts (
    id         UUID        NOT NULL,
    channel_id UUID        NOT NULL,
    author_id  UUID        NOT NULL,
    content    TEXT        NOT NULL DEFAULT '',
    media_url  TEXT,
    media_type TEXT,
    file_name  TEXT,
    file_size  BIGINT,
    pinned     BOOLEAN     NOT NULL DEFAULT FALSE,
    view_count INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at  TIMESTAMPTZ,
    PRIMARY KEY (id),
    CONSTRAINT fk_channel_posts_channel
        FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE
);

CREATE INDEX    idx_channel_posts_channel_created ON channel_posts (channel_id, created_at DESC);
CREATE INDEX    idx_channel_posts_author_id       ON channel_posts (author_id);
CREATE INDEX    idx_channel_posts_pinned          ON channel_posts (channel_id, pinned);
CREATE INDEX    idx_channel_posts_content         ON channel_posts (channel_id, content);