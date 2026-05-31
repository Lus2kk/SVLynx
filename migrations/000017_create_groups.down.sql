DROP TABLE IF EXISTS groups;

DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_role') THEN
        DROP TYPE group_role;
    END IF;
END $$;

DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_type') THEN
        DROP TYPE group_type;
    END IF;
END $$;