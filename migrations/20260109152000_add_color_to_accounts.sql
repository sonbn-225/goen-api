-- +goose Up
-- +goose StatementBegin

ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS color varchar;

CREATE INDEX IF NOT EXISTS idx_accounts_color ON accounts(color);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_accounts_color;
ALTER TABLE accounts
  DROP COLUMN IF EXISTS color;

-- +goose StatementEnd
