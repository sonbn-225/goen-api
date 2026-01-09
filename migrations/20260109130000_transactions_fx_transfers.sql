-- +goose Up

-- Transactions: currency is derived from linked accounts.
-- For transfers, allow different amounts per side (FX transfers).

ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS from_amount numeric(18,2),
  ADD COLUMN IF NOT EXISTS to_amount numeric(18,2);

-- Backfill existing transfers (historically same-currency):
UPDATE transactions
SET from_amount = COALESCE(from_amount, amount),
    to_amount   = COALESCE(to_amount, amount)
WHERE type = 'transfer' AND deleted_at IS NULL;

-- Drop redundant currency column.
ALTER TABLE transactions
  DROP COLUMN IF EXISTS currency;

-- +goose Down

-- Re-add currency column (optional). Cannot reliably backfill without account context.
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS currency text;

ALTER TABLE transactions
  DROP COLUMN IF EXISTS from_amount,
  DROP COLUMN IF EXISTS to_amount;
