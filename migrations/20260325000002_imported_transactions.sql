-- +goose Up
-- Staging table for imported transactions (supports v1, v2, and other sources)
-- Users can export transactions from v2 and re-import them with optional mapping

CREATE TABLE IF NOT EXISTS imported_transactions (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  source text NOT NULL DEFAULT 'goen-v1',

  transaction_date date NOT NULL,
  amount numeric(18,2) NOT NULL,
  description text,
  transaction_type text,

  imported_account_name text,
  imported_category_name text,

  mapped_account_id text,
  mapped_category_id text,

  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,

  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_imported_transactions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_imported_transactions_mapped_account_id FOREIGN KEY (mapped_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_imported_transactions_mapped_category_id FOREIGN KEY (mapped_category_id) REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_user_date
  ON imported_transactions(user_id, transaction_date DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_user_mapped
  ON imported_transactions(user_id, mapped_account_id, mapped_category_id);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_raw_payload_gin
  ON imported_transactions USING GIN (raw_payload);

-- +goose Down
DROP TABLE IF EXISTS imported_transactions;
