-- +goose Up

-- Extend security_event_type enum to store listing/registration events (e.g., NLIS).
-- +goose StatementBegin
DO $$
BEGIN
  ALTER TYPE security_event_type ADD VALUE IF NOT EXISTS 'listing';
EXCEPTION
  WHEN undefined_object THEN
    -- In case enum doesn't exist yet (fresh DB), it will be created by the baseline migration.
    NULL;
END $$;
-- +goose StatementEnd

-- +goose Down

-- NOTE: Postgres cannot easily remove enum values.
-- We can recreate the enum without 'listing', but this will fail if any rows use 'listing'.
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'security_event_type') THEN
    ALTER TYPE security_event_type RENAME TO security_event_type_old;

    CREATE TYPE security_event_type AS ENUM (
      'dividend_cash','split','reverse_split','rights_issue','bonus_issue','additional_issue'
    );

    ALTER TABLE security_events
      ALTER COLUMN event_type TYPE security_event_type
      USING event_type::text::security_event_type;

    DROP TYPE security_event_type_old;
  END IF;
END $$;
-- +goose StatementEnd
