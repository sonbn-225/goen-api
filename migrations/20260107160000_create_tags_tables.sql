-- +goose Up

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

-- +goose Down
DROP TABLE IF EXISTS tags;
