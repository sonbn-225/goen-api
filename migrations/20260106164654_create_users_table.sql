-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  email text NULL,
  phone text NULL,
  display_name text NULL,
  status text NOT NULL DEFAULT 'active',
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT users_email_or_phone_chk CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uq ON users (lower(email)) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_uq ON users (phone) WHERE phone IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
