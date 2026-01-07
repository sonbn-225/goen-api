-- +goose Up

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE transaction_type AS ENUM ('expense','income','transfer');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE transaction_status AS ENUM ('posted','voided');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- Transactions
CREATE TABLE IF NOT EXISTS transactions (
  id text PRIMARY KEY,
  client_id text,
  external_ref text,

  type transaction_type NOT NULL,
  occurred_at timestamptz NOT NULL,
  amount numeric(18,2) NOT NULL,
  currency text,

  description text,
  account_id text,
  from_account_id text,
  to_account_id text,
  exchange_rate numeric(18,8),

  counterparty text,
  notes text,
  status transaction_status NOT NULL DEFAULT 'posted',

  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  created_by text,
  updated_by text,
  deleted_at timestamptz,

  CONSTRAINT fk_transactions_account_id FOREIGN KEY (account_id) REFERENCES accounts(id),
  CONSTRAINT fk_transactions_from_account_id FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_transactions_to_account_id FOREIGN KEY (to_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_transactions_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_transactions_updated_by FOREIGN KEY (updated_by) REFERENCES users(id)
);

-- Idempotency (expense/income via account_id + external_ref)
CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_account_external_ref
  ON transactions(account_id, external_ref)
  WHERE external_ref IS NOT NULL AND account_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_from_account_id ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account_id ON transactions(to_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at DESC);

-- Line items (split)
CREATE TABLE IF NOT EXISTS transaction_line_items (
  id text PRIMARY KEY,
  transaction_id text NOT NULL,
  category_id text,
  amount numeric(18,2) NOT NULL,
  note text,

  CONSTRAINT fk_tli_transaction FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tli_transaction_id ON transaction_line_items(transaction_id);

-- +goose Down
DROP TABLE IF EXISTS transaction_line_items;
DROP TABLE IF EXISTS transactions;

DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
