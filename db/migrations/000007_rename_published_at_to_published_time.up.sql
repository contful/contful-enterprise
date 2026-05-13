-- Rename entries.published_at to entries.published_time for naming consistency
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'entries' AND column_name = 'published_at'
    ) THEN
        ALTER TABLE entries RENAME COLUMN published_at TO published_time;
    END IF;
END $$;
