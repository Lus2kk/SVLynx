CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE,
    nickname        TEXT UNIQUE,
    name            TEXT DEFAULT '',
    status          TEXT,
    avatar_color    TEXT,
    telegram_id     BIGINT UNIQUE,
    username        TEXT UNIQUE,
    first_name      TEXT,
    photo_url       TEXT,
    created_at      TIMESTAMP DEFAULT now()
);