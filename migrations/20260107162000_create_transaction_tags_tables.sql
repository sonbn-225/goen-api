-- +goose Up

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

-- +goose Down
DROP TABLE IF EXISTS transaction_tags;
