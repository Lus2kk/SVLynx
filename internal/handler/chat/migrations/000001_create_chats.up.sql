CREATE TABLE IF NOT EXISTS chats (
    id UUID  PRIMARY KEY,  
    type VARCHAR(20) NOT NULL CHECK (type IN ('direct', 'group', 'channel')),
    name VARCHAR(50) ,
    avatar_url TEXT ,
    owner_id UUID  NOT NULL,
    creation_time TIMESTAMPTZ DEFAULT NOW()
);