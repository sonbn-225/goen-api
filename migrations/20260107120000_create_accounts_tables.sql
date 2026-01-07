-- +goose Up
-- +goose StatementBegin

-- Enums
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

-- Tables
CREATE TABLE IF NOT EXISTS accounts (
  id text PRIMARY KEY,
  client_id text,
  name varchar NOT NULL,
  account_type account_type NOT NULL,
  currency varchar NOT NULL,
  parent_account_id text,
  status account_status NOT NULL DEFAULT 'active',
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

-- Exactly one active owner per account
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_accounts_active_owner
  ON user_accounts(account_id)
  WHERE permission = 'owner' AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_user_accounts_user_id ON user_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_accounts_account_id ON user_accounts(account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_accounts;
DROP TABLE IF EXISTS accounts;

-- Types are shared; drop cautiously.
DROP TYPE IF EXISTS user_account_status;
DROP TYPE IF EXISTS user_account_permission;
DROP TYPE IF EXISTS account_status;
DROP TYPE IF EXISTS account_type;

-- +goose StatementEnd
