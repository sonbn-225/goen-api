-- +goose Up

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'holding_source_of_truth') THEN
    BEGIN
      ALTER TYPE holding_source_of_truth ADD VALUE IF NOT EXISTS 'lots';
    EXCEPTION
      WHEN duplicate_object THEN
        NULL;
    END;
  END IF;
END $$;

-- +goose Down
-- Note: PostgreSQL does not support removing enum values safely.
