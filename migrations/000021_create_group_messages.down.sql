DROP TABLE IF EXISTS group_messages;

DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_message_status') THEN
        DROP TYPE group_message_status;
    END IF;
END $$;

DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_message_type') THEN
        DROP TYPE group_message_type;
    END IF;
END $$;