-- +goose Up

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lot_provenance') THEN
    CREATE TYPE lot_provenance AS ENUM ('regular_buy', 'stock_dividend', 'rights_offering');
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lot_status') THEN
    CREATE TYPE lot_status AS ENUM ('active', 'adjusted');
  END IF;
END $$;

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

-- +goose Down

DROP TABLE IF EXISTS realized_trade_logs;
DROP TABLE IF EXISTS share_lots;

DROP TYPE IF EXISTS lot_status;
DROP TYPE IF EXISTS lot_provenance;
