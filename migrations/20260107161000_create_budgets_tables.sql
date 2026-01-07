-- +goose Up

-- Enums
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE budget_period AS ENUM ('month','week','custom');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE budget_rollover_mode AS ENUM ('reset','carry_forward','accumulate');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS budgets (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  name text,
  period budget_period NOT NULL,
  period_start date,
  period_end date,
  amount numeric(18,2) NOT NULL,
  currency text,
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

-- +goose Down
DROP TABLE IF EXISTS budgets;

DROP TYPE IF EXISTS budget_rollover_mode;
DROP TYPE IF EXISTS budget_period;
