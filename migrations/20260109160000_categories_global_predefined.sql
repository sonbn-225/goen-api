-- +goose Up

-- Remove any user-created categories (custom categories are no longer supported).
-- Preserve data integrity by nulling references first.
UPDATE transaction_line_items
SET category_id = NULL
WHERE category_id IN (SELECT id FROM categories WHERE user_id IS NOT NULL);

UPDATE budgets
SET category_id = NULL
WHERE category_id IN (SELECT id FROM categories WHERE user_id IS NOT NULL);

DELETE FROM categories WHERE user_id IS NOT NULL;

-- Drop old indexes/constraints referencing user_id.
ALTER TABLE categories DROP CONSTRAINT IF EXISTS fk_categories_user_id;
DROP INDEX IF EXISTS idx_categories_user_id;
DROP INDEX IF EXISTS uq_categories_user_name_parent_type;
DROP INDEX IF EXISTS uq_categories_global_name_parent_type;

-- Make categories global-only.
ALTER TABLE categories DROP COLUMN IF EXISTS user_id;

-- Add presentation fields.
ALTER TABLE categories ADD COLUMN IF NOT EXISTS icon text;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS color text;

-- Prevent duplicate logical categories.
CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_name_parent_type
  ON categories (lower(name), parent_category_id, COALESCE(type, ''))
  WHERE deleted_at IS NULL;

-- Set icon/color for existing seeded categories.
UPDATE categories SET icon = 'cash', color = 'green' WHERE id = 'cat_def_income' AND (icon IS NULL OR color IS NULL);

UPDATE categories SET icon = 'salad', color = 'orange' WHERE id = 'cat_def_food' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'car', color = 'blue' WHERE id = 'cat_def_transport' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'shopping-bag', color = 'violet' WHERE id = 'cat_def_shopping' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'receipt', color = 'cyan' WHERE id = 'cat_def_bills' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'heart', color = 'red' WHERE id = 'cat_def_health' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'mask', color = 'grape' WHERE id = 'cat_def_entertainment' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'book', color = 'teal' WHERE id = 'cat_def_education' AND (icon IS NULL OR color IS NULL);
UPDATE categories SET icon = 'dots', color = 'gray' WHERE id = 'cat_def_other_expense' AND (icon IS NULL OR color IS NULL);

