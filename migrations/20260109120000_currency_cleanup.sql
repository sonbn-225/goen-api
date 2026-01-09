-- +goose Up

-- Budgets: currency is derived from user settings (no per-budget currency)
ALTER TABLE budgets
  DROP COLUMN IF EXISTS currency;

-- Debts: currency is derived from the referenced account
ALTER TABLE debts
  ADD COLUMN IF NOT EXISTS account_id text;

-- Add FK constraint if missing
-- +goose StatementBegin
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_debts_account_id'
  ) THEN
    ALTER TABLE debts
      ADD CONSTRAINT fk_debts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id);
  END IF;
END $$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_debts_account_id ON debts(account_id);

ALTER TABLE debts
  DROP COLUMN IF EXISTS currency;

-- Rotating savings: currency is derived from the referenced account
ALTER TABLE rotating_savings_groups
  DROP COLUMN IF EXISTS currency;

-- After removing currency, a group must always reference an account.
ALTER TABLE rotating_savings_groups
  ALTER COLUMN account_id SET NOT NULL;

-- Investment accounts: currency is derived from the referenced account
ALTER TABLE investment_accounts
  DROP COLUMN IF EXISTS currency;

-- +goose Down

ALTER TABLE investment_accounts
  ADD COLUMN IF NOT EXISTS currency varchar;

ALTER TABLE rotating_savings_groups
  ALTER COLUMN account_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS currency text;

ALTER TABLE debts
  ADD COLUMN IF NOT EXISTS currency text;

-- +goose StatementBegin
DO $$ BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_debts_account_id'
  ) THEN
    ALTER TABLE debts
      DROP CONSTRAINT fk_debts_account_id;
  END IF;
END $$;
-- +goose StatementEnd

DROP INDEX IF EXISTS idx_debts_account_id;

ALTER TABLE debts
  DROP COLUMN IF EXISTS account_id;

ALTER TABLE budgets
  ADD COLUMN IF NOT EXISTS currency text;
