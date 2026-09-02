DROP INDEX IF EXISTS idx_venue_last_verified;

ALTER TABLE venue DROP CONSTRAINT IF EXISTS venue_data_source_valid;

ALTER TABLE venue
    DROP COLUMN IF EXISTS last_verified_at,
    DROP COLUMN IF EXISTS data_source,
    DROP COLUMN IF EXISTS website;
