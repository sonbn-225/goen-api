-- +goose Up
-- Expand default categories for practical day-to-day spending

-- Keep electricity and water separated clearly; utility parent stays for other utility fees.
UPDATE categories
SET is_active = false, updated_at = now()
WHERE id = 'cat_def_bills_utilities';

INSERT INTO categories (id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at)
VALUES
  -- Bills (utility details)
  ('cat_def_bills_gas', 'cat_def_bills', 'expense', 52, true, 'flame', 'cyan', now(), now()),
  ('cat_def_bills_waste', 'cat_def_bills', 'expense', 53, true, 'trash', 'cyan', now(), now()),

  -- Shopping details
  ('cat_def_shopping_jewelry', 'cat_def_shopping', 'expense', 47, true, 'diamond', 'violet', now(), now())
ON CONFLICT (id) DO UPDATE
SET parent_category_id = EXCLUDED.parent_category_id,
    type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active,
    icon = EXCLUDED.icon,
    color = EXCLUDED.color,
    updated_at = now();

-- +goose Down
UPDATE categories
SET is_active = true, updated_at = now()
WHERE id = 'cat_def_bills_utilities';

DELETE FROM categories
WHERE id IN (
  'cat_def_bills_gas',
  'cat_def_bills_waste',
  'cat_def_shopping_jewelry'
);
