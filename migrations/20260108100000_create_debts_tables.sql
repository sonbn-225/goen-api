-- +goose Up

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE debt_direction AS ENUM ('borrowed','lent');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE debt_status AS ENUM ('active','overdue','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE debt_interest_rule AS ENUM ('interest_first','principal_first');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE debt_installment_status AS ENUM ('pending','paid','overdue');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- Debts
CREATE TABLE IF NOT EXISTS debts (
  id text PRIMARY KEY,
  client_id text,
  user_id text NOT NULL,
  direction debt_direction NOT NULL,
  name text,
  principal numeric(18,2) NOT NULL,
  currency text NOT NULL,
  start_date date NOT NULL,
  due_date date NOT NULL,
  interest_rate numeric(18,8),
  interest_rule debt_interest_rule,
  outstanding_principal numeric(18,2) NOT NULL,
  accrued_interest numeric(18,2) NOT NULL DEFAULT 0,
  status debt_status NOT NULL DEFAULT 'active',
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_debts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT ck_debts_due_after_start CHECK (due_date >= start_date),
  CONSTRAINT ck_debts_outstanding_nonneg CHECK (outstanding_principal >= 0),
  CONSTRAINT ck_debts_accrued_nonneg CHECK (accrued_interest >= 0)
);

CREATE INDEX IF NOT EXISTS idx_debts_user_status ON debts(user_id, status);
CREATE INDEX IF NOT EXISTS idx_debts_user_due_date ON debts(user_id, due_date);

-- Debt payment links (Transaction ↔ Debt)
CREATE TABLE IF NOT EXISTS debt_payment_links (
  id text PRIMARY KEY,
  debt_id text NOT NULL,
  transaction_id text NOT NULL,
  principal_paid numeric(18,2),
  interest_paid numeric(18,2),
  created_at timestamptz NOT NULL,

  CONSTRAINT fk_debt_payment_links_debt_id FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
  CONSTRAINT fk_debt_payment_links_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_debt_payment_links_transaction_id
  ON debt_payment_links(transaction_id);

CREATE INDEX IF NOT EXISTS idx_debt_payment_links_debt_id
  ON debt_payment_links(debt_id, created_at DESC);

-- Debt installment schedule
CREATE TABLE IF NOT EXISTS debt_installments (
  id text PRIMARY KEY,
  debt_id text NOT NULL,
  installment_no int NOT NULL,
  due_date date NOT NULL,
  amount_due numeric(18,2) NOT NULL,
  amount_paid numeric(18,2) NOT NULL DEFAULT 0,
  status debt_installment_status NOT NULL DEFAULT 'pending',

  CONSTRAINT fk_debt_installments_debt_id FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
  CONSTRAINT ck_debt_installments_installment_no CHECK (installment_no > 0),
  CONSTRAINT ck_debt_installments_amounts CHECK (amount_due >= 0 AND amount_paid >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_debt_installments_debt_no
  ON debt_installments(debt_id, installment_no);

CREATE INDEX IF NOT EXISTS idx_debt_installments_debt_due_date
  ON debt_installments(debt_id, due_date);

-- +goose Down
DROP TABLE IF EXISTS debt_installments;
DROP TABLE IF EXISTS debt_payment_links;
DROP TABLE IF EXISTS debts;

DROP TYPE IF EXISTS debt_installment_status;
DROP TYPE IF EXISTS debt_interest_rule;
DROP TYPE IF EXISTS debt_status;
DROP TYPE IF EXISTS debt_direction;
