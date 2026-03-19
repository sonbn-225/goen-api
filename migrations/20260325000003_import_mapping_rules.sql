-- +goose Up
-- Mapping rules for imported objects (account/category)

CREATE TABLE IF NOT EXISTS import_mapping_rules (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  kind text NOT NULL,
  source_name text NOT NULL,
  normalized_source_name text NOT NULL,
  mapped_id text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,

  CONSTRAINT fk_import_mapping_rules_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT chk_import_mapping_rules_kind CHECK (kind IN ('account', 'category')),
  CONSTRAINT uq_import_mapping_rules_unique UNIQUE (user_id, kind, normalized_source_name)
);

CREATE INDEX IF NOT EXISTS idx_import_mapping_rules_user_kind
  ON import_mapping_rules(user_id, kind, normalized_source_name);

-- +goose Down
DROP TABLE IF EXISTS import_mapping_rules;
