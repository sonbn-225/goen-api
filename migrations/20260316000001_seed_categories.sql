-- +goose Up
-- Seed categories for the application
-- This migration is separate from the baseline schema to allow idempotent seeds
-- Policy: cat_sys_* IDs are reserved system categories and must remain stable.
-- Taxonomy expansion/adjustment should be done via cat_def_* categories only.

INSERT INTO categories (id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at)
VALUES
  -- ============ SYSTEM CATEGORIES ============
  ('cat_sys_internal', NULL, 'both', 10000, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_adjustment', 'cat_sys_internal', 'both', 10001, true, 'settings', 'gray', now(), now()),
  ('cat_sys_internal_sync', 'cat_sys_internal', 'both', 10002, true, 'settings', 'gray', now(), now()),

  -- ============ INCOME CATEGORIES ============
  ('cat_def_income', NULL, 'income', 5, true, 'cash', 'green', now(), now()),
  ('cat_def_income_salary', 'cat_def_income', 'income', 6, true, 'cash', 'green', now(), now()),
  ('cat_def_income_bonus', 'cat_def_income', 'income', 7, true, 'cash', 'green', now(), now()),
  ('cat_def_income_business', 'cat_def_income', 'income', 9, true, 'briefcase', 'green', now(), now()),
  ('cat_def_income_invest_interest', 'cat_def_income', 'income', 10, true, 'percent', 'green', now(), now()),
  ('cat_def_income_invest_dividend', 'cat_def_income', 'income', 11, true, 'chart-line', 'green', now(), now()),
  ('cat_def_income_rental', 'cat_def_income', 'income', 12, true, 'home', 'green', now(), now()),
  ('cat_def_income_gift', 'cat_def_income', 'income', 13, true, 'gift', 'green', now(), now()),
  ('cat_def_income_refund', 'cat_def_income', 'income', 14, true, 'rotate', 'green', now(), now()),
  ('cat_def_income_sale', 'cat_def_income', 'income', 17, true, 'tag', 'green', now(), now()),

  -- ============ FOOD & DRINKS ============
  ('cat_def_food', NULL, 'expense', 10, true, 'salad', 'orange', now(), now()),
  ('cat_def_food_groceries', 'cat_def_food', 'expense', 11, true, 'basket', 'orange', now(), now()),
  ('cat_def_food_eating_out', 'cat_def_food', 'expense', 12, true, 'utensils', 'orange', now(), now()),
  ('cat_def_food_coffee', 'cat_def_food', 'expense', 13, true, 'coffee', 'orange', now(), now()),
  ('cat_def_food_delivery', 'cat_def_food', 'expense', 14, true, 'scooter', 'orange', now(), now()),
  ('cat_def_food_snacks', 'cat_def_food', 'expense', 15, true, 'cookie', 'orange', now(), now()),

  -- ============ TRANSPORT ============
  ('cat_def_transport', NULL, 'expense', 20, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_gas', 'cat_def_transport', 'expense', 31, true, 'gas-pump', 'blue', now(), now()),
  ('cat_def_transport_taxi', 'cat_def_transport', 'expense', 32, true, 'car', 'blue', now(), now()),
  ('cat_def_transport_public', 'cat_def_transport', 'expense', 33, true, 'bus', 'blue', now(), now()),
  ('cat_def_transport_parking', 'cat_def_transport', 'expense', 34, true, 'parking', 'blue', now(), now()),
  ('cat_def_transport_maintenance', 'cat_def_transport', 'expense', 36, true, 'wrench', 'blue', now(), now()),
  ('cat_def_transport_insurance', 'cat_def_transport', 'expense', 37, true, 'shield', 'blue', now(), now()),

  -- ============ SHOPPING ============
  ('cat_def_shopping', NULL, 'expense', 30, true, 'shopping-bag', 'violet', now(), now()),
  ('cat_def_shopping_household', 'cat_def_shopping', 'expense', 41, true, 'home', 'violet', now(), now()),
  ('cat_def_shopping_clothes', 'cat_def_shopping', 'expense', 42, true, 'shirt', 'violet', now(), now()),
  ('cat_def_shopping_electronics', 'cat_def_shopping', 'expense', 43, true, 'laptop', 'violet', now(), now()),
  ('cat_def_shopping_personal_care', 'cat_def_shopping', 'expense', 44, true, 'spray-bottle', 'violet', now(), now()),
  ('cat_def_shopping_gifts', 'cat_def_shopping', 'expense', 45, true, 'gift', 'violet', now(), now()),
  ('cat_def_shopping_online', 'cat_def_shopping', 'expense', 46, true, 'shopping-cart', 'violet', now(), now()),
  ('cat_def_shopping_jewelry', 'cat_def_shopping', 'expense', 47, true, 'diamond', 'violet', now(), now()),

  -- ============ BILLS & HOUSING ============
  ('cat_def_bills', NULL, 'expense', 40, true, 'receipt', 'cyan', now(), now()),
  ('cat_def_bills_rent', 'cat_def_bills', 'expense', 41, true, 'home', 'cyan', now(), now()),
  ('cat_def_bills_internet', 'cat_def_bills', 'expense', 43, true, 'wifi', 'cyan', now(), now()),
  ('cat_def_bills_phone', 'cat_def_bills', 'expense', 44, true, 'phone', 'cyan', now(), now()),
  ('cat_def_bills_mortgage', 'cat_def_bills', 'expense', 45, true, 'building', 'cyan', now(), now()),
  ('cat_def_bills_repairs', 'cat_def_bills', 'expense', 47, true, 'hammer', 'cyan', now(), now()),
  ('cat_def_bills_subscriptions', 'cat_def_bills', 'expense', 48, true, 'device-tv', 'cyan', now(), now()),
  ('cat_def_bills_electricity', 'cat_def_bills', 'expense', 50, true, 'zap', 'cyan', now(), now()),
  ('cat_def_bills_water', 'cat_def_bills', 'expense', 51, true, 'droplet', 'cyan', now(), now()),
  ('cat_def_bills_gas', 'cat_def_bills', 'expense', 52, true, 'flame', 'cyan', now(), now()),
  ('cat_def_bills_waste', 'cat_def_bills', 'expense', 53, true, 'trash', 'cyan', now(), now()),

  -- ============ HEALTH ============
  ('cat_def_health', NULL, 'expense', 50, true, 'heart', 'red', now(), now()),
  ('cat_def_health_medical', 'cat_def_health', 'expense', 51, true, 'stethoscope', 'red', now(), now()),
  ('cat_def_health_pharmacy', 'cat_def_health', 'expense', 52, true, 'pill', 'red', now(), now()),
  ('cat_def_health_insurance', 'cat_def_health', 'expense', 53, true, 'shield', 'red', now(), now()),
  ('cat_def_health_dental', 'cat_def_health', 'expense', 54, true, 'tooth', 'red', now(), now()),
  ('cat_def_health_gym', 'cat_def_health', 'expense', 56, true, 'barbell', 'red', now(), now()),

  -- ============ ENTERTAINMENT ============
  ('cat_def_entertainment', NULL, 'expense', 60, true, 'mask', 'grape', now(), now()),
  ('cat_def_ent_movies', 'cat_def_entertainment', 'expense', 61, true, 'film', 'grape', now(), now()),
  ('cat_def_ent_games', 'cat_def_entertainment', 'expense', 62, true, 'joystick', 'grape', now(), now()),
  ('cat_def_ent_travel', 'cat_def_entertainment', 'expense', 63, true, 'map', 'grape', now(), now()),
  ('cat_def_ent_streaming', 'cat_def_entertainment', 'expense', 64, true, 'device-tv', 'grape', now(), now()),
  ('cat_def_ent_hobbies', 'cat_def_entertainment', 'expense', 66, true, 'palette', 'grape', now(), now()),
  ('cat_def_ent_music', 'cat_def_entertainment', 'expense', 67, true, 'music', 'grape', now(), now()),

  -- ============ EDUCATION ============
  ('cat_def_education', NULL, 'expense', 70, true, 'book', 'teal', now(), now()),
  ('cat_def_edu_courses', 'cat_def_education', 'expense', 71, true, 'book', 'teal', now(), now()),
  ('cat_def_edu_books', 'cat_def_education', 'expense', 72, true, 'book-open', 'teal', now(), now()),
  ('cat_def_edu_tuition', 'cat_def_education', 'expense', 73, true, 'school', 'teal', now(), now()),
  ('cat_def_edu_supplies', 'cat_def_education', 'expense', 74, true, 'pencil', 'teal', now(), now()),

  -- ============ BEAUTY & GROOMING ============
  ('cat_def_beauty', NULL, 'expense', 18, true, 'sparkles', 'pink', now(), now()),
  ('cat_def_beauty_haircut', 'cat_def_beauty', 'expense', 19, true, 'scissors', 'pink', now(), now()),
  ('cat_def_beauty_spa', 'cat_def_beauty', 'expense', 20, true, 'leaf', 'pink', now(), now()),
  ('cat_def_beauty_cosmetics', 'cat_def_beauty', 'expense', 21, true, 'sparkles', 'pink', now(), now()),

  -- ============ FAMILY ============
  ('cat_def_family', NULL, 'expense', 80, true, 'users', 'pink', now(), now()),
  ('cat_def_family_childcare', 'cat_def_family', 'expense', 81, true, 'baby-carriage', 'pink', now(), now()),
  ('cat_def_family_kids', 'cat_def_family', 'expense', 82, true, 'balloon', 'pink', now(), now()),
  ('cat_def_family_tuition', 'cat_def_family', 'expense', 83, true, 'school', 'pink', now(), now()),
  ('cat_def_family_support', 'cat_def_family', 'expense', 84, true, 'heart-handshake', 'pink', now(), now()),

  -- ============ PETS ============
  ('cat_def_pets', NULL, 'expense', 85, true, 'paw', 'lime', now(), now()),
  ('cat_def_pets_food', 'cat_def_pets', 'expense', 86, true, 'bone', 'lime', now(), now()),
  ('cat_def_pets_vet', 'cat_def_pets', 'expense', 87, true, 'stethoscope', 'lime', now(), now()),
  ('cat_def_pets_grooming', 'cat_def_pets', 'expense', 88, true, 'scissors', 'lime', now(), now()),

  -- ============ FINANCIAL ============
  ('cat_def_financial', NULL, 'expense', 88, true, 'building-bank', 'yellow', now(), now()),
  ('cat_def_financial_bank_fees', 'cat_def_financial', 'expense', 89, true, 'receipt-tax', 'yellow', now(), now()),
  ('cat_def_financial_invest', 'cat_def_financial', 'expense', 91, true, 'chart-line', 'yellow', now(), now()),
  ('cat_sys_invest_buy', 'cat_def_financial', 'expense', 911, true, 'arrow-down', 'yellow', now(), now()),
  ('cat_sys_invest_sell', 'cat_def_financial', 'income', 912, true, 'arrow-up', 'yellow', now(), now()),
  ('cat_sys_invest_fees', 'cat_def_financial', 'expense', 913, true, 'receipt-tax', 'yellow', now(), now()),
  ('cat_def_financial_insurance', 'cat_def_financial', 'expense', 92, true, 'umbrella', 'yellow', now(), now()),
  ('cat_def_financial_debt', 'cat_def_financial', 'expense', 93, true, 'arrow-down', 'yellow', now(), now()),

  -- ============ TECHNOLOGY ============
  ('cat_def_tech', NULL, 'expense', 56, true, 'cpu', 'cyan', now(), now()),
  ('cat_def_tech_gadgets', 'cat_def_tech', 'expense', 57, true, 'device-mobile', 'cyan', now(), now()),
  ('cat_def_tech_software', 'cat_def_tech', 'expense', 58, true, 'code', 'cyan', now(), now()),
  ('cat_def_tech_cloud', 'cat_def_tech', 'expense', 59, true, 'cloud', 'cyan', now(), now()),

  -- ============ OTHER ============
  ('cat_def_other_expense', NULL, 'expense', 90, true, 'dots', 'gray', now(), now()),
  ('cat_def_other_fees', 'cat_def_other_expense', 'expense', 92, true, 'receipt-tax', 'gray', now(), now()),
  ('cat_def_other_donations', 'cat_def_other_expense', 'expense', 93, true, 'heart-handshake', 'gray', now(), now()),
  ('cat_def_other_taxes', 'cat_def_other_expense', 'expense', 94, true, 'file-text', 'gray', now(), now()),

  -- ============ UNIQUE FROM V1 (NO SUFFIX) ============
  -- Social & Relationships
  ('cat_def_social', NULL, 'expense', 190, true, 'users', 'grape', now(), now()),
  ('cat_def_social_gifts', 'cat_def_social', 'expense', 191, true, 'gift', 'grape', now(), now()),
  ('cat_def_social_celebration', 'cat_def_social', 'expense', 192, true, 'heart-handshake', 'grape', now(), now()),
  ('cat_def_social_meeting', 'cat_def_social', 'expense', 193, true, 'users', 'grape', now(), now()),

  -- Work & Business
  ('cat_def_work', NULL, 'expense', 195, true, 'briefcase', 'indigo', now(), now()),
  ('cat_def_work_supplies', 'cat_def_work', 'expense', 196, true, 'pencil', 'indigo', now(), now()),
  ('cat_def_work_equipment', 'cat_def_work', 'expense', 197, true, 'laptop', 'indigo', now(), now()),
  ('cat_def_work_meeting', 'cat_def_work', 'expense', 198, true, 'handshake', 'indigo', now(), now()),

  -- Securities Investment
  ('cat_def_securities', NULL, 'both', 230, true, 'chart-line', 'yellow', now(), now()),
  ('cat_def_securities_buy', 'cat_def_securities', 'expense', 231, true, 'chart-line', 'yellow', now(), now()),
  ('cat_def_securities_sell', 'cat_def_securities', 'income', 232, true, 'chart-line', 'yellow', now(), now()),
  ('cat_def_securities_dividend_cash', 'cat_def_securities', 'income', 233, true, 'coin', 'yellow', now(), now()),
  ('cat_def_securities_dividend_stock', 'cat_def_securities', 'income', 234, true, 'coins', 'yellow', now(), now()),
  ('cat_def_securities_gain', 'cat_def_securities', 'income', 235, true, 'trending-up', 'yellow', now(), now()),
  ('cat_def_securities_fee', 'cat_def_securities', 'expense', 236, true, 'receipt-tax', 'yellow', now(), now()),
  ('cat_def_securities_tax', 'cat_def_securities', 'expense', 237, true, 'building-bank', 'yellow', now(), now()),

  -- Internal Transfers
  ('cat_def_internal', NULL, 'both', 235, true, 'swap', 'cyan', now(), now()),
  ('cat_def_internal_out', 'cat_def_internal', 'expense', 241, true, 'arrow-right', 'cyan', now(), now()),
  ('cat_def_internal_in', 'cat_def_internal', 'income', 242, true, 'arrow-left', 'cyan', now(), now()),
  ('cat_def_internal_topup', 'cat_def_internal', 'expense', 243, true, 'arrow-down', 'cyan', now(), now()),
  ('cat_def_internal_withdraw', 'cat_def_internal', 'income', 244, true, 'arrow-up', 'cyan', now(), now()),

  -- Income extensions
  ('cat_def_income_freelance', 'cat_def_income', 'income', 18, true, 'briefcase', 'green', now(), now()),
  ('cat_def_income_commission', 'cat_def_income', 'income', 19, true, 'percentage', 'green', now(), now()),
  ('cat_def_income_overtime', 'cat_def_income', 'income', 20, true, 'clock', 'green', now(), now()),
  ('cat_def_income_allowance', 'cat_def_income', 'income', 21, true, 'wallet', 'green', now(), now()),
  ('cat_def_income_pension', 'cat_def_income', 'income', 22, true, 'building-bank', 'green', now(), now()),
  ('cat_def_income_scholarship', 'cat_def_income', 'income', 23, true, 'school', 'green', now(), now()),
  ('cat_def_income_affiliate', 'cat_def_income', 'income', 24, true, 'link', 'green', now(), now()),
  ('cat_def_income_cashback', 'cat_def_income', 'income', 25, true, 'rotate', 'green', now(), now()),
  ('cat_def_income_royalties', 'cat_def_income', 'income', 26, true, 'coin', 'green', now(), now()),

  -- Food extensions
  ('cat_def_food_breakfast', 'cat_def_food', 'expense', 16, true, 'sunrise', 'orange', now(), now()),
  ('cat_def_food_lunch', 'cat_def_food', 'expense', 17, true, 'sun', 'orange', now(), now()),
  ('cat_def_food_dinner', 'cat_def_food', 'expense', 18, true, 'moon', 'orange', now(), now()),
  ('cat_def_food_bakery', 'cat_def_food', 'expense', 19, true, 'bread', 'orange', now(), now()),
  ('cat_def_food_fruits', 'cat_def_food', 'expense', 20, true, 'apple', 'orange', now(), now()),

  -- Transport extensions
  ('cat_def_transport_tolls', 'cat_def_transport', 'expense', 38, true, 'road', 'blue', now(), now()),
  ('cat_def_transport_bike_maintenance', 'cat_def_transport', 'expense', 39, true, 'bike', 'blue', now(), now()),
  ('cat_def_transport_car_wash', 'cat_def_transport', 'expense', 40, true, 'droplet', 'blue', now(), now()),
  ('cat_def_transport_train', 'cat_def_transport', 'expense', 41, true, 'train', 'blue', now(), now()),
  ('cat_def_transport_airfare', 'cat_def_transport', 'expense', 42, true, 'plane', 'blue', now(), now()),

  -- Shopping extensions
  ('cat_def_shopping_furniture', 'cat_def_shopping', 'expense', 48, true, 'sofa', 'violet', now(), now()),
  ('cat_def_shopping_appliances', 'cat_def_shopping', 'expense', 49, true, 'device-tv', 'violet', now(), now()),
  ('cat_def_shopping_baby_items', 'cat_def_shopping', 'expense', 50, true, 'baby-carriage', 'violet', now(), now()),
  ('cat_def_shopping_sports', 'cat_def_shopping', 'expense', 51, true, 'ball-football', 'violet', now(), now()),
  ('cat_def_shopping_stationery', 'cat_def_shopping', 'expense', 52, true, 'pencil', 'violet', now(), now()),

  -- Bills extensions
  ('cat_def_bills_hoa_fee', 'cat_def_bills', 'expense', 54, true, 'building-community', 'cyan', now(), now()),
  ('cat_def_bills_security', 'cat_def_bills', 'expense', 55, true, 'shield', 'cyan', now(), now()),
  ('cat_def_bills_parking_fee', 'cat_def_bills', 'expense', 56, true, 'parking', 'cyan', now(), now()),
  ('cat_def_bills_cleaning', 'cat_def_bills', 'expense', 57, true, 'spray-bottle', 'cyan', now(), now()),
  ('cat_def_bills_tv', 'cat_def_bills', 'expense', 58, true, 'device-tv', 'cyan', now(), now()),

  -- Health extensions
  ('cat_def_health_vision', 'cat_def_health', 'expense', 57, true, 'glasses', 'red', now(), now()),
  ('cat_def_health_therapy', 'cat_def_health', 'expense', 58, true, 'brain', 'red', now(), now()),
  ('cat_def_health_checkup', 'cat_def_health', 'expense', 59, true, 'stethoscope', 'red', now(), now()),
  ('cat_def_health_vitamins', 'cat_def_health', 'expense', 60, true, 'pill', 'red', now(), now()),
  ('cat_def_health_mental_health', 'cat_def_health', 'expense', 61, true, 'heart-handshake', 'red', now(), now()),

  -- Entertainment extensions
  ('cat_def_ent_events', 'cat_def_entertainment', 'expense', 68, true, 'calendar-event', 'grape', now(), now()),
  ('cat_def_ent_sports', 'cat_def_entertainment', 'expense', 69, true, 'ball-football', 'grape', now(), now()),
  ('cat_def_ent_museum', 'cat_def_entertainment', 'expense', 70, true, 'building-bank', 'grape', now(), now()),
  ('cat_def_ent_photography', 'cat_def_entertainment', 'expense', 71, true, 'camera', 'grape', now(), now()),
  ('cat_def_ent_outdoor', 'cat_def_entertainment', 'expense', 72, true, 'mountain', 'grape', now(), now()),

  -- Education extensions
  ('cat_def_edu_certification', 'cat_def_education', 'expense', 75, true, 'certificate', 'teal', now(), now()),
  ('cat_def_edu_exam', 'cat_def_education', 'expense', 76, true, 'clipboard-check', 'teal', now(), now()),
  ('cat_def_edu_language', 'cat_def_education', 'expense', 77, true, 'language', 'teal', now(), now()),
  ('cat_def_edu_workshop', 'cat_def_education', 'expense', 78, true, 'users-group', 'teal', now(), now()),
  ('cat_def_edu_training_tools', 'cat_def_education', 'expense', 79, true, 'tool', 'teal', now(), now()),

  -- Beauty extensions
  ('cat_def_beauty_nail', 'cat_def_beauty', 'expense', 22, true, 'brush', 'pink', now(), now()),
  ('cat_def_beauty_skincare', 'cat_def_beauty', 'expense', 23, true, 'droplet', 'pink', now(), now()),
  ('cat_def_beauty_makeup', 'cat_def_beauty', 'expense', 24, true, 'sparkles', 'pink', now(), now()),
  ('cat_def_beauty_barber', 'cat_def_beauty', 'expense', 25, true, 'cut', 'pink', now(), now()),

  -- Family extensions
  ('cat_def_family_parents', 'cat_def_family', 'expense', 85, true, 'users', 'pink', now(), now()),
  ('cat_def_family_elders', 'cat_def_family', 'expense', 86, true, 'heart-handshake', 'pink', now(), now()),
  ('cat_def_family_house_help', 'cat_def_family', 'expense', 87, true, 'home', 'pink', now(), now()),
  ('cat_def_family_allowance', 'cat_def_family', 'expense', 88, true, 'wallet', 'pink', now(), now()),

  -- Pets extensions
  ('cat_def_pets_accessories', 'cat_def_pets', 'expense', 89, true, 'bone', 'lime', now(), now()),
  ('cat_def_pets_training', 'cat_def_pets', 'expense', 90, true, 'school', 'lime', now(), now()),
  ('cat_def_pets_boarding', 'cat_def_pets', 'expense', 91, true, 'home', 'lime', now(), now()),

  -- Financial extensions
  ('cat_def_financial_savings', 'cat_def_financial', 'expense', 94, true, 'pig-money', 'yellow', now(), now()),
  ('cat_def_financial_investment_fund', 'cat_def_financial', 'expense', 95, true, 'chart-pie', 'yellow', now(), now()),
  ('cat_def_financial_retirement', 'cat_def_financial', 'expense', 96, true, 'clock', 'yellow', now(), now()),
  ('cat_def_financial_credit_card_fee', 'cat_def_financial', 'expense', 97, true, 'credit-card', 'yellow', now(), now()),
  ('cat_def_financial_loan_interest', 'cat_def_financial', 'expense', 98, true, 'percentage', 'yellow', now(), now()),
  ('cat_def_financial_tax_payment', 'cat_def_financial', 'expense', 99, true, 'file-text', 'yellow', now(), now()),
  ('cat_def_financial_currency_exchange', 'cat_def_financial', 'expense', 100, true, 'arrows-exchange', 'yellow', now(), now()),

  -- Technology extensions
  ('cat_def_tech_internet_services', 'cat_def_tech', 'expense', 60, true, 'wifi', 'cyan', now(), now()),
  ('cat_def_tech_hosting', 'cat_def_tech', 'expense', 61, true, 'server', 'cyan', now(), now()),
  ('cat_def_tech_domains', 'cat_def_tech', 'expense', 62, true, 'world', 'cyan', now(), now()),
  ('cat_def_tech_accessories', 'cat_def_tech', 'expense', 63, true, 'device-mobile', 'cyan', now(), now()),
  ('cat_def_tech_repairs', 'cat_def_tech', 'expense', 64, true, 'tool', 'cyan', now(), now()),

  -- Other extensions
  ('cat_def_other_fines', 'cat_def_other_expense', 'expense', 95, true, 'alert-circle', 'gray', now(), now()),
  ('cat_def_other_loss_damage', 'cat_def_other_expense', 'expense', 96, true, 'shield-x', 'gray', now(), now()),
  ('cat_def_other_unexpected', 'cat_def_other_expense', 'expense', 97, true, 'question-mark', 'gray', now(), now()),
  ('cat_def_other_misc_services', 'cat_def_other_expense', 'expense', 98, true, 'tools', 'gray', now(), now()),
  ('cat_def_other_membership', 'cat_def_other_expense', 'expense', 99, true, 'id-badge', 'gray', now(), now()),

  -- Social extensions
  ('cat_def_social_charity', 'cat_def_social', 'expense', 194, true, 'heart-handshake', 'grape', now(), now()),
  ('cat_def_social_networking', 'cat_def_social', 'expense', 195, true, 'users-group', 'grape', now(), now()),
  ('cat_def_social_ceremony', 'cat_def_social', 'expense', 196, true, 'confetti', 'grape', now(), now()),

  -- Work extensions
  ('cat_def_work_travel', 'cat_def_work', 'expense', 199, true, 'briefcase', 'indigo', now(), now()),
  ('cat_def_work_business_services', 'cat_def_work', 'expense', 200, true, 'building', 'indigo', now(), now()),
  ('cat_def_work_marketing', 'cat_def_work', 'expense', 201, true, 'speakerphone', 'indigo', now(), now()),

  -- Securities extensions
  ('cat_def_securities_rights_issue', 'cat_def_securities', 'expense', 238, true, 'ticket', 'yellow', now(), now()),
  ('cat_def_securities_ipo', 'cat_def_securities', 'expense', 239, true, 'rocket', 'yellow', now(), now()),
  ('cat_def_securities_bond_coupon', 'cat_def_securities', 'income', 240, true, 'receipt', 'yellow', now(), now()),
  ('cat_def_securities_bond_fee', 'cat_def_securities', 'expense', 241, true, 'receipt-tax', 'yellow', now(), now()),

  -- Insurance
  ('cat_def_insurance', NULL, 'expense', 245, true, 'shield', 'blue', now(), now()),
  ('cat_def_insurance_health', 'cat_def_insurance', 'expense', 246, true, 'heart', 'blue', now(), now()),
  ('cat_def_insurance_vehicle', 'cat_def_insurance', 'expense', 247, true, 'car', 'blue', now(), now()),
  ('cat_def_insurance_home', 'cat_def_insurance', 'expense', 248, true, 'home', 'blue', now(), now()),
  ('cat_def_insurance_life', 'cat_def_insurance', 'expense', 249, true, 'umbrella', 'blue', now(), now()),

  -- Saving goals
  ('cat_def_goals', NULL, 'expense', 250, true, 'target', 'teal', now(), now()),
  ('cat_def_goals_emergency_fund', 'cat_def_goals', 'expense', 251, true, 'shield', 'teal', now(), now()),
  ('cat_def_goals_home_goal', 'cat_def_goals', 'expense', 252, true, 'home', 'teal', now(), now()),
  ('cat_def_goals_education_goal', 'cat_def_goals', 'expense', 253, true, 'school', 'teal', now(), now()),
  ('cat_def_goals_vehicle_goal', 'cat_def_goals', 'expense', 254, true, 'car', 'teal', now(), now()),

  ('cat_sys_rotating_savings_contribution', 'cat_sys_internal', 'expense', 10020, true, 'users', 'cyan', now(), now()),
  ('cat_sys_rotating_savings_payout', 'cat_sys_internal', 'income', 10021, true, 'coins', 'cyan', now(), now())
ON CONFLICT (id) DO NOTHING;

-- +goose Down
-- This is a seed migration, so we don't delete on rollback
-- Users can maintain category data independently of this migration