-- Expand predefined categories. IDs are stable text keys.
INSERT INTO categories (id, name, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at)
VALUES
  -- Income
  ('cat_def_income_business', 'Business income', 'cat_def_income', 'income', 9, true, 'briefcase', 'green', now(), now()),
  ('cat_def_income_invest_interest', 'Interest', 'cat_def_income', 'income', 10, true, 'percentage', 'green', now(), now()),
  ('cat_def_income_invest_dividend', 'Dividends', 'cat_def_income', 'income', 11, true, 'chart-line', 'green', now(), now()),
  ('cat_def_income_rental', 'Rental income', 'cat_def_income', 'income', 12, true, 'home', 'green', now(), now()),
  ('cat_def_income_gift', 'Gifts received', 'cat_def_income', 'income', 13, true, 'gift', 'green', now(), now()),
  ('cat_def_income_refund', 'Refunds', 'cat_def_income', 'income', 14, true, 'rotate', 'green', now(), now()),
  ('cat_def_income_reimbursement', 'Reimbursements', 'cat_def_income', 'income', 15, true, 'receipt-refund', 'green', now(), now()),
  ('cat_def_income_cashback', 'Cashback', 'cat_def_income', 'income', 16, true, 'coin', 'green', now(), now()),
  ('cat_def_income_sale', 'Sell items', 'cat_def_income', 'income', 17, true, 'tag', 'green', now(), now()),

  -- Food & Drinks
  ('cat_def_food_delivery', 'Delivery', 'cat_def_food', 'expense', 14, true, 'scooter', 'orange', now(), now()),
  ('cat_def_food_snacks', 'Snacks', 'cat_def_food', 'expense', 15, true, 'cookie', 'orange', now(), now()),
  ('cat_def_food_alcohol', 'Alcohol', 'cat_def_food', 'expense', 16, true, 'glass', 'orange', now(), now()),

  -- Transport
  ('cat_def_transport_tolls', 'Tolls', 'cat_def_transport', 'expense', 25, true, 'road', 'blue', now(), now()),
  ('cat_def_transport_maintenance', 'Vehicle maintenance', 'cat_def_transport', 'expense', 26, true, 'tools', 'blue', now(), now()),
  ('cat_def_transport_insurance', 'Vehicle insurance', 'cat_def_transport', 'expense', 27, true, 'shield', 'blue', now(), now()),
  ('cat_def_transport_car_payment', 'Car payment', 'cat_def_transport', 'expense', 28, true, 'credit-card', 'blue', now(), now()),

  -- Shopping
  ('cat_def_shopping_personal_care', 'Personal care', 'cat_def_shopping', 'expense', 34, true, 'sparkles', 'violet', now(), now()),
  ('cat_def_shopping_gifts', 'Gifts', 'cat_def_shopping', 'expense', 35, true, 'gift', 'violet', now(), now()),
  ('cat_def_shopping_online', 'Online shopping', 'cat_def_shopping', 'expense', 36, true, 'shopping-cart', 'violet', now(), now()),
  ('cat_def_shopping_cosmetics', 'Cosmetics', 'cat_def_shopping', 'expense', 37, true, 'sparkles', 'violet', now(), now()),

  -- Bills
  ('cat_def_bills_mortgage', 'Mortgage', 'cat_def_bills', 'expense', 45, true, 'home', 'cyan', now(), now()),
  ('cat_def_bills_hoa', 'HOA / Building fees', 'cat_def_bills', 'expense', 46, true, 'building', 'cyan', now(), now()),
  ('cat_def_bills_repairs', 'Home repairs', 'cat_def_bills', 'expense', 47, true, 'hammer', 'cyan', now(), now()),
  ('cat_def_bills_subscriptions', 'Subscriptions', 'cat_def_bills', 'expense', 48, true, 'device-tv', 'cyan', now(), now()),
  ('cat_def_bills_insurance', 'Home insurance', 'cat_def_bills', 'expense', 49, true, 'shield-home', 'cyan', now(), now()),
  ('cat_def_bills_electricity', 'Electricity', 'cat_def_bills', 'expense', 50, true, 'bolt', 'cyan', now(), now()),
  ('cat_def_bills_water', 'Water', 'cat_def_bills', 'expense', 51, true, 'droplet', 'cyan', now(), now()),
  ('cat_def_bills_gas', 'Gas (utility)', 'cat_def_bills', 'expense', 52, true, 'flame', 'cyan', now(), now()),
  ('cat_def_bills_trash', 'Trash', 'cat_def_bills', 'expense', 53, true, 'trash', 'cyan', now(), now()),
  ('cat_def_bills_property_tax', 'Property tax', 'cat_def_bills', 'expense', 54, true, 'building-bank', 'cyan', now(), now()),

  -- Health
  ('cat_def_health_dental', 'Dental', 'cat_def_health', 'expense', 54, true, 'tooth', 'red', now(), now()),
  ('cat_def_health_vision', 'Vision', 'cat_def_health', 'expense', 55, true, 'eye', 'red', now(), now()),
  ('cat_def_health_gym', 'Gym / Fitness', 'cat_def_health', 'expense', 56, true, 'barbell', 'red', now(), now()),
  ('cat_def_health_mental', 'Mental health', 'cat_def_health', 'expense', 57, true, 'brain', 'red', now(), now()),

  -- Entertainment
  ('cat_def_ent_streaming', 'Streaming', 'cat_def_entertainment', 'expense', 64, true, 'device-tv', 'grape', now(), now()),
  ('cat_def_ent_events', 'Events', 'cat_def_entertainment', 'expense', 65, true, 'ticket', 'grape', now(), now()),
  ('cat_def_ent_hobbies', 'Hobbies', 'cat_def_entertainment', 'expense', 66, true, 'palette', 'grape', now(), now()),
  ('cat_def_ent_music', 'Music', 'cat_def_entertainment', 'expense', 67, true, 'music', 'grape', now(), now()),
  ('cat_def_ent_sports', 'Sports', 'cat_def_entertainment', 'expense', 68, true, 'ball-basketball', 'grape', now(), now()),

  -- Education
  ('cat_def_edu_tuition', 'Tuition', 'cat_def_education', 'expense', 73, true, 'school', 'teal', now(), now()),
  ('cat_def_edu_supplies', 'Supplies', 'cat_def_education', 'expense', 74, true, 'pencil', 'teal', now(), now()),
  ('cat_def_edu_certifications', 'Certifications', 'cat_def_education', 'expense', 75, true, 'certificate', 'teal', now(), now()),

  -- Family
  ('cat_def_family', 'Family', NULL, 'expense', 80, true, 'users', 'pink', now(), now()),
  ('cat_def_family_childcare', 'Childcare', 'cat_def_family', 'expense', 81, true, 'baby-carriage', 'pink', now(), now()),
  ('cat_def_family_kids', 'Kids', 'cat_def_family', 'expense', 82, true, 'balloon', 'pink', now(), now()),
  ('cat_def_family_parents', 'Parents', 'cat_def_family', 'expense', 83, true, 'heart-handshake', 'pink', now(), now()),

  -- Pets
  ('cat_def_pets', 'Pets', NULL, 'expense', 85, true, 'paw', 'lime', now(), now()),
  ('cat_def_pets_food', 'Pet food', 'cat_def_pets', 'expense', 86, true, 'bone', 'lime', now(), now()),
  ('cat_def_pets_vet', 'Vet', 'cat_def_pets', 'expense', 87, true, 'stethoscope', 'lime', now(), now()),
  ('cat_def_pets_grooming', 'Grooming', 'cat_def_pets', 'expense', 88, true, 'cut', 'lime', now(), now()),

  -- Financial
  ('cat_def_financial', 'Financial', NULL, 'expense', 88, true, 'building-bank', 'yellow', now(), now()),
  ('cat_def_financial_bank_fees', 'Bank fees', 'cat_def_financial', 'expense', 89, true, 'receipt-tax', 'yellow', now(), now()),
  ('cat_def_financial_loan_interest', 'Loan interest', 'cat_def_financial', 'expense', 90, true, 'percentage', 'yellow', now(), now()),
  ('cat_def_financial_invest_fees', 'Investment fees', 'cat_def_financial', 'expense', 91, true, 'chart-line', 'yellow', now(), now()),

  -- Other expense
  ('cat_def_other_fees', 'Fees', 'cat_def_other_expense', 'expense', 92, true, 'receipt-tax', 'gray', now(), now()),
  ('cat_def_other_donations', 'Donations', 'cat_def_other_expense', 'expense', 93, true, 'heart-handshake', 'gray', now(), now()),
  ('cat_def_other_taxes', 'Taxes', 'cat_def_other_expense', 'expense', 94, true, 'building-bank', 'gray', now(), now())
ON CONFLICT (id) DO NOTHING;

-- +goose Down

-- Best-effort rollback: keep schema compatible by restoring user_id.
ALTER TABLE categories ADD COLUMN IF NOT EXISTS user_id text;

ALTER TABLE categories DROP COLUMN IF EXISTS icon;
ALTER TABLE categories DROP COLUMN IF EXISTS color;

DROP INDEX IF EXISTS uq_categories_name_parent_type;

-- Note: data removals (custom categories) are not restored.
