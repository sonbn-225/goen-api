-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS audit_events (
  id text PRIMARY KEY,
  account_id text NOT NULL,
  actor_user_id text NOT NULL,
  action varchar NOT NULL,
  entity_type varchar NOT NULL,
  entity_id text NOT NULL,
  occurred_at timestamptz NOT NULL,
  diff jsonb,

  CONSTRAINT fk_audit_events_account_id FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_audit_events_actor_user_id FOREIGN KEY (actor_user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_id_occurred_at
  ON audit_events(account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_account_entity
  ON audit_events(account_id, entity_type, entity_id, occurred_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_events;

-- +goose StatementEnd
