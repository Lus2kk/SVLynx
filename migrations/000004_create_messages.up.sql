CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read');

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL,
    content TEXT,
    status message_status NOT NULL DEFAULT 'sent',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);