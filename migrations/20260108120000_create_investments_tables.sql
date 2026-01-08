-- +goose Up

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE security_asset_class AS ENUM ('stock','fund','crypto','bond','other');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE security_price_daily_source AS ENUM ('market_data_service');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE security_event_type AS ENUM ('dividend_cash','split','reverse_split','rights_issue','bonus_issue','additional_issue');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE security_event_election_status AS ENUM ('draft','confirmed','cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE trade_side AS ENUM ('buy','sell');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE holding_source_of_truth AS ENUM ('trades','sync');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- Investment accounts
CREATE TABLE IF NOT EXISTS investment_accounts (
  id text PRIMARY KEY,
  account_id text NOT NULL,
  broker_name varchar,
  currency varchar NOT NULL,
  sync_enabled boolean NOT NULL DEFAULT false,
  sync_settings jsonb,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_investment_accounts_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_investment_accounts_account_id
  ON investment_accounts(account_id);

-- Securities
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

-- Daily prices (read-only)
CREATE TABLE IF NOT EXISTS security_price_dailies (
  id text PRIMARY KEY,
  security_id text NOT NULL,
  price_date date NOT NULL,
  open numeric(18,8),
  high numeric(18,8),
  low numeric(18,8),
  close numeric(18,8) NOT NULL,
  adj_close numeric(18,8),
  volume numeric(18,2),
  currency varchar,
  source security_price_daily_source NOT NULL,
  source_row_id varchar,
  fetched_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_security_price_dailies_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_security_price_dailies_security_date
  ON security_price_dailies(security_id, price_date);

CREATE INDEX IF NOT EXISTS idx_security_price_dailies_price_date
  ON security_price_dailies(price_date);

-- Corporate events (read-only)
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
  source security_price_daily_source NOT NULL,
  source_event_id varchar,
  note varchar,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_security_events_security_id FOREIGN KEY (security_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_security_events_security_effective_date
  ON security_events(security_id, effective_date);

CREATE INDEX IF NOT EXISTS idx_security_events_security_ex_date
  ON security_events(security_id, ex_date);

-- Elections (user writable)
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

-- Trades
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

-- Holdings
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

-- +goose Down
DROP TABLE IF EXISTS holdings;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS security_event_elections;
DROP TABLE IF EXISTS security_events;
DROP TABLE IF EXISTS security_price_dailies;
DROP TABLE IF EXISTS securities;
DROP TABLE IF EXISTS investment_accounts;

DROP TYPE IF EXISTS holding_source_of_truth;
DROP TYPE IF EXISTS trade_side;
DROP TYPE IF EXISTS security_event_election_status;
DROP TYPE IF EXISTS security_event_type;
DROP TYPE IF EXISTS security_price_daily_source;
DROP TYPE IF EXISTS security_asset_class;
