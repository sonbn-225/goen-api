-- +goose Up

CREATE TABLE IF NOT EXISTS group_expense_participants (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  transaction_id text NOT NULL,
  participant_name text NOT NULL,
  original_amount numeric(18,2) NOT NULL,
  share_amount numeric(18,2) NOT NULL,
  is_settled boolean NOT NULL DEFAULT false,
  settlement_transaction_id text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_gep_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_gep_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  CONSTRAINT fk_gep_settlement_transaction_id FOREIGN KEY (settlement_transaction_id) REFERENCES transactions(id) ON DELETE SET NULL,
  CONSTRAINT ck_gep_amounts_positive CHECK (original_amount > 0 AND share_amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_gep_user_id ON group_expense_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_gep_transaction_id ON group_expense_participants(transaction_id);
CREATE INDEX IF NOT EXISTS idx_gep_user_settled ON group_expense_participants(user_id, is_settled);

-- Helps autocomplete participants.
CREATE INDEX IF NOT EXISTS idx_gep_user_participant_name ON group_expense_participants(user_id, lower(participant_name));

-- +goose Down

DROP TABLE IF EXISTS group_expense_participants;
