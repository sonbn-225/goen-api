-- +goose Up
-- +goose NO TRANSACTION
-- Baseline schema (squashed)
-- This migration is intended for fresh installs / DB resets.

-- TimescaleDB
-- The Postgres image in this workspace preloads TimescaleDB.
-- This makes the extension available to CREATE EXTENSION here.
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE account_type AS ENUM ('bank','wallet','cash','broker','card','savings');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE account_status AS ENUM ('active','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE user_account_permission AS ENUM ('owner','viewer','editor');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE user_account_status AS ENUM ('active','revoked');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE transaction_type AS ENUM ('expense','income','transfer');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE transaction_status AS ENUM ('pending','posted','cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE budget_period AS ENUM ('month','week','custom');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE budget_rollover_mode AS ENUM ('reset','carry_forward','accumulate');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE savings_instrument_status AS ENUM ('active','matured','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE rotating_savings_cycle_frequency AS ENUM ('weekly','monthly','custom');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE rotating_savings_group_status AS ENUM ('active','completed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE rotating_savings_contribution_kind AS ENUM (
  'uncollected',
  'payout',
  'collected',
  'partial_collected'
);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE debt_direction AS ENUM ('borrowed','lent');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE debt_status AS ENUM ('active','overdue','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE debt_interest_rule AS ENUM ('interest_first','principal_first');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE debt_installment_status AS ENUM ('pending','paid','overdue');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE security_asset_class AS ENUM ('stock','fund','crypto','bond','other');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE security_event_type AS ENUM ('cash_dividend','stock_dividend','split','reverse_split','rights_issue','bonus_issue','additional_issue','listing');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE security_event_election_status AS ENUM ('draft','confirmed','cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE trade_side AS ENUM ('buy','sell');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  -- Includes the later enum evolution to support lot-based holdings.
  CREATE TYPE holding_source_of_truth AS ENUM ('trades','sync','lots');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lot_provenance') THEN
    CREATE TYPE lot_provenance AS ENUM ('regular_buy','stock_dividend','rights_offering');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lot_status') THEN
    CREATE TYPE lot_status AS ENUM ('active', 'adjusted');
  END IF;
END $$;

-- +goose StatementEnd


-- Tables
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  username text NOT NULL UNIQUE,
  email text NULL,
  phone text NULL,
  display_name text NULL,
  avatar_url text NULL,
  status text NOT NULL DEFAULT 'active',
  password_hash text NOT NULL,
  settings jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,
  CONSTRAINT users_email_or_phone_chk CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uq ON users (lower(email)) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_uq ON users (phone) WHERE phone IS NOT NULL;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  token text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_refresh_tokens_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS contacts (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  name text NOT NULL,
  email text,
  phone text,
  avatar_url text, -- Local override or synced from linked user
  linked_user_id uuid,
  notes text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_contacts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_contacts_linked_user_id FOREIGN KEY (linked_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts(user_id);
CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(lower(email));
CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone);
CREATE INDEX IF NOT EXISTS idx_contacts_linked_user_id ON contacts(linked_user_id);

CREATE TABLE IF NOT EXISTS accounts (
  id uuid PRIMARY KEY,
  name varchar NOT NULL,
  account_type account_type NOT NULL,
  currency varchar NOT NULL,
  parent_account_id uuid,
  status account_status NOT NULL DEFAULT 'active',
  account_number text,
  color varchar,
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_accounts_parent FOREIGN KEY (parent_account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_accounts_parent_account_id ON accounts(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_accounts_color ON accounts(color);

CREATE TABLE IF NOT EXISTS user_accounts (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL,
  user_id uuid NOT NULL,
  permission user_account_permission NOT NULL,
  status user_account_status NOT NULL DEFAULT 'active',
  revoked_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_user_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT uq_user_accounts_account_user UNIQUE (account_id, user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_accounts_active_owner
  ON user_accounts(account_id)
  WHERE permission = 'owner' AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_user_accounts_user_id ON user_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_accounts_account_id ON user_accounts(account_id);

CREATE TABLE IF NOT EXISTS transactions (
  id uuid PRIMARY KEY,
  external_ref text,

  type transaction_type NOT NULL,
  occurred_at timestamptz NOT NULL,
  amount numeric(18,2) NOT NULL,
  from_amount numeric(18,2),
  to_amount numeric(18,2),

  account_id uuid,
  from_account_id uuid,
  to_account_id uuid,
  exchange_rate numeric(18,8),
  status transaction_status NOT NULL DEFAULT 'pending',

  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_transactions_account_id FOREIGN KEY (account_id) REFERENCES accounts(id),
  CONSTRAINT fk_transactions_from_account_id FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_transactions_to_account_id FOREIGN KEY (to_account_id) REFERENCES accounts(id),
  CONSTRAINT ck_transactions_amount_positive CHECK (amount > 0),
  CONSTRAINT ck_transactions_from_amount_positive CHECK (from_amount IS NULL OR from_amount > 0),
  CONSTRAINT ck_transactions_to_amount_positive CHECK (to_amount IS NULL OR to_amount > 0),
  CONSTRAINT ck_transactions_type_linkage CHECK (
    (
      type IN ('expense', 'income')
      AND account_id IS NOT NULL
      AND from_account_id IS NULL
      AND to_account_id IS NULL
      AND from_amount IS NULL
      AND to_amount IS NULL
      AND exchange_rate IS NULL
    )
    OR
    (
      type = 'transfer'
      AND account_id IS NULL
      AND from_account_id IS NOT NULL
      AND to_account_id IS NOT NULL
      AND from_account_id <> to_account_id
      AND ((from_amount IS NULL AND to_amount IS NULL) OR (from_amount IS NOT NULL AND to_amount IS NOT NULL))
    )
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_account_external_ref
  ON transactions(account_id, external_ref)
  WHERE external_ref IS NOT NULL AND account_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_from_account_id ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account_id ON transactions(to_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at DESC);

-- Optimized for cursor pagination and date-range scans.
-- Matches ORDER BY t.occurred_at DESC, t.id DESC and cursor predicate:
--   AND (t.occurred_at, t.id) < ($time, $id)
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at_id
  ON transactions(occurred_at DESC, id DESC)
  WHERE deleted_at IS NULL;

-- Optimized for account-scoped lists where the query matches one of:
--   t.account_id = $x OR t.from_account_id = $x OR t.to_account_id = $x
CREATE INDEX IF NOT EXISTS idx_transactions_account_occurred_at_id
  ON transactions(account_id, occurred_at DESC, id DESC)
  WHERE deleted_at IS NULL AND account_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_from_account_occurred_at_id
  ON transactions(from_account_id, occurred_at DESC, id DESC)
  WHERE deleted_at IS NULL AND from_account_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_to_account_occurred_at_id
  ON transactions(to_account_id, occurred_at DESC, id DESC)
  WHERE deleted_at IS NULL AND to_account_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS transaction_line_items (
  id uuid PRIMARY KEY,
  transaction_id uuid NOT NULL,
  category_id uuid,
  amount numeric(18,2) NOT NULL,
  note text,

  CONSTRAINT fk_tli_transaction FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  CONSTRAINT ck_tli_amount_positive CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_tli_transaction_id ON transaction_line_items(transaction_id);

-- Prevent attaching line items to transfer transactions.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_tli_non_transfer()
RETURNS trigger AS $$
DECLARE
  tx_type transaction_type;
BEGIN
  SELECT type INTO tx_type
  FROM transactions
  WHERE id = NEW.transaction_id;

  IF tx_type = 'transfer' THEN
    RAISE EXCEPTION 'line items are not allowed for transfer transactions';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_enforce_tli_non_transfer ON transaction_line_items;
CREATE TRIGGER trg_enforce_tli_non_transfer
BEFORE INSERT OR UPDATE OF transaction_id
ON transaction_line_items
FOR EACH ROW
EXECUTE FUNCTION enforce_tli_non_transfer();

CREATE TABLE IF NOT EXISTS group_expense_participants (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  transaction_id uuid NOT NULL,
  participant_name text NOT NULL,
  original_amount numeric(18,2) NOT NULL,
  share_amount numeric(18,2) NOT NULL,
  is_settled boolean NOT NULL DEFAULT false,
  settlement_transaction_id uuid,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_gep_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_gep_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  CONSTRAINT fk_gep_settlement_transaction_id FOREIGN KEY (settlement_transaction_id) REFERENCES transactions(id) ON DELETE SET NULL,
  CONSTRAINT ck_gep_amounts_positive CHECK (original_amount > 0 AND share_amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_gep_username ON group_expense_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_gep_transaction_id ON group_expense_participants(transaction_id);
CREATE INDEX IF NOT EXISTS idx_gep_user_settled ON group_expense_participants(user_id, is_settled);
CREATE INDEX IF NOT EXISTS idx_gep_user_participant_name ON group_expense_participants(user_id, lower(participant_name));

CREATE TABLE IF NOT EXISTS categories (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  key text NOT NULL UNIQUE,
  parent_category_id uuid,
  type text,
  sort_order int,
  is_active boolean NOT NULL DEFAULT true,
  icon text,
  color text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_categories_parent FOREIGN KEY (parent_category_id) REFERENCES categories(id),
  CONSTRAINT chk_categories_type CHECK (type IS NULL OR type IN ('expense','income','both'))
);

CREATE INDEX IF NOT EXISTS idx_categories_parent_category_id ON categories(parent_category_id);

CREATE INDEX IF NOT EXISTS idx_tli_category_id ON transaction_line_items(category_id);

-- Helps category-based reporting that filters by category then joins to transactions.
CREATE INDEX IF NOT EXISTS idx_tli_category_transaction_id
  ON transaction_line_items(category_id, transaction_id);

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE transaction_line_items
    ADD CONSTRAINT fk_tli_category FOREIGN KEY (category_id) REFERENCES categories(id);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

-- +goose StatementEnd


CREATE TABLE IF NOT EXISTS tags (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  name_vi text,
  name_en text,
  color text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_tags_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE,
  CONSTRAINT ck_tags_names_present CHECK (name_vi IS NOT NULL OR name_en IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tags_user_name_vi ON tags(user_id, lower(name_vi)) WHERE name_vi IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_tags_user_name_en ON tags(user_id, lower(name_en)) WHERE name_en IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

CREATE TABLE IF NOT EXISTS transaction_tags (
  transaction_id uuid NOT NULL,
  tag_id uuid NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT pk_transaction_tags PRIMARY KEY (transaction_id, tag_id),
  CONSTRAINT fk_transaction_tags_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  CONSTRAINT fk_transaction_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transaction_tags_transaction_id ON transaction_tags(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_tags_tag_id ON transaction_tags(tag_id);

CREATE TABLE IF NOT EXISTS transaction_line_item_tags (
  line_item_id uuid NOT NULL,
  tag_id uuid NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT pk_transaction_line_item_tags PRIMARY KEY (line_item_id, tag_id),
  CONSTRAINT fk_tlit_line_item FOREIGN KEY (line_item_id) REFERENCES transaction_line_items(id) ON DELETE CASCADE,
  CONSTRAINT fk_tlit_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transaction_line_item_tags_line_item_id ON transaction_line_item_tags(line_item_id);
CREATE INDEX IF NOT EXISTS idx_transaction_line_item_tags_tag_id ON transaction_line_item_tags(tag_id);

CREATE TABLE IF NOT EXISTS budgets (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  name text,
  period budget_period NOT NULL,
  period_start date,
  period_end date,
  amount numeric(18,2) NOT NULL,
  alert_threshold_percent int,
  rollover_mode budget_rollover_mode,
  category_id uuid,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_budgets_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE,
  CONSTRAINT fk_budgets_category_id FOREIGN KEY (category_id) REFERENCES categories(id),
  CONSTRAINT chk_budgets_alert_threshold CHECK (alert_threshold_percent IS NULL OR (alert_threshold_percent >= 0 AND alert_threshold_percent <= 100))
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_period ON budgets(user_id, period, period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_budgets_user_category ON budgets(user_id, category_id);

CREATE TABLE IF NOT EXISTS savings_instruments (
  id uuid PRIMARY KEY,
  savings_account_id uuid NOT NULL UNIQUE,
  parent_account_id uuid NOT NULL,
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
  deleted_at timestamptz,

  CONSTRAINT fk_savings_instruments_savings_account FOREIGN KEY (savings_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_savings_instruments_parent_account FOREIGN KEY (parent_account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_savings_instruments_parent_account_id ON savings_instruments(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_savings_instruments_status ON savings_instruments(status);

CREATE TABLE IF NOT EXISTS rotating_savings_groups (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  account_id uuid NOT NULL,
  name text NOT NULL,
  member_count int NOT NULL,
  user_slots int NOT NULL,
  contribution_amount numeric(18,2) NOT NULL,
  payout_cycle_no int,
  fixed_interest_amount numeric(18,2),
  cycle_frequency rotating_savings_cycle_frequency NOT NULL,
  start_date date NOT NULL,
  status rotating_savings_group_status NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_rotating_savings_groups_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_rotating_savings_groups_account_id FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_groups_user_status ON rotating_savings_groups(user_id, status);

CREATE TABLE IF NOT EXISTS rotating_savings_contributions (
  id uuid PRIMARY KEY,
  group_id uuid NOT NULL,
  transaction_id uuid NOT NULL,
  kind rotating_savings_contribution_kind NOT NULL,
  cycle_no int,
  due_date date,
  amount numeric(18,2) NOT NULL,
  slots_taken int NOT NULL DEFAULT 0,
  collected_fee_per_slot numeric(18,2) NOT NULL DEFAULT 0,
  occurred_at timestamptz NOT NULL,
  note text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_rotating_savings_contributions_group_id FOREIGN KEY (group_id) REFERENCES rotating_savings_groups(id) ON DELETE CASCADE,
  CONSTRAINT fk_rotating_savings_contributions_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rotating_savings_contributions_transaction_id
  ON rotating_savings_contributions(transaction_id);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_contributions_group_occurred_at
  ON rotating_savings_contributions(group_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_contributions_group_cycle_no
  ON rotating_savings_contributions(group_id, cycle_no);

CREATE TABLE IF NOT EXISTS rotating_savings_audit_logs (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  group_id uuid,
  action text NOT NULL,
  details jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_rotating_savings_audit_logs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_rotating_savings_audit_logs_group_id FOREIGN KEY (group_id) REFERENCES rotating_savings_groups(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_rotating_savings_audit_logs_group_id ON rotating_savings_audit_logs(group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rotating_savings_audit_logs_user_id ON rotating_savings_audit_logs(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS debts (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  account_id uuid,
  direction debt_direction NOT NULL,
  name text,
  principal numeric(18,2) NOT NULL,
  start_date date NOT NULL,
  due_date date NOT NULL,
  interest_rate numeric(18,8),
  interest_rule debt_interest_rule,
  outstanding_principal numeric(18,2) NOT NULL,
  accrued_interest numeric(18,2) NOT NULL DEFAULT 0,
  status debt_status NOT NULL DEFAULT 'active',
  contact_id uuid,
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_debts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_debts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id),
  CONSTRAINT fk_debts_contact_id FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE SET NULL,
  CONSTRAINT ck_debts_due_after_start CHECK (due_date >= start_date),
  CONSTRAINT ck_debts_outstanding_nonneg CHECK (outstanding_principal >= 0),
  CONSTRAINT ck_debts_accrued_nonneg CHECK (accrued_interest >= 0)
);

CREATE INDEX IF NOT EXISTS idx_debts_user_status ON debts(user_id, status);
CREATE INDEX IF NOT EXISTS idx_debts_user_due_date ON debts(user_id, due_date);
CREATE INDEX IF NOT EXISTS idx_debts_account_id ON debts(account_id);

CREATE TABLE IF NOT EXISTS debt_payment_links (
  id uuid PRIMARY KEY,
  debt_id uuid NOT NULL,
  transaction_id uuid NOT NULL,
  principal_paid numeric(18,2),
  interest_paid numeric(18,2),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_debt_payment_links_debt_id FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
  CONSTRAINT fk_debt_payment_links_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_debt_payment_links_transaction_id
  ON debt_payment_links(transaction_id);

CREATE INDEX IF NOT EXISTS idx_debt_payment_links_debt_id
  ON debt_payment_links(debt_id, created_at DESC);

CREATE TABLE IF NOT EXISTS debt_installments (
  id uuid PRIMARY KEY,
  debt_id uuid NOT NULL,
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

CREATE TABLE IF NOT EXISTS investment_accounts (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL,
  fee_settings jsonb,
  tax_settings jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_investment_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_investment_accounts_account_id
  ON investment_accounts(account_id);

CREATE TABLE IF NOT EXISTS securities (
  id uuid PRIMARY KEY,
  symbol varchar NOT NULL,
  name varchar,
  asset_class security_asset_class,
  currency varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_securities_symbol
  ON securities(symbol);

CREATE TABLE IF NOT EXISTS security_price_dailies (
  id uuid NOT NULL,
  security_id uuid NOT NULL,
  price_date date NOT NULL,
  open numeric(18,8),
  high numeric(18,8),
  low numeric(18,8),
  close numeric(18,8) NOT NULL,
  volume numeric(18,2),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  PRIMARY KEY (id, price_date),
  CONSTRAINT fk_security_price_dailies_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_price_dailies_security_date
  ON security_price_dailies(security_id, price_date);

CREATE INDEX IF NOT EXISTS idx_security_price_dailies_price_date
  ON security_price_dailies(price_date);

CREATE TABLE IF NOT EXISTS security_events (
  id uuid PRIMARY KEY,
  security_id uuid NOT NULL,
  event_type security_event_type NOT NULL,
  ex_date date,
  record_date date,
  pay_date date,
  effective_date date,
  cash_amount_per_share numeric(18,8),
  ratio_numerator numeric(18,8),
  ratio_denominator numeric(18,8),
  subscription_price numeric(18,8),
  currency varchar,
  vnstock_event_id varchar,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_security_events_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_security_events_security_effective_date
  ON security_events(security_id, effective_date);

CREATE INDEX IF NOT EXISTS idx_security_events_security_ex_date
  ON security_events(security_id, ex_date);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_events_security_vnstock_event
  ON security_events(security_id, vnstock_event_id)
  WHERE vnstock_event_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS market_data_sync_states (
  sync_key text PRIMARY KEY,
  min_interval_seconds integer NOT NULL DEFAULT 86400,
  last_started_at timestamptz,
  last_success_at timestamptz,
  last_failure_at timestamptz,
  last_status text NOT NULL DEFAULT 'never',
  last_error text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_market_data_sync_states_last_success_at
  ON market_data_sync_states(last_success_at);

CREATE TABLE IF NOT EXISTS security_event_elections (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  broker_account_id uuid NOT NULL,
  security_event_id uuid NOT NULL,
  security_id uuid NOT NULL,
  entitlement_date date NOT NULL,
  holding_quantity_at_entitlement_date numeric(18,8) NOT NULL,
  entitled_quantity numeric(18,8) NOT NULL,
  elected_quantity numeric(18,8) NOT NULL,
  status security_event_election_status NOT NULL,
  confirmed_at timestamptz,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_security_event_elections_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_security_event_elections_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_security_event_elections_security_event_id FOREIGN KEY (security_event_id) REFERENCES security_events(id) ON DELETE CASCADE,
  CONSTRAINT fk_security_event_elections_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE,
  CONSTRAINT ck_security_event_elections_nonneg CHECK (elected_quantity >= 0 AND entitled_quantity >= 0 AND holding_quantity_at_entitlement_date >= 0),
  CONSTRAINT ck_security_event_elections_elected_le_entitled CHECK (elected_quantity <= entitled_quantity)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_event_elections_broker_event
  ON security_event_elections(broker_account_id, security_event_id);

CREATE INDEX IF NOT EXISTS idx_security_event_elections_user_status
  ON security_event_elections(user_id, status);

CREATE TABLE IF NOT EXISTS trades (
  id uuid PRIMARY KEY,
  broker_account_id uuid NOT NULL,
  security_id uuid NOT NULL,
  fee_transaction_id uuid,
  tax_transaction_id uuid,
  side trade_side NOT NULL,
  quantity numeric(18,8) NOT NULL,
  price numeric(18,8) NOT NULL,
  fees numeric(18,2) NOT NULL DEFAULT 0,
  taxes numeric(18,2) NOT NULL DEFAULT 0,
  occurred_at timestamptz NOT NULL,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_trades_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_trades_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE,
  CONSTRAINT fk_trades_fee_transaction_id FOREIGN KEY (fee_transaction_id) REFERENCES transactions(id) ON DELETE SET NULL,
  CONSTRAINT fk_trades_tax_transaction_id FOREIGN KEY (tax_transaction_id) REFERENCES transactions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_trades_broker_occurred_at
  ON trades(broker_account_id, occurred_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_trades_fee_transaction_id
  ON trades(fee_transaction_id)
  WHERE fee_transaction_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_trades_tax_transaction_id
  ON trades(tax_transaction_id)
  WHERE tax_transaction_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS holdings (
  id uuid PRIMARY KEY,
  broker_account_id uuid NOT NULL,
  security_id uuid NOT NULL,
  quantity numeric(18,8) NOT NULL,
  cost_basis_total numeric(18,2),
  avg_cost numeric(18,8),
  market_price numeric(18,8),
  market_value numeric(18,2),
  unrealized_pnl numeric(18,2),
  as_of timestamptz,
  source_of_truth holding_source_of_truth NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_holdings_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_holdings_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_holdings_broker_security
  ON holdings(broker_account_id, security_id);

CREATE TABLE IF NOT EXISTS share_lots (
  id uuid PRIMARY KEY,
  broker_account_id uuid NOT NULL,
  security_id uuid NOT NULL,
  quantity numeric(18,8) NOT NULL,
  acquisition_date date NOT NULL,
  cost_basis_per_share numeric(18,8) NOT NULL DEFAULT 0,
  provenance lot_provenance NOT NULL DEFAULT 'regular_buy',
  status lot_status NOT NULL DEFAULT 'active',
  buy_trade_id uuid,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_share_lots_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_share_lots_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE,
  CONSTRAINT fk_share_lots_buy_trade_id FOREIGN KEY (buy_trade_id) REFERENCES trades(id) ON DELETE SET NULL,
  CONSTRAINT ck_share_lots_nonneg CHECK (quantity >= 0 AND cost_basis_per_share >= 0)
);

CREATE INDEX IF NOT EXISTS idx_share_lots_broker_security_acq
  ON share_lots(broker_account_id, security_id, acquisition_date ASC);

CREATE INDEX IF NOT EXISTS idx_share_lots_status
  ON share_lots(status);

CREATE TABLE IF NOT EXISTS realized_trade_logs (
  id uuid PRIMARY KEY,
  broker_account_id uuid NOT NULL,
  security_id uuid NOT NULL,
  sell_trade_id uuid NOT NULL,
  source_share_lot_id uuid NOT NULL,
  quantity numeric(18,8) NOT NULL,
  acquisition_date date NOT NULL,
  cost_basis_total numeric(18,2) NOT NULL,
  sell_price numeric(18,8) NOT NULL,
  proceeds numeric(18,2) NOT NULL,
  realized_pnl numeric(18,2) NOT NULL,
  provenance lot_provenance NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_realized_trade_logs_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_sell_trade_id FOREIGN KEY (sell_trade_id) REFERENCES trades(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_source_share_lot_id FOREIGN KEY (source_share_lot_id) REFERENCES share_lots(id) ON DELETE RESTRICT,
  CONSTRAINT ck_realized_trade_logs_nonneg CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_realized_trade_logs_broker_security_sell
  ON realized_trade_logs(broker_account_id, security_id, sell_trade_id);

CREATE TABLE IF NOT EXISTS audit_events (
  id uuid NOT NULL,
  account_id uuid NOT NULL,
  actor_user_id uuid NOT NULL,
  action varchar NOT NULL,
  entity_type varchar NOT NULL,
  entity_id uuid NOT NULL,
  occurred_at timestamptz NOT NULL,
  diff jsonb,

  PRIMARY KEY (id, occurred_at),

  CONSTRAINT fk_audit_events_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_audit_events_actor_user_id FOREIGN KEY (actor_user_id) REFERENCES users(id) ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_id_occurred_at
  ON audit_events(account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_entity
  ON audit_events(account_id, entity_type, entity_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS imported_transactions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  source text NOT NULL DEFAULT 'goen-v1',

  transaction_date date NOT NULL,
  amount numeric(18,2) NOT NULL,
  description text,
  transaction_type text,

  imported_account_name text,
  imported_category_name text,

  mapped_account_id uuid,
  mapped_category_id uuid,

  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,

  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_imported_transactions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_imported_transactions_mapped_account_id FOREIGN KEY (mapped_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_imported_transactions_mapped_category_id FOREIGN KEY (mapped_category_id) REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_user_date
  ON imported_transactions(user_id, transaction_date DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_user_mapped
  ON imported_transactions(user_id, mapped_account_id, mapped_category_id);

CREATE INDEX IF NOT EXISTS idx_imported_transactions_raw_payload_gin
  ON imported_transactions USING GIN (raw_payload);

CREATE TABLE IF NOT EXISTS import_mapping_rules (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  kind text NOT NULL,
  source_name text NOT NULL,
  normalized_source_name text NOT NULL,
  mapped_id uuid NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_import_mapping_rules_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT chk_import_mapping_rules_kind CHECK (kind IN ('account', 'category')),
  CONSTRAINT uq_import_mapping_rules_unique UNIQUE (user_id, kind, normalized_source_name)
);

CREATE INDEX IF NOT EXISTS idx_import_mapping_rules_user_kind
  ON import_mapping_rules(user_id, kind, normalized_source_name);

-- Convert suitable time-series tables to hypertables.
-- Requirements: any UNIQUE/PRIMARY KEY constraints must include the time column.
SELECT public.create_hypertable('security_price_dailies', 'price_date', if_not_exists => TRUE);
SELECT public.create_hypertable('audit_events', 'occurred_at', if_not_exists => TRUE);


-- Seed data (categories) moved to separate migration file.
-- See: 20260316000001_seed_categories.sql





-- +goose Down
-- Intentionally omitted: this migration is a squashed baseline for fresh DBs.
