-- +goose Up

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE savings_instrument_status AS ENUM ('active','matured','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE rotating_savings_cycle_frequency AS ENUM ('weekly','monthly','custom');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE rotating_savings_group_status AS ENUM ('active','completed','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE rotating_savings_contribution_kind AS ENUM ('contribution','payout');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- Savings instruments (1-1 extension of savings Account)
CREATE TABLE IF NOT EXISTS savings_instruments (
  id text PRIMARY KEY,
  savings_account_id text NOT NULL UNIQUE,
  parent_account_id text NOT NULL,
  principal numeric(18,2) NOT NULL,
  interest_rate numeric(18,8),
  term_months int,
  start_date date,
  maturity_date date,
  auto_renew boolean NOT NULL DEFAULT false,
  accrued_interest numeric(18,2) NOT NULL DEFAULT 0,
  status savings_instrument_status NOT NULL DEFAULT 'active',
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_savings_instruments_savings_account FOREIGN KEY (savings_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_savings_instruments_parent_account FOREIGN KEY (parent_account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_savings_instruments_parent_account_id ON savings_instruments(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_savings_instruments_status ON savings_instruments(status);

-- Rotating savings (hụi/họ)
CREATE TABLE IF NOT EXISTS rotating_savings_groups (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  self_label text,
  account_id text,
  name text NOT NULL,
  currency text NOT NULL,
  member_count int NOT NULL,
  contribution_amount numeric(18,2) NOT NULL,
  early_payout_fee_rate numeric(18,8),
  cycle_frequency rotating_savings_cycle_frequency NOT NULL,
  start_date date NOT NULL,
  status rotating_savings_group_status NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_rotating_savings_groups_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_rotating_savings_groups_account_id FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_groups_user_status ON rotating_savings_groups(user_id, status);

CREATE TABLE IF NOT EXISTS rotating_savings_contributions (
  id text PRIMARY KEY,
  group_id text NOT NULL,
  transaction_id text NOT NULL,
  kind rotating_savings_contribution_kind NOT NULL,
  cycle_no int,
  due_date date,
  amount numeric(18,2) NOT NULL,
  occurred_at timestamptz NOT NULL,
  note text,
  created_at timestamptz NOT NULL,

  CONSTRAINT fk_rotating_savings_contributions_group_id FOREIGN KEY (group_id) REFERENCES rotating_savings_groups(id) ON DELETE CASCADE,
  CONSTRAINT fk_rotating_savings_contributions_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rotating_savings_contributions_transaction_id
  ON rotating_savings_contributions(transaction_id);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_contributions_group_occurred_at
  ON rotating_savings_contributions(group_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_contributions_group_cycle_no
  ON rotating_savings_contributions(group_id, cycle_no);

-- +goose Down
DROP TABLE IF EXISTS rotating_savings_contributions;
DROP TABLE IF EXISTS rotating_savings_groups;
DROP TABLE IF EXISTS savings_instruments;

DROP TYPE IF EXISTS rotating_savings_contribution_kind;
DROP TYPE IF EXISTS rotating_savings_group_status;
DROP TYPE IF EXISTS rotating_savings_cycle_frequency;
DROP TYPE IF EXISTS savings_instrument_status;
