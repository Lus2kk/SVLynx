CREATE TABLE IF NOT EXISTS chat_members (
    chat_id     UUID        NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL,
    joined_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(chat_id, user_id)
);