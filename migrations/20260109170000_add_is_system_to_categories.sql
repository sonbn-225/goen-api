-- +goose Up

-- Categories: some are reserved for the system and cannot be assigned by users.
ALTER TABLE categories
  ADD COLUMN IF NOT EXISTS is_system boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_categories_is_system ON categories(is_system);

-- Seed a few system-only categories for internal/system-generated transactions.
-- These are intentionally not exposed to users in the categories list endpoint.
INSERT INTO categories (id, name, parent_category_id, type, sort_order, is_active, is_system, icon, color, created_at, updated_at)
VALUES
  ('cat_sys_internal', 'System', NULL, 'both', 10000, true, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_adjustment', 'System adjustment', 'cat_sys_internal', 'both', 10001, true, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_sync', 'System sync', 'cat_sys_internal', 'both', 10002, true, true, 'settings', 'gray', now(), now())
ON CONFLICT (id) DO NOTHING;

-- +goose Down

DROP INDEX IF EXISTS idx_categories_is_system;

ALTER TABLE categories
  DROP COLUMN IF EXISTS is_system;
