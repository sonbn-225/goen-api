-- +goose Up
INSERT INTO categories (id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at)
VALUES
  ('cat_sys_rotating_savings_contribution', 'cat_sys_internal', 'expense', 10020, true, 'users', 'cyan', now(), now()),
  ('cat_sys_rotating_savings_payout', 'cat_sys_internal', 'income', 10021, true, 'coins', 'cyan', now(), now())
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM categories
WHERE id IN ('cat_sys_rotating_savings_contribution', 'cat_sys_rotating_savings_payout')
  AND id NOT IN (
    SELECT DISTINCT category_id
    FROM transaction_line_items
    WHERE category_id IN ('cat_sys_rotating_savings_contribution', 'cat_sys_rotating_savings_payout')
  );
