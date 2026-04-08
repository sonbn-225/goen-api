-- +goose Up
-- Seed categories for the application
-- This migration is separate from the baseline schema to allow idempotent seeds
-- Policy: cat_sys_* IDs are reserved system categories and must remain stable.
-- Taxonomy expansion/adjustment should be done via cat_def_* categories only.

CREATE TEMP TABLE seed_categories (
    key text PRIMARY KEY,
    parent_key text NULL,
    type text NOT NULL,
    sort_order integer NOT NULL,
    is_active boolean NOT NULL,
    icon text,
    color text
);

INSERT INTO seed_categories (key, parent_key, type, sort_order, is_active, icon, color)
VALUES
    -- ============ SYSTEM CATEGORIES ============
    ('cat_sys_internal', NULL, 'both', 10000, true, 'settings', 'gray'),
    ('cat_sys_internal_adjustment', 'cat_sys_internal', 'both', 10001, true, 'settings', 'gray'),
    ('cat_sys_internal_sync', 'cat_sys_internal', 'both', 10002, true, 'settings', 'gray'),

    -- ============ INCOME CATEGORIES ============
    ('cat_def_income', NULL, 'income', 5, true, 'cash', 'green'),
    ('cat_def_income_salary', 'cat_def_income', 'income', 6, true, 'cash', 'green'),
    ('cat_def_income_bonus', 'cat_def_income', 'income', 7, true, 'cash', 'green'),
    ('cat_def_income_business', 'cat_def_income', 'income', 9, true, 'briefcase', 'green'),
    ('cat_def_income_invest_interest', 'cat_def_income', 'income', 10, true, 'percent', 'green'),
    ('cat_def_income_invest_dividend', 'cat_def_income', 'income', 11, true, 'chart-line', 'green'),
    ('cat_def_income_rental', 'cat_def_income', 'income', 12, true, 'home', 'green'),
    ('cat_def_income_gift', 'cat_def_income', 'income', 13, true, 'gift', 'green'),
    ('cat_def_income_refund', 'cat_def_income', 'income', 14, true, 'rotate', 'green'),
    ('cat_def_income_sale', 'cat_def_income', 'income', 17, true, 'tag', 'green'),

    -- ============ FOOD & DRINKS ============
    ('cat_def_food', NULL, 'expense', 10, true, 'salad', 'orange'),
    ('cat_def_food_groceries', 'cat_def_food', 'expense', 11, true, 'basket', 'orange'),
    ('cat_def_food_eating_out', 'cat_def_food', 'expense', 12, true, 'utensils', 'orange'),
    ('cat_def_food_coffee', 'cat_def_food', 'expense', 13, true, 'coffee', 'orange'),
    ('cat_def_food_delivery', 'cat_def_food', 'expense', 14, true, 'scooter', 'orange'),
    ('cat_def_food_snacks', 'cat_def_food', 'expense', 15, true, 'cookie', 'orange'),

    -- ============ TRANSPORT ============
    ('cat_def_transport', NULL, 'expense', 20, true, 'car', 'blue'),
    ('cat_def_transport_gas', 'cat_def_transport', 'expense', 31, true, 'gas-pump', 'blue'),
    ('cat_def_transport_taxi', 'cat_def_transport', 'expense', 32, true, 'car', 'blue'),
    ('cat_def_transport_public', 'cat_def_transport', 'expense', 33, true, 'bus', 'blue'),
    ('cat_def_transport_parking', 'cat_def_transport', 'expense', 34, true, 'parking', 'blue'),
    ('cat_def_transport_maintenance', 'cat_def_transport', 'expense', 36, true, 'wrench', 'blue'),
    ('cat_def_transport_insurance', 'cat_def_transport', 'expense', 37, true, 'shield', 'blue'),

    -- ============ SHOPPING ============
    ('cat_def_shopping', NULL, 'expense', 30, true, 'shopping-bag', 'violet'),
    ('cat_def_shopping_household', 'cat_def_shopping', 'expense', 41, true, 'home', 'violet'),
    ('cat_def_shopping_clothes', 'cat_def_shopping', 'expense', 42, true, 'shirt', 'violet'),
    ('cat_def_shopping_electronics', 'cat_def_shopping', 'expense', 43, true, 'laptop', 'violet'),
    ('cat_def_shopping_personal_care', 'cat_def_shopping', 'expense', 44, true, 'spray-bottle', 'violet'),
    ('cat_def_shopping_gifts', 'cat_def_shopping', 'expense', 45, true, 'gift', 'violet'),
    ('cat_def_shopping_online', 'cat_def_shopping', 'expense', 46, true, 'shopping-cart', 'violet'),
    ('cat_def_shopping_jewelry', 'cat_def_shopping', 'expense', 47, true, 'diamond', 'violet'),

    -- ============ BILLS & HOUSING ============
    ('cat_def_bills', NULL, 'expense', 40, true, 'receipt', 'cyan'),
    ('cat_def_bills_rent', 'cat_def_bills', 'expense', 41, true, 'home', 'cyan'),
    ('cat_def_bills_internet', 'cat_def_bills', 'expense', 43, true, 'wifi', 'cyan'),
    ('cat_def_bills_phone', 'cat_def_bills', 'expense', 44, true, 'phone', 'cyan'),
    ('cat_def_bills_mortgage', 'cat_def_bills', 'expense', 45, true, 'building', 'cyan'),
    ('cat_def_bills_repairs', 'cat_def_bills', 'expense', 47, true, 'hammer', 'cyan'),
    ('cat_def_bills_subscriptions', 'cat_def_bills', 'expense', 48, true, 'device-tv', 'cyan'),
    ('cat_def_bills_electricity', 'cat_def_bills', 'expense', 50, true, 'zap', 'cyan'),
    ('cat_def_bills_water', 'cat_def_bills', 'expense', 51, true, 'droplet', 'cyan'),
    ('cat_def_bills_gas', 'cat_def_bills', 'expense', 52, true, 'flame', 'cyan'),
    ('cat_def_bills_waste', 'cat_def_bills', 'expense', 53, true, 'trash', 'cyan'),

    -- ============ HEALTH ============
    ('cat_def_health', NULL, 'expense', 50, true, 'heart', 'red'),
    ('cat_def_health_medical', 'cat_def_health', 'expense', 51, true, 'stethoscope', 'red'),
    ('cat_def_health_pharmacy', 'cat_def_health', 'expense', 52, true, 'pill', 'red'),
    ('cat_def_health_insurance', 'cat_def_health', 'expense', 53, true, 'shield', 'red'),
    ('cat_def_health_dental', 'cat_def_health', 'expense', 54, true, 'tooth', 'red'),
    ('cat_def_health_gym', 'cat_def_health', 'expense', 56, true, 'barbell', 'red'),

    -- ============ ENTERTAINMENT ============
    ('cat_def_entertainment', NULL, 'expense', 60, true, 'mask', 'grape'),
    ('cat_def_ent_movies', 'cat_def_entertainment', 'expense', 61, true, 'film', 'grape'),
    ('cat_def_ent_games', 'cat_def_entertainment', 'expense', 62, true, 'joystick', 'grape'),
    ('cat_def_ent_travel', 'cat_def_entertainment', 'expense', 63, true, 'map', 'grape'),
    ('cat_def_ent_streaming', 'cat_def_entertainment', 'expense', 64, true, 'device-tv', 'grape'),
    ('cat_def_ent_hobbies', 'cat_def_entertainment', 'expense', 66, true, 'palette', 'grape'),
    ('cat_def_ent_music', 'cat_def_entertainment', 'expense', 67, true, 'music', 'grape'),

    -- ============ EDUCATION ============
    ('cat_def_education', NULL, 'expense', 70, true, 'book', 'teal'),
    ('cat_def_edu_courses', 'cat_def_education', 'expense', 71, true, 'book', 'teal'),
    ('cat_def_edu_books', 'cat_def_education', 'expense', 72, true, 'book-open', 'teal'),
    ('cat_def_edu_tuition', 'cat_def_education', 'expense', 73, true, 'school', 'teal'),
    ('cat_def_edu_supplies', 'cat_def_education', 'expense', 74, true, 'pencil', 'teal'),

    -- ============ BEAUTY & GROOMING ============
    ('cat_def_beauty', NULL, 'expense', 18, true, 'sparkles', 'pink'),
    ('cat_def_beauty_haircut', 'cat_def_beauty', 'expense', 19, true, 'scissors', 'pink'),
    ('cat_def_beauty_spa', 'cat_def_beauty', 'expense', 20, true, 'leaf', 'pink'),
    ('cat_def_beauty_cosmetics', 'cat_def_beauty', 'expense', 21, true, 'sparkles', 'pink'),

    -- ============ FAMILY ============
    ('cat_def_family', NULL, 'expense', 80, true, 'users', 'pink'),
    ('cat_def_family_childcare', 'cat_def_family', 'expense', 81, true, 'baby-carriage', 'pink'),
    ('cat_def_family_kids', 'cat_def_family', 'expense', 82, true, 'balloon', 'pink'),
    ('cat_def_family_tuition', 'cat_def_family', 'expense', 83, true, 'school', 'pink'),
    ('cat_def_family_support', 'cat_def_family', 'expense', 84, true, 'heart-handshake', 'pink'),

    -- ============ PETS ============
    ('cat_def_pets', NULL, 'expense', 85, true, 'paw', 'lime'),
    ('cat_def_pets_food', 'cat_def_pets', 'expense', 86, true, 'bone', 'lime'),
    ('cat_def_pets_vet', 'cat_def_pets', 'expense', 87, true, 'stethoscope', 'lime'),
    ('cat_def_pets_grooming', 'cat_def_pets', 'expense', 88, true, 'scissors', 'lime'),

    -- ============ FINANCIAL ============
    ('cat_def_financial', NULL, 'expense', 88, true, 'building-bank', 'yellow'),
    ('cat_def_financial_bank_fees', 'cat_def_financial', 'expense', 89, true, 'receipt-tax', 'yellow'),
    ('cat_def_financial_invest', 'cat_def_financial', 'expense', 91, true, 'chart-line', 'yellow'),
    ('cat_sys_invest_buy', 'cat_def_financial', 'expense', 911, true, 'arrow-down', 'yellow'),
    ('cat_sys_invest_sell', 'cat_def_financial', 'income', 912, true, 'arrow-up', 'yellow'),
    ('cat_sys_invest_fees', 'cat_def_financial', 'expense', 913, true, 'receipt-tax', 'yellow'),
    ('cat_def_financial_insurance', 'cat_def_financial', 'expense', 92, true, 'umbrella', 'yellow'),
    ('cat_def_financial_debt', 'cat_def_financial', 'expense', 93, true, 'arrow-down', 'yellow'),

    -- ============ TECHNOLOGY ============
    ('cat_def_tech', NULL, 'expense', 56, true, 'cpu', 'cyan'),
    ('cat_def_tech_gadgets', 'cat_def_tech', 'expense', 57, true, 'device-mobile', 'cyan'),
    ('cat_def_tech_software', 'cat_def_tech', 'expense', 58, true, 'code', 'cyan'),
    ('cat_def_tech_cloud', 'cat_def_tech', 'expense', 59, true, 'cloud', 'cyan'),

    -- ============ OTHER ============
    ('cat_def_other_expense', NULL, 'expense', 90, true, 'dots', 'gray'),
    ('cat_def_other_fees', 'cat_def_other_expense', 'expense', 92, true, 'receipt-tax', 'gray'),
    ('cat_def_other_donations', 'cat_def_other_expense', 'expense', 93, true, 'heart-handshake', 'gray'),
    ('cat_def_other_taxes', 'cat_def_other_expense', 'expense', 94, true, 'file-text', 'gray'),

    -- ============ UNIQUE FROM V1 (NO SUFFIX) ============
    -- Social & Relationships
    ('cat_def_social', NULL, 'expense', 190, true, 'users', 'grape'),
    ('cat_def_social_gifts', 'cat_def_social', 'expense', 191, true, 'gift', 'grape'),
    ('cat_def_social_celebration', 'cat_def_social', 'expense', 192, true, 'heart-handshake', 'grape'),
    ('cat_def_social_meeting', 'cat_def_social', 'expense', 193, true, 'users', 'grape'),

    -- Work & Business
    ('cat_def_work', NULL, 'expense', 195, true, 'briefcase', 'indigo'),
    ('cat_def_work_supplies', 'cat_def_work', 'expense', 196, true, 'pencil', 'indigo'),
    ('cat_def_work_equipment', 'cat_def_work', 'expense', 197, true, 'laptop', 'indigo'),
    ('cat_def_work_meeting', 'cat_def_work', 'expense', 198, true, 'handshake', 'indigo'),

    -- Securities Investment
    ('cat_def_securities', NULL, 'both', 230, true, 'chart-line', 'yellow'),
    ('cat_def_securities_buy', 'cat_def_securities', 'expense', 231, true, 'chart-line', 'yellow'),
    ('cat_def_securities_sell', 'cat_def_securities', 'income', 232, true, 'chart-line', 'yellow'),
    ('cat_def_securities_dividend_cash', 'cat_def_securities', 'income', 233, true, 'coin', 'yellow'),
    ('cat_def_securities_dividend_stock', 'cat_def_securities', 'income', 234, true, 'coins', 'yellow'),
    ('cat_def_securities_gain', 'cat_def_securities', 'income', 235, true, 'trending-up', 'yellow'),
    ('cat_def_securities_fee', 'cat_def_securities', 'expense', 236, true, 'receipt-tax', 'yellow'),
    ('cat_def_securities_tax', 'cat_def_securities', 'expense', 237, true, 'building-bank', 'yellow'),

    -- Internal Transfers
    ('cat_def_internal', NULL, 'both', 235, true, 'swap', 'cyan'),
    ('cat_def_internal_out', 'cat_def_internal', 'expense', 241, true, 'arrow-right', 'cyan'),
    ('cat_def_internal_in', 'cat_def_internal', 'income', 242, true, 'arrow-left', 'cyan'),
    ('cat_def_internal_topup', 'cat_def_internal', 'expense', 243, true, 'arrow-down', 'cyan'),
    ('cat_def_internal_withdraw', 'cat_def_internal', 'income', 244, true, 'arrow-up', 'cyan'),

    -- Income extensions
    ('cat_def_income_freelance', 'cat_def_income', 'income', 18, true, 'briefcase', 'green'),
    ('cat_def_income_commission', 'cat_def_income', 'income', 19, true, 'percentage', 'green'),
    ('cat_def_income_overtime', 'cat_def_income', 'income', 20, true, 'clock', 'green'),
    ('cat_def_income_allowance', 'cat_def_income', 'income', 21, true, 'wallet', 'green'),
    ('cat_def_income_pension', 'cat_def_income', 'income', 22, true, 'building-bank', 'green'),
    ('cat_def_income_scholarship', 'cat_def_income', 'income', 23, true, 'school', 'green'),
    ('cat_def_income_affiliate', 'cat_def_income', 'income', 24, true, 'link', 'green'),
    ('cat_def_income_cashback', 'cat_def_income', 'income', 25, true, 'rotate', 'green'),
    ('cat_def_income_royalties', 'cat_def_income', 'income', 26, true, 'coin', 'green'),

    -- Food extensions
    ('cat_def_food_breakfast', 'cat_def_food', 'expense', 16, true, 'sunrise', 'orange'),
    ('cat_def_food_lunch', 'cat_def_food', 'expense', 17, true, 'sun', 'orange'),
    ('cat_def_food_dinner', 'cat_def_food', 'expense', 18, true, 'moon', 'orange'),
    ('cat_def_food_bakery', 'cat_def_food', 'expense', 19, true, 'bread', 'orange'),
    ('cat_def_food_fruits', 'cat_def_food', 'expense', 20, true, 'apple', 'orange'),

    -- Transport extensions
    ('cat_def_transport_tolls', 'cat_def_transport', 'expense', 38, true, 'road', 'blue'),
    ('cat_def_transport_bike_maintenance', 'cat_def_transport', 'expense', 39, true, 'bike', 'blue'),
    ('cat_def_transport_car_wash', 'cat_def_transport', 'expense', 40, true, 'droplet', 'blue'),
    ('cat_def_transport_train', 'cat_def_transport', 'expense', 41, true, 'train', 'blue'),
    ('cat_def_transport_airfare', 'cat_def_transport', 'expense', 42, true, 'plane', 'blue'),

    -- Shopping extensions
    ('cat_def_shopping_furniture', 'cat_def_shopping', 'expense', 48, true, 'sofa', 'violet'),
    ('cat_def_shopping_appliances', 'cat_def_shopping', 'expense', 49, true, 'device-tv', 'violet'),
    ('cat_def_shopping_baby_items', 'cat_def_shopping', 'expense', 50, true, 'baby-carriage', 'violet'),
    ('cat_def_shopping_sports', 'cat_def_shopping', 'expense', 51, true, 'ball-football', 'violet'),
    ('cat_def_shopping_stationery', 'cat_def_shopping', 'expense', 52, true, 'pencil', 'violet'),

    -- Bills extensions
    ('cat_def_bills_hoa_fee', 'cat_def_bills', 'expense', 54, true, 'building-community', 'cyan'),
    ('cat_def_bills_security', 'cat_def_bills', 'expense', 55, true, 'shield', 'cyan'),
    ('cat_def_bills_parking_fee', 'cat_def_bills', 'expense', 56, true, 'parking', 'cyan'),
    ('cat_def_bills_cleaning', 'cat_def_bills', 'expense', 57, true, 'spray-bottle', 'cyan'),
    ('cat_def_bills_tv', 'cat_def_bills', 'expense', 58, true, 'device-tv', 'cyan'),

    -- Health extensions
    ('cat_def_health_vision', 'cat_def_health', 'expense', 57, true, 'glasses', 'red'),
    ('cat_def_health_therapy', 'cat_def_health', 'expense', 58, true, 'brain', 'red'),
    ('cat_def_health_checkup', 'cat_def_health', 'expense', 59, true, 'stethoscope', 'red'),
    ('cat_def_health_vitamins', 'cat_def_health', 'expense', 60, true, 'pill', 'red'),
    ('cat_def_health_mental_health', 'cat_def_health', 'expense', 61, true, 'heart-handshake', 'red'),

    -- Entertainment extensions
    ('cat_def_ent_events', 'cat_def_entertainment', 'expense', 68, true, 'calendar-event', 'grape'),
    ('cat_def_ent_sports', 'cat_def_entertainment', 'expense', 69, true, 'ball-football', 'grape'),
    ('cat_def_ent_museum', 'cat_def_entertainment', 'expense', 70, true, 'building-bank', 'grape'),
    ('cat_def_ent_photography', 'cat_def_entertainment', 'expense', 71, true, 'camera', 'grape'),
    ('cat_def_ent_outdoor', 'cat_def_entertainment', 'expense', 72, true, 'mountain', 'grape'),

    -- Education extensions
    ('cat_def_edu_certification', 'cat_def_education', 'expense', 75, true, 'certificate', 'teal'),
    ('cat_def_edu_exam', 'cat_def_education', 'expense', 76, true, 'clipboard-check', 'teal'),
    ('cat_def_edu_language', 'cat_def_education', 'expense', 77, true, 'language', 'teal'),
    ('cat_def_edu_workshop', 'cat_def_education', 'expense', 78, true, 'users-group', 'teal'),
    ('cat_def_edu_training_tools', 'cat_def_education', 'expense', 79, true, 'tool', 'teal'),

    -- Beauty extensions
    ('cat_def_beauty_nail', 'cat_def_beauty', 'expense', 22, true, 'brush', 'pink'),
    ('cat_def_beauty_skincare', 'cat_def_beauty', 'expense', 23, true, 'droplet', 'pink'),
    ('cat_def_beauty_makeup', 'cat_def_beauty', 'expense', 24, true, 'sparkles', 'pink'),
    ('cat_def_beauty_barber', 'cat_def_beauty', 'expense', 25, true, 'cut', 'pink'),

    -- Family extensions
    ('cat_def_family_parents', 'cat_def_family', 'expense', 85, true, 'users', 'pink'),
    ('cat_def_family_elders', 'cat_def_family', 'expense', 86, true, 'heart-handshake', 'pink'),
    ('cat_def_family_house_help', 'cat_def_family', 'expense', 87, true, 'home', 'pink'),
    ('cat_def_family_allowance', 'cat_def_family', 'expense', 88, true, 'wallet', 'pink'),

    -- Pets extensions
    ('cat_def_pets_accessories', 'cat_def_pets', 'expense', 89, true, 'bone', 'lime'),
    ('cat_def_pets_training', 'cat_def_pets', 'expense', 90, true, 'school', 'lime'),
    ('cat_def_pets_boarding', 'cat_def_pets', 'expense', 91, true, 'home', 'lime'),

    -- Financial extensions
    ('cat_def_financial_savings', 'cat_def_financial', 'expense', 94, true, 'pig-money', 'yellow'),
    ('cat_def_financial_investment_fund', 'cat_def_financial', 'expense', 95, true, 'chart-pie', 'yellow'),
    ('cat_def_financial_retirement', 'cat_def_financial', 'expense', 96, true, 'clock', 'yellow'),
    ('cat_def_financial_credit_card_fee', 'cat_def_financial', 'expense', 97, true, 'credit-card', 'yellow'),
    ('cat_def_financial_loan_interest', 'cat_def_financial', 'expense', 98, true, 'percentage', 'yellow'),
    ('cat_def_financial_tax_payment', 'cat_def_financial', 'expense', 99, true, 'file-text', 'yellow'),
    ('cat_def_financial_currency_exchange', 'cat_def_financial', 'expense', 100, true, 'arrows-exchange', 'yellow'),

    -- Technology extensions
    ('cat_def_tech_internet_services', 'cat_def_tech', 'expense', 60, true, 'wifi', 'cyan'),
    ('cat_def_tech_hosting', 'cat_def_tech', 'expense', 61, true, 'server', 'cyan'),
    ('cat_def_tech_domains', 'cat_def_tech', 'expense', 62, true, 'world', 'cyan'),
    ('cat_def_tech_accessories', 'cat_def_tech', 'expense', 63, true, 'device-mobile', 'cyan'),
    ('cat_def_tech_repairs', 'cat_def_tech', 'expense', 64, true, 'tool', 'cyan'),

    -- Other extensions
    ('cat_def_other_fines', 'cat_def_other_expense', 'expense', 95, true, 'alert-circle', 'gray'),
    ('cat_def_other_loss_damage', 'cat_def_other_expense', 'expense', 96, true, 'shield-x', 'gray'),
    ('cat_def_other_unexpected', 'cat_def_other_expense', 'expense', 97, true, 'question-mark', 'gray'),
    ('cat_def_other_misc_services', 'cat_def_other_expense', 'expense', 98, true, 'tools', 'gray'),
    ('cat_def_other_membership', 'cat_def_other_expense', 'expense', 99, true, 'id-badge', 'gray'),

    -- Social extensions
    ('cat_def_social_charity', 'cat_def_social', 'expense', 194, true, 'heart-handshake', 'grape'),
    ('cat_def_social_networking', 'cat_def_social', 'expense', 195, true, 'users-group', 'grape'),
    ('cat_def_social_ceremony', 'cat_def_social', 'expense', 196, true, 'confetti', 'grape'),

    -- Work extensions
    ('cat_def_work_travel', 'cat_def_work', 'expense', 199, true, 'briefcase', 'indigo'),
    ('cat_def_work_business_services', 'cat_def_work', 'expense', 200, true, 'building', 'indigo'),
    ('cat_def_work_marketing', 'cat_def_work', 'expense', 201, true, 'speakerphone', 'indigo'),

    -- Securities extensions
    ('cat_def_securities_rights_issue', 'cat_def_securities', 'expense', 238, true, 'ticket', 'yellow'),
    ('cat_def_securities_ipo', 'cat_def_securities', 'expense', 239, true, 'rocket', 'yellow'),
    ('cat_def_securities_bond_coupon', 'cat_def_securities', 'income', 240, true, 'receipt', 'yellow'),
    ('cat_def_securities_bond_fee', 'cat_def_securities', 'expense', 241, true, 'receipt-tax', 'yellow'),

    -- Insurance
    ('cat_def_insurance', NULL, 'expense', 245, true, 'shield', 'blue'),
    ('cat_def_insurance_health', 'cat_def_insurance', 'expense', 246, true, 'heart', 'blue'),
    ('cat_def_insurance_vehicle', 'cat_def_insurance', 'expense', 247, true, 'car', 'blue'),
    ('cat_def_insurance_home', 'cat_def_insurance', 'expense', 248, true, 'home', 'blue'),
    ('cat_def_insurance_life', 'cat_def_insurance', 'expense', 249, true, 'umbrella', 'blue'),

    -- Saving goals
    ('cat_def_goals', NULL, 'expense', 250, true, 'target', 'teal'),
    ('cat_def_goals_emergency_fund', 'cat_def_goals', 'expense', 251, true, 'shield', 'teal'),
    ('cat_def_goals_home_goal', 'cat_def_goals', 'expense', 252, true, 'home', 'teal'),
    ('cat_def_goals_education_goal', 'cat_def_goals', 'expense', 253, true, 'school', 'teal'),
    ('cat_def_goals_vehicle_goal', 'cat_def_goals', 'expense', 254, true, 'car', 'teal'),

    ('cat_sys_rotating_savings_contribution', 'cat_sys_internal', 'expense', 10020, true, 'users', 'cyan'),
    ('cat_sys_rotating_savings_payout', 'cat_sys_internal', 'income', 10021, true, 'coins', 'cyan');


INSERT INTO categories (key, type, sort_order, is_active, icon, color)
SELECT key, type, sort_order, is_active, icon, color
FROM seed_categories
ON CONFLICT (key) DO UPDATE
SET type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active,
    icon = EXCLUDED.icon,
    color = EXCLUDED.color;

UPDATE categories c
SET parent_category_id = p.id
FROM seed_categories s
LEFT JOIN categories p ON p.key = s.parent_key
WHERE c.key = s.key
  AND c.parent_category_id IS DISTINCT FROM p.id;

-- Xoá tường minh thay vì dùng ON COMMIT DROP để đảm bảo an toàn với cơ chế transaction của goose
DROP TABLE seed_categories;

-- +goose Down
-- This is a seed migration, so we don't delete on rollback
-- Users can maintain category data independently of this migration