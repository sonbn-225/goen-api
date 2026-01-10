-- +goose Up
ALTER TABLE investment_accounts
  DROP COLUMN IF EXISTS broker_name,
  DROP COLUMN IF EXISTS sync_enabled,
  DROP COLUMN IF EXISTS sync_settings,
  ADD COLUMN IF NOT EXISTS fee_settings jsonb,
  ADD COLUMN IF NOT EXISTS tax_settings jsonb;

-- +goose Down
ALTER TABLE investment_accounts
  DROP COLUMN IF EXISTS fee_settings,
  DROP COLUMN IF EXISTS tax_settings,
  ADD COLUMN IF NOT EXISTS broker_name varchar,
  ADD COLUMN IF NOT EXISTS sync_enabled boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS sync_settings jsonb;
