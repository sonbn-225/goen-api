-- +goose Up

CREATE TABLE IF NOT EXISTS categories (
  id text PRIMARY KEY,
  user_id text,
  name text NOT NULL,
  parent_category_id text,
  type text,
  sort_order int,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz,

  CONSTRAINT fk_categories_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_categories_parent FOREIGN KEY (parent_category_id) REFERENCES categories(id),
  CONSTRAINT chk_categories_type CHECK (type IS NULL OR type IN ('expense','income','both'))
);

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_category_id ON categories(parent_category_id);

-- Prevent duplicate logical categories.
-- Global (shared) categories
CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_global_name_parent_type
  ON categories (lower(name), parent_category_id, COALESCE(type, ''))
  WHERE user_id IS NULL AND deleted_at IS NULL;

-- Per-user custom categories
CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_user_name_parent_type
  ON categories (user_id, lower(name), parent_category_id, COALESCE(type, ''))
  WHERE user_id IS NOT NULL AND deleted_at IS NULL;

-- Seed shared predefined categories (parent/child). IDs are stable text keys.
INSERT INTO categories (id, user_id, name, parent_category_id, type, sort_order, is_active, created_at, updated_at)
VALUES
  ('cat_def_income', NULL, 'Income', NULL, 'income', 5, true, now(), now()),

  ('cat_def_food', NULL, 'Food & Drinks', NULL, 'expense', 10, true, now(), now()),
  ('cat_def_transport', NULL, 'Transport', NULL, 'expense', 20, true, now(), now()),
  ('cat_def_shopping', NULL, 'Shopping', NULL, 'expense', 30, true, now(), now()),
  ('cat_def_bills', NULL, 'Bills', NULL, 'expense', 40, true, now(), now()),
  ('cat_def_health', NULL, 'Health', NULL, 'expense', 50, true, now(), now()),
  ('cat_def_entertainment', NULL, 'Entertainment', NULL, 'expense', 60, true, now(), now()),
  ('cat_def_education', NULL, 'Education', NULL, 'expense', 70, true, now(), now()),
  ('cat_def_other_expense', NULL, 'Other', NULL, 'expense', 90, true, now(), now()),

  ('cat_def_income_salary', NULL, 'Salary', 'cat_def_income', 'income', 6, true, now(), now()),
  ('cat_def_income_bonus', NULL, 'Bonus', 'cat_def_income', 'income', 7, true, now(), now()),
  ('cat_def_income_other', NULL, 'Other income', 'cat_def_income', 'income', 8, true, now(), now()),

  ('cat_def_food_groceries', NULL, 'Groceries', 'cat_def_food', 'expense', 11, true, now(), now()),
  ('cat_def_food_eating_out', NULL, 'Eating out', 'cat_def_food', 'expense', 12, true, now(), now()),
  ('cat_def_food_coffee', NULL, 'Coffee & Tea', 'cat_def_food', 'expense', 13, true, now(), now()),

  ('cat_def_transport_gas', NULL, 'Gas', 'cat_def_transport', 'expense', 21, true, now(), now()),
  ('cat_def_transport_taxi', NULL, 'Taxi / Grab', 'cat_def_transport', 'expense', 22, true, now(), now()),
  ('cat_def_transport_public', NULL, 'Public transit', 'cat_def_transport', 'expense', 23, true, now(), now()),
  ('cat_def_transport_parking', NULL, 'Parking', 'cat_def_transport', 'expense', 24, true, now(), now()),

  ('cat_def_shopping_household', NULL, 'Household', 'cat_def_shopping', 'expense', 31, true, now(), now()),
  ('cat_def_shopping_clothes', NULL, 'Clothes', 'cat_def_shopping', 'expense', 32, true, now(), now()),
  ('cat_def_shopping_electronics', NULL, 'Electronics', 'cat_def_shopping', 'expense', 33, true, now(), now()),

  ('cat_def_bills_rent', NULL, 'Rent', 'cat_def_bills', 'expense', 41, true, now(), now()),
  ('cat_def_bills_utilities', NULL, 'Utilities', 'cat_def_bills', 'expense', 42, true, now(), now()),
  ('cat_def_bills_internet', NULL, 'Internet', 'cat_def_bills', 'expense', 43, true, now(), now()),
  ('cat_def_bills_phone', NULL, 'Phone', 'cat_def_bills', 'expense', 44, true, now(), now()),

  ('cat_def_health_medical', NULL, 'Medical', 'cat_def_health', 'expense', 51, true, now(), now()),
  ('cat_def_health_pharmacy', NULL, 'Pharmacy', 'cat_def_health', 'expense', 52, true, now(), now()),
  ('cat_def_health_insurance', NULL, 'Insurance', 'cat_def_health', 'expense', 53, true, now(), now()),

  ('cat_def_ent_movies', NULL, 'Movies', 'cat_def_entertainment', 'expense', 61, true, now(), now()),
  ('cat_def_ent_games', NULL, 'Games', 'cat_def_entertainment', 'expense', 62, true, now(), now()),
  ('cat_def_ent_travel', NULL, 'Travel', 'cat_def_entertainment', 'expense', 63, true, now(), now()),

  ('cat_def_edu_courses', NULL, 'Courses', 'cat_def_education', 'expense', 71, true, now(), now()),
  ('cat_def_edu_books', NULL, 'Books', 'cat_def_education', 'expense', 72, true, now(), now()),

  ('cat_def_other_misc', NULL, 'Misc', 'cat_def_other_expense', 'expense', 91, true, now(), now())
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_tli_category_id ON transaction_line_items(category_id);

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE transaction_line_items
    ADD CONSTRAINT fk_tli_category FOREIGN KEY (category_id) REFERENCES categories(id);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose Down

ALTER TABLE transaction_line_items DROP CONSTRAINT IF EXISTS fk_tli_category;

DROP TABLE IF EXISTS categories;
