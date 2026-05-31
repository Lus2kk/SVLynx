DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_message_type') THEN
        CREATE TYPE group_message_type AS ENUM ('text', 'voice', 'image', 'video', 'audio', 'file');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_message_status') THEN
        CREATE TYPE group_message_status AS ENUM ('sent', 'delivered', 'read');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS group_messages (
    id          UUID                NOT NULL,
    group_id    UUID                NOT NULL,
    topic_id    UUID,
    sender_id   UUID                NOT NULL,
    content     TEXT                NOT NULL DEFAULT '',
    type        group_message_type  NOT NULL DEFAULT 'text',
    status      group_message_status NOT NULL DEFAULT 'sent',
    media_url   TEXT,
    media_type  TEXT,
    file_name   TEXT,
    file_size   BIGINT,
    duration    INTEGER,
    transcript  TEXT,
    reply_to_id UUID,
    pinned      BOOLEAN             NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    edited_at   TIMESTAMPTZ,
    PRIMARY KEY (id),
    CONSTRAINT fk_group_messages_group
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
    CONSTRAINT fk_group_messages_topic
        FOREIGN KEY (topic_id) REFERENCES group_topics (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_group_messages_group_created ON group_messages (group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_group_messages_topic_created ON group_messages (topic_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_group_messages_sender_id     ON group_messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_group_messages_pinned        ON group_messages (group_id, pinned);
CREATE INDEX IF NOT EXISTS idx_group_messages_content       ON group_messages (group_id, content);