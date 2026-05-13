-- Revert: rename entries.published_time back to entries.published_at
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'entries' AND column_name = 'published_time'
    ) THEN
        ALTER TABLE entries RENAME COLUMN published_time TO published_at;
    END IF;
END $$;
