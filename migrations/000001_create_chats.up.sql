CREATE TABLE IF NOT EXISTS chats (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    creation_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);