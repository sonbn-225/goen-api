-- +goose Up
-- +goose NO TRANSACTION
-- Baseline schema (squashed)
-- This migration is intended for fresh installs / DB resets.

-- TimescaleDB
-- The Postgres image in this workspace preloads TimescaleDB.
-- This makes the extension available to CREATE EXTENSION here.
CREATE EXTENSION IF NOT EXISTS timescaledb;

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
  CREATE TYPE transaction_status AS ENUM ('posted','voided');
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
  CREATE TYPE rotating_savings_group_status AS ENUM ('active','completed','closed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE rotating_savings_contribution_kind AS ENUM ('contribution','payout');
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
  CREATE TYPE security_event_type AS ENUM ('dividend_cash','split','reverse_split','rights_issue','bonus_issue','additional_issue','listing');
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
    CREATE TYPE lot_provenance AS ENUM ('regular_buy', 'stock_dividend', 'rights_offering');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lot_status') THEN
    CREATE TYPE lot_status AS ENUM ('active', 'adjusted');
  END IF;
END $$;

-- +goose StatementEnd


-- Tables
CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  email text NULL,
  phone text NULL,
  display_name text NULL,
  status text NOT NULL DEFAULT 'active',
  password_hash text NOT NULL,
  settings jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT users_email_or_phone_chk CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uq ON users (lower(email)) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_uq ON users (phone) WHERE phone IS NOT NULL;

CREATE TABLE IF NOT EXISTS accounts (
  id text PRIMARY KEY,
  client_id text,
  name varchar NOT NULL,
  account_type account_type NOT NULL,
  currency varchar NOT NULL,
  parent_account_id text,
  status account_status NOT NULL DEFAULT 'active',
  account_number text,
  color varchar,
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  created_by text,
  updated_by text,
  deleted_at timestamptz,

  CONSTRAINT fk_accounts_parent FOREIGN KEY (parent_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_accounts_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_accounts_updated_by FOREIGN KEY (updated_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_accounts_parent_account_id ON accounts(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_accounts_created_by ON accounts(created_by);
CREATE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_accounts_color ON accounts(color);

CREATE TABLE IF NOT EXISTS user_accounts (
  id text PRIMARY KEY,
  account_id text NOT NULL,
  user_id text NOT NULL,
  permission user_account_permission NOT NULL,
  status user_account_status NOT NULL DEFAULT 'active',
  revoked_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  created_by text,
  updated_by text,

  CONSTRAINT fk_user_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_accounts_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_user_accounts_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  CONSTRAINT uq_user_accounts_account_user UNIQUE (account_id, user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_accounts_active_owner
  ON user_accounts(account_id)
  WHERE permission = 'owner' AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_user_accounts_user_id ON user_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_accounts_account_id ON user_accounts(account_id);

CREATE TABLE IF NOT EXISTS transactions (
  id text PRIMARY KEY,
  client_id text,
  external_ref text,

  type transaction_type NOT NULL,
  occurred_at timestamptz NOT NULL,
  amount numeric(18,2) NOT NULL,
  from_amount numeric(18,2),
  to_amount numeric(18,2),

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
  id text PRIMARY KEY,
  transaction_id text NOT NULL,
  category_id text,
  amount numeric(18,2) NOT NULL,
  note text,

  CONSTRAINT fk_tli_transaction FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tli_transaction_id ON transaction_line_items(transaction_id);

CREATE TABLE IF NOT EXISTS categories (
  id text PRIMARY KEY,
  name text NOT NULL,
  parent_category_id text,
  type text,
  sort_order int,
  is_active boolean NOT NULL DEFAULT true,
  is_system boolean NOT NULL DEFAULT false,
  icon text,
  color text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_categories_parent FOREIGN KEY (parent_category_id) REFERENCES categories(id),
  CONSTRAINT chk_categories_type CHECK (type IS NULL OR type IN ('expense','income','both'))
);

CREATE INDEX IF NOT EXISTS idx_categories_parent_category_id ON categories(parent_category_id);
CREATE INDEX IF NOT EXISTS idx_categories_is_system ON categories(is_system);

CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_name_parent_type
  ON categories (lower(name), parent_category_id, COALESCE(type, ''))
  WHERE deleted_at IS NULL;

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
  id text PRIMARY KEY,
  user_id text NOT NULL,
  name text NOT NULL,
  color text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_tags_user_id FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tags_user_name
  ON tags(user_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

CREATE TABLE IF NOT EXISTS transaction_tags (
  transaction_id text NOT NULL,
  tag_id text NOT NULL,
  created_at timestamptz NOT NULL,

  CONSTRAINT pk_transaction_tags PRIMARY KEY (transaction_id, tag_id),
  CONSTRAINT fk_transaction_tags_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  CONSTRAINT fk_transaction_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transaction_tags_transaction_id ON transaction_tags(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_tags_tag_id ON transaction_tags(tag_id);

CREATE TABLE IF NOT EXISTS budgets (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  name text,
  period budget_period NOT NULL,
  period_start date,
  period_end date,
  amount numeric(18,2) NOT NULL,
  alert_threshold_percent int,
  rollover_mode budget_rollover_mode,
  category_id text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_budgets_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_budgets_category_id FOREIGN KEY (category_id) REFERENCES categories(id),
  CONSTRAINT chk_budgets_alert_threshold CHECK (alert_threshold_percent IS NULL OR (alert_threshold_percent >= 0 AND alert_threshold_percent <= 100))
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_period ON budgets(user_id, period, period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_budgets_user_category ON budgets(user_id, category_id);

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

CREATE TABLE IF NOT EXISTS rotating_savings_groups (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  self_label text,
  account_id text NOT NULL,
  name text NOT NULL,
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

CREATE TABLE IF NOT EXISTS debts (
  id text PRIMARY KEY,
  client_id text,
  user_id text NOT NULL,
  account_id text,
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
  closed_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_debts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_debts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id),
  CONSTRAINT ck_debts_due_after_start CHECK (due_date >= start_date),
  CONSTRAINT ck_debts_outstanding_nonneg CHECK (outstanding_principal >= 0),
  CONSTRAINT ck_debts_accrued_nonneg CHECK (accrued_interest >= 0)
);

CREATE INDEX IF NOT EXISTS idx_debts_user_status ON debts(user_id, status);
CREATE INDEX IF NOT EXISTS idx_debts_user_due_date ON debts(user_id, due_date);
CREATE INDEX IF NOT EXISTS idx_debts_account_id ON debts(account_id);

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

CREATE TABLE IF NOT EXISTS investment_accounts (
  id text PRIMARY KEY,
  account_id text NOT NULL,
  fee_settings jsonb,
  tax_settings jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_investment_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_investment_accounts_account_id
  ON investment_accounts(account_id);

CREATE TABLE IF NOT EXISTS securities (
  id text PRIMARY KEY,
  symbol varchar NOT NULL,
  name varchar,
  asset_class security_asset_class,
  currency varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_securities_symbol
  ON securities(symbol);

CREATE TABLE IF NOT EXISTS security_price_dailies (
  id text NOT NULL,
  security_id text NOT NULL,
  price_date date NOT NULL,
  open numeric(18,8),
  high numeric(18,8),
  low numeric(18,8),
  close numeric(18,8) NOT NULL,
  volume numeric(18,2),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  PRIMARY KEY (id, price_date),
  CONSTRAINT fk_security_price_dailies_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_price_dailies_security_date
  ON security_price_dailies(security_id, price_date);

CREATE INDEX IF NOT EXISTS idx_security_price_dailies_price_date
  ON security_price_dailies(price_date);

CREATE TABLE IF NOT EXISTS security_events (
  id text PRIMARY KEY,
  security_id text NOT NULL,
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
  updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_market_data_sync_states_last_success_at
  ON market_data_sync_states(last_success_at);

CREATE TABLE IF NOT EXISTS security_event_elections (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  broker_account_id text NOT NULL,
  security_event_id text NOT NULL,
  security_id text NOT NULL,
  entitlement_date date NOT NULL,
  holding_quantity_at_entitlement_date numeric(18,8) NOT NULL,
  entitled_quantity numeric(18,8) NOT NULL,
  elected_quantity numeric(18,8) NOT NULL,
  status security_event_election_status NOT NULL,
  confirmed_at timestamptz,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_security_event_elections_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
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
  id text PRIMARY KEY,
  client_id text,
  broker_account_id text NOT NULL,
  security_id text NOT NULL,
  fee_transaction_id text,
  tax_transaction_id text,
  side trade_side NOT NULL,
  quantity numeric(18,8) NOT NULL,
  price numeric(18,8) NOT NULL,
  fees numeric(18,2) NOT NULL DEFAULT 0,
  taxes numeric(18,2) NOT NULL DEFAULT 0,
  occurred_at timestamptz NOT NULL,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

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
  id text PRIMARY KEY,
  broker_account_id text NOT NULL,
  security_id text NOT NULL,
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

  CONSTRAINT fk_holdings_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_holdings_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_holdings_broker_security
  ON holdings(broker_account_id, security_id);

CREATE TABLE IF NOT EXISTS share_lots (
  id text PRIMARY KEY,
  broker_account_id text NOT NULL,
  security_id text NOT NULL,
  quantity numeric(18,8) NOT NULL,
  acquisition_date date NOT NULL,
  cost_basis_per_share numeric(18,8) NOT NULL DEFAULT 0,
  provenance lot_provenance NOT NULL DEFAULT 'regular_buy',
  status lot_status NOT NULL DEFAULT 'active',
  buy_trade_id text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

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
  id text PRIMARY KEY,
  broker_account_id text NOT NULL,
  security_id text NOT NULL,
  sell_trade_id text NOT NULL,
  source_share_lot_id text NOT NULL,
  quantity numeric(18,8) NOT NULL,
  acquisition_date date NOT NULL,
  cost_basis_total numeric(18,2) NOT NULL,
  sell_price numeric(18,8) NOT NULL,
  proceeds numeric(18,2) NOT NULL,
  realized_pnl numeric(18,2) NOT NULL,
  provenance lot_provenance NOT NULL,
  created_at timestamptz NOT NULL,

  CONSTRAINT fk_realized_trade_logs_broker_account_id FOREIGN KEY (broker_account_id) REFERENCES investment_accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_sell_trade_id FOREIGN KEY (sell_trade_id) REFERENCES trades(id) ON DELETE CASCADE,
  CONSTRAINT fk_realized_trade_logs_source_share_lot_id FOREIGN KEY (source_share_lot_id) REFERENCES share_lots(id) ON DELETE RESTRICT,
  CONSTRAINT ck_realized_trade_logs_nonneg CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_realized_trade_logs_broker_security_sell
  ON realized_trade_logs(broker_account_id, security_id, sell_trade_id);

CREATE TABLE IF NOT EXISTS audit_events (
  id text NOT NULL,
  account_id text NOT NULL,
  actor_user_id text NOT NULL,
  action varchar NOT NULL,
  entity_type varchar NOT NULL,
  entity_id text NOT NULL,
  occurred_at timestamptz NOT NULL,
  diff jsonb,

  PRIMARY KEY (id, occurred_at),

  CONSTRAINT fk_audit_events_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_audit_events_actor_user_id FOREIGN KEY (actor_user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_id_occurred_at
  ON audit_events(account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_entity
  ON audit_events(account_id, entity_type, entity_id, occurred_at DESC);

-- Convert suitable time-series tables to hypertables.
-- Requirements: any UNIQUE/PRIMARY KEY constraints must include the time column.
SELECT public.create_hypertable('security_price_dailies', 'price_date', if_not_exists => TRUE);
SELECT public.create_hypertable('audit_events', 'occurred_at', if_not_exists => TRUE);


-- Seed data (categories)
INSERT INTO categories (id, name, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at)
VALUES
  -- Income
  ('cat_def_income', 'Income', NULL, 'income', 5, true, 'cash', 'green', now(), now()),

  -- Expense parents
  ('cat_def_food', 'Food & Drinks', NULL, 'expense', 10, true, 'salad', 'orange', now(), now()),
  ('cat_def_transport', 'Transport', NULL, 'expense', 20, true, 'car', 'blue', now(), now()),
  ('cat_def_shopping', 'Shopping', NULL, 'expense', 30, true, 'shopping-bag', 'violet', now(), now()),
  ('cat_def_bills', 'Bills', NULL, 'expense', 40, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_health', 'Health', NULL, 'expense', 50, true, 'heart', 'red', now(), now()),
  ('cat_def_entertainment', 'Entertainment', NULL, 'expense', 60, true, 'mask', 'grape', now(), now()),
  ('cat_def_education', 'Education', NULL, 'expense', 70, true, 'book', 'teal', now(), now()),
  ('cat_def_other_expense', 'Other', NULL, 'expense', 90, true, 'dots', 'gray', now(), now()),

  -- Income children
  ('cat_def_income_salary', 'Salary', 'cat_def_income', 'income', 6, true, 'cash', 'green', now(), now()),
  ('cat_def_income_bonus', 'Bonus', 'cat_def_income', 'income', 7, true, 'cash', 'green', now(), now()),
  ('cat_def_income_other', 'Other income', 'cat_def_income', 'income', 8, true, 'cash', 'green', now(), now()),
  ('cat_def_income_business', 'Business income', 'cat_def_income', 'income', 9, true, 'briefcase', 'green', now(), now()),
  ('cat_def_income_invest_interest', 'Interest', 'cat_def_income', 'income', 10, true, 'percentage', 'green', now(), now()),
  ('cat_def_income_invest_dividend', 'Dividends', 'cat_def_income', 'income', 11, true, 'chart-line', 'green', now(), now()),
  ('cat_def_income_rental', 'Rental income', 'cat_def_income', 'income', 12, true, 'home', 'green', now(), now()),
  ('cat_def_income_gift', 'Gifts received', 'cat_def_income', 'income', 13, true, 'gift', 'green', now(), now()),
  ('cat_def_income_refund', 'Refunds', 'cat_def_income', 'income', 14, true, 'rotate', 'green', now(), now()),
  ('cat_def_income_reimbursement', 'Reimbursements', 'cat_def_income', 'income', 15, true, 'receipt-refund', 'green', now(), now()),
  ('cat_def_income_cashback', 'Cashback', 'cat_def_income', 'income', 16, true, 'coin', 'green', now(), now()),
  ('cat_def_income_sale', 'Sell items', 'cat_def_income', 'income', 17, true, 'tag', 'green', now(), now()),

  -- Food & Drinks children
  ('cat_def_food_groceries', 'Groceries', 'cat_def_food', 'expense', 11, true, 'salad', 'orange', now(), now()),
  ('cat_def_food_eating_out', 'Eating out', 'cat_def_food', 'expense', 12, true, 'salad', 'orange', now(), now()),
  ('cat_def_food_coffee', 'Coffee & Tea', 'cat_def_food', 'expense', 13, true, 'salad', 'orange', now(), now()),
  ('cat_def_food_delivery', 'Delivery', 'cat_def_food', 'expense', 14, true, 'scooter', 'orange', now(), now()),
  ('cat_def_food_snacks', 'Snacks', 'cat_def_food', 'expense', 15, true, 'cookie', 'orange', now(), now()),
  ('cat_def_food_alcohol', 'Alcohol', 'cat_def_food', 'expense', 16, true, 'glass', 'orange', now(), now()),

  -- Transport children
  ('cat_def_transport_gas', 'Gas', 'cat_def_transport', 'expense', 21, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_taxi', 'Taxi / Grab', 'cat_def_transport', 'expense', 22, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_public', 'Public transit', 'cat_def_transport', 'expense', 23, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_parking', 'Parking', 'cat_def_transport', 'expense', 24, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_tolls', 'Tolls', 'cat_def_transport', 'expense', 25, true, 'road', 'blue', now(), now()),
  ('cat_def_transport_maintenance', 'Vehicle maintenance', 'cat_def_transport', 'expense', 26, true, 'tools', 'blue', now(), now()),
  ('cat_def_transport_insurance', 'Vehicle insurance', 'cat_def_transport', 'expense', 27, true, 'shield', 'blue', now(), now()),
  ('cat_def_transport_car_payment', 'Car payment', 'cat_def_transport', 'expense', 28, true, 'credit-card', 'blue', now(), now()),

  -- Shopping children
  ('cat_def_shopping_household', 'Household', 'cat_def_shopping', 'expense', 31, true, 'shopping-bag', 'violet', now(), now()),
  ('cat_def_shopping_clothes', 'Clothes', 'cat_def_shopping', 'expense', 32, true, 'shopping-bag', 'violet', now(), now()),
  ('cat_def_shopping_electronics', 'Electronics', 'cat_def_shopping', 'expense', 33, true, 'shopping-bag', 'violet', now(), now()),
  ('cat_def_shopping_personal_care', 'Personal care', 'cat_def_shopping', 'expense', 34, true, 'sparkles', 'violet', now(), now()),
  ('cat_def_shopping_gifts', 'Gifts', 'cat_def_shopping', 'expense', 35, true, 'gift', 'violet', now(), now()),
  ('cat_def_shopping_online', 'Online shopping', 'cat_def_shopping', 'expense', 36, true, 'shopping-cart', 'violet', now(), now()),
  ('cat_def_shopping_cosmetics', 'Cosmetics', 'cat_def_shopping', 'expense', 37, true, 'sparkles', 'violet', now(), now()),

  -- Bills children
  ('cat_def_bills_rent', 'Rent', 'cat_def_bills', 'expense', 41, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_bills_utilities', 'Utilities', 'cat_def_bills', 'expense', 42, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_bills_internet', 'Internet', 'cat_def_bills', 'expense', 43, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_bills_phone', 'Phone', 'cat_def_bills', 'expense', 44, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_bills_mortgage', 'Mortgage', 'cat_def_bills', 'expense', 45, true, 'home', 'cyan', now(), now()),
  ('cat_def_bills_hoa', 'HOA / Building fees', 'cat_def_bills', 'expense', 46, true, 'building', 'cyan', now(), now()),
  ('cat_def_bills_repairs', 'Home repairs', 'cat_def_bills', 'expense', 47, true, 'hammer', 'cyan', now(), now()),
  ('cat_def_bills_subscriptions', 'Subscriptions', 'cat_def_bills', 'expense', 48, true, 'device-tv', 'cyan', now(), now()),
  ('cat_def_bills_insurance', 'Home insurance', 'cat_def_bills', 'expense', 49, true, 'shield-home', 'cyan', now(), now()),
  ('cat_def_bills_electricity', 'Electricity', 'cat_def_bills', 'expense', 50, true, 'bolt', 'cyan', now(), now()),
  ('cat_def_bills_water', 'Water', 'cat_def_bills', 'expense', 51, true, 'droplet', 'cyan', now(), now()),
  ('cat_def_bills_gas', 'Gas (utility)', 'cat_def_bills', 'expense', 52, true, 'flame', 'cyan', now(), now()),
  ('cat_def_bills_trash', 'Trash', 'cat_def_bills', 'expense', 53, true, 'trash', 'cyan', now(), now()),
  ('cat_def_bills_property_tax', 'Property tax', 'cat_def_bills', 'expense', 54, true, 'building-bank', 'cyan', now(), now()),

  -- Health children
  ('cat_def_health_medical', 'Medical', 'cat_def_health', 'expense', 51, true, 'heart', 'red', now(), now()),
  ('cat_def_health_pharmacy', 'Pharmacy', 'cat_def_health', 'expense', 52, true, 'heart', 'red', now(), now()),
  ('cat_def_health_insurance', 'Insurance', 'cat_def_health', 'expense', 53, true, 'heart', 'red', now(), now()),
  ('cat_def_health_dental', 'Dental', 'cat_def_health', 'expense', 54, true, 'tooth', 'red', now(), now()),
  ('cat_def_health_vision', 'Vision', 'cat_def_health', 'expense', 55, true, 'eye', 'red', now(), now()),
  ('cat_def_health_gym', 'Gym / Fitness', 'cat_def_health', 'expense', 56, true, 'barbell', 'red', now(), now()),
  ('cat_def_health_mental', 'Mental health', 'cat_def_health', 'expense', 57, true, 'brain', 'red', now(), now()),

  -- Entertainment children
  ('cat_def_ent_movies', 'Movies', 'cat_def_entertainment', 'expense', 61, true, 'mask', 'grape', now(), now()),
  ('cat_def_ent_games', 'Games', 'cat_def_entertainment', 'expense', 62, true, 'mask', 'grape', now(), now()),
  ('cat_def_ent_travel', 'Travel', 'cat_def_entertainment', 'expense', 63, true, 'mask', 'grape', now(), now()),
  ('cat_def_ent_streaming', 'Streaming', 'cat_def_entertainment', 'expense', 64, true, 'device-tv', 'grape', now(), now()),
  ('cat_def_ent_events', 'Events', 'cat_def_entertainment', 'expense', 65, true, 'ticket', 'grape', now(), now()),
  ('cat_def_ent_hobbies', 'Hobbies', 'cat_def_entertainment', 'expense', 66, true, 'palette', 'grape', now(), now()),
  ('cat_def_ent_music', 'Music', 'cat_def_entertainment', 'expense', 67, true, 'music', 'grape', now(), now()),
  ('cat_def_ent_sports', 'Sports', 'cat_def_entertainment', 'expense', 68, true, 'ball-basketball', 'grape', now(), now()),

  -- Education children
  ('cat_def_edu_courses', 'Courses', 'cat_def_education', 'expense', 71, true, 'book', 'teal', now(), now()),
  ('cat_def_edu_books', 'Books', 'cat_def_education', 'expense', 72, true, 'book', 'teal', now(), now()),
  ('cat_def_edu_tuition', 'Tuition', 'cat_def_education', 'expense', 73, true, 'school', 'teal', now(), now()),
  ('cat_def_edu_supplies', 'Supplies', 'cat_def_education', 'expense', 74, true, 'pencil', 'teal', now(), now()),
  ('cat_def_edu_certifications', 'Certifications', 'cat_def_education', 'expense', 75, true, 'certificate', 'teal', now(), now()),

  -- Family
  ('cat_def_family', 'Family', NULL, 'expense', 80, true, 'users', 'pink', now(), now()),
  ('cat_def_family_childcare', 'Childcare', 'cat_def_family', 'expense', 81, true, 'baby-carriage', 'pink', now(), now()),
  ('cat_def_family_kids', 'Kids', 'cat_def_family', 'expense', 82, true, 'balloon', 'pink', now(), now()),
  ('cat_def_family_parents', 'Parents', 'cat_def_family', 'expense', 83, true, 'heart-handshake', 'pink', now(), now()),

  -- Pets
  ('cat_def_pets', 'Pets', NULL, 'expense', 85, true, 'paw', 'lime', now(), now()),
  ('cat_def_pets_food', 'Pet food', 'cat_def_pets', 'expense', 86, true, 'bone', 'lime', now(), now()),
  ('cat_def_pets_vet', 'Vet', 'cat_def_pets', 'expense', 87, true, 'stethoscope', 'lime', now(), now()),
  ('cat_def_pets_grooming', 'Grooming', 'cat_def_pets', 'expense', 88, true, 'cut', 'lime', now(), now()),

  -- Financial
  ('cat_def_financial', 'Financial', NULL, 'expense', 88, true, 'building-bank', 'yellow', now(), now()),
  ('cat_def_financial_bank_fees', 'Bank fees', 'cat_def_financial', 'expense', 89, true, 'receipt-tax', 'yellow', now(), now()),
  ('cat_def_financial_loan_interest', 'Loan interest', 'cat_def_financial', 'expense', 90, true, 'percentage', 'yellow', now(), now()),
  ('cat_def_financial_invest_fees', 'Investment fees', 'cat_def_financial', 'expense', 91, true, 'chart-line', 'yellow', now(), now()),

  -- Other expense
  ('cat_def_other_fees', 'Fees', 'cat_def_other_expense', 'expense', 92, true, 'receipt-tax', 'gray', now(), now()),
  ('cat_def_other_donations', 'Donations', 'cat_def_other_expense', 'expense', 93, true, 'heart-handshake', 'gray', now(), now()),
  ('cat_def_other_taxes', 'Taxes', 'cat_def_other_expense', 'expense', 94, true, 'building-bank', 'gray', now(), now()),

  -- System-only categories
  ('cat_sys_internal', 'System', NULL, 'both', 10000, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_adjustment', 'System adjustment', 'cat_sys_internal', 'both', 10001, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_sync', 'System sync', 'cat_sys_internal', 'both', 10002, true, 'settings', 'gray', now(), now())
ON CONFLICT (id) DO NOTHING;

UPDATE categories
SET is_system = true
WHERE id IN ('cat_sys_internal','cat_sys_internal_adjustment','cat_sys_internal_sync');


-- +goose Down
-- Intentionally omitted: this migration is a squashed baseline for fresh DBs.
