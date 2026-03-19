-- +goose Up
-- Refactor transaction status model and remove transactions.description column
-- 1) transaction_status: posted/voided -> pending/posted/cancelled
-- 2) migrate existing transactions.description into first line item note
-- 3) drop transactions.description

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_type
    WHERE typname = 'transaction_status'
  ) THEN
    ALTER TYPE transaction_status RENAME TO transaction_status_old;
  END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_type
    WHERE typname = 'transaction_status'
  ) THEN
    CREATE TYPE transaction_status AS ENUM ('pending', 'posted', 'cancelled');
  END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE transactions
  ALTER COLUMN status DROP DEFAULT;

ALTER TABLE transactions
  ALTER COLUMN status TYPE transaction_status
  USING (
    CASE
      WHEN status::text = 'posted' THEN 'posted'::transaction_status
      WHEN status::text = 'voided' THEN 'cancelled'::transaction_status
      ELSE 'pending'::transaction_status
    END
  );

ALTER TABLE transactions
  ALTER COLUMN status SET DEFAULT 'pending';

UPDATE transaction_line_items tli
SET note = tx.description
FROM transactions tx
WHERE tli.transaction_id = tx.id
  AND tx.description IS NOT NULL
  AND btrim(tx.description) <> ''
  AND (tli.note IS NULL OR btrim(tli.note) = '')
  AND tli.id = (
    SELECT li.id
    FROM transaction_line_items li
    WHERE li.transaction_id = tx.id
    ORDER BY li.id
    LIMIT 1
  );

INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
SELECT gen_random_uuid()::text, tx.id, NULL, tx.amount, tx.description
FROM transactions tx
WHERE tx.description IS NOT NULL
  AND btrim(tx.description) <> ''
  AND NOT EXISTS (
    SELECT 1
    FROM transaction_line_items li
    WHERE li.transaction_id = tx.id
  );

ALTER TABLE transactions
  DROP COLUMN IF EXISTS description;

-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_type
    WHERE typname = 'transaction_status_old'
  ) THEN
    DROP TYPE transaction_status_old;
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Irreversible migration:
-- - description data is moved and original column removed
-- - status enum/value remap is one-way in this migration
SELECT 1;
