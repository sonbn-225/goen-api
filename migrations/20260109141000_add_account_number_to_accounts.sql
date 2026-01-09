-- +goose Up

ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS account_number text;

CREATE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts(account_number);

-- +goose Down

DROP INDEX IF EXISTS idx_accounts_account_number;

ALTER TABLE accounts
  DROP COLUMN IF EXISTS account_number;
