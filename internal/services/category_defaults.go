package services

// Default categories are seeded per-user the first time they request categories
// and have none yet. This provides a ready-to-use parent/child list.

type defaultCategoryDef struct {
	Key       string
	Name      string
	ParentKey string
	Type      *string // expense|income|both|nil
	SortOrder *int
}

func strPtr(v string) *string { return &v }
func intPtr(v int) *int       { return &v }

func defaultCategories() []defaultCategoryDef {
	// Keep this list relatively small-but-usable.
	// Parents first, then children. Keys are stable identifiers used to link parent/child.
	return []defaultCategoryDef{
		// Expense parents
		{Key: "food", Name: "Food & Drinks", Type: strPtr("expense"), SortOrder: intPtr(10)},
		{Key: "transport", Name: "Transport", Type: strPtr("expense"), SortOrder: intPtr(20)},
		{Key: "shopping", Name: "Shopping", Type: strPtr("expense"), SortOrder: intPtr(30)},
		{Key: "bills", Name: "Bills", Type: strPtr("expense"), SortOrder: intPtr(40)},
		{Key: "health", Name: "Health", Type: strPtr("expense"), SortOrder: intPtr(50)},
		{Key: "entertainment", Name: "Entertainment", Type: strPtr("expense"), SortOrder: intPtr(60)},
		{Key: "education", Name: "Education", Type: strPtr("expense"), SortOrder: intPtr(70)},
		{Key: "other_expense", Name: "Other", Type: strPtr("expense"), SortOrder: intPtr(90)},

		// Expense children
		{Key: "food_groceries", Name: "Groceries", ParentKey: "food", Type: strPtr("expense"), SortOrder: intPtr(11)},
		{Key: "food_eating_out", Name: "Eating out", ParentKey: "food", Type: strPtr("expense"), SortOrder: intPtr(12)},
		{Key: "food_coffee", Name: "Coffee & Tea", ParentKey: "food", Type: strPtr("expense"), SortOrder: intPtr(13)},

		{Key: "transport_gas", Name: "Gas", ParentKey: "transport", Type: strPtr("expense"), SortOrder: intPtr(21)},
		{Key: "transport_taxi", Name: "Taxi / Grab", ParentKey: "transport", Type: strPtr("expense"), SortOrder: intPtr(22)},
		{Key: "transport_public", Name: "Public transit", ParentKey: "transport", Type: strPtr("expense"), SortOrder: intPtr(23)},
		{Key: "transport_parking", Name: "Parking", ParentKey: "transport", Type: strPtr("expense"), SortOrder: intPtr(24)},

		{Key: "shopping_household", Name: "Household", ParentKey: "shopping", Type: strPtr("expense"), SortOrder: intPtr(31)},
		{Key: "shopping_clothes", Name: "Clothes", ParentKey: "shopping", Type: strPtr("expense"), SortOrder: intPtr(32)},
		{Key: "shopping_electronics", Name: "Electronics", ParentKey: "shopping", Type: strPtr("expense"), SortOrder: intPtr(33)},

		{Key: "bills_rent", Name: "Rent", ParentKey: "bills", Type: strPtr("expense"), SortOrder: intPtr(41)},
		{Key: "bills_utilities", Name: "Utilities", ParentKey: "bills", Type: strPtr("expense"), SortOrder: intPtr(42)},
		{Key: "bills_internet", Name: "Internet", ParentKey: "bills", Type: strPtr("expense"), SortOrder: intPtr(43)},
		{Key: "bills_phone", Name: "Phone", ParentKey: "bills", Type: strPtr("expense"), SortOrder: intPtr(44)},

		{Key: "health_medical", Name: "Medical", ParentKey: "health", Type: strPtr("expense"), SortOrder: intPtr(51)},
		{Key: "health_pharmacy", Name: "Pharmacy", ParentKey: "health", Type: strPtr("expense"), SortOrder: intPtr(52)},
		{Key: "health_insurance", Name: "Insurance", ParentKey: "health", Type: strPtr("expense"), SortOrder: intPtr(53)},

		{Key: "ent_movies", Name: "Movies", ParentKey: "entertainment", Type: strPtr("expense"), SortOrder: intPtr(61)},
		{Key: "ent_games", Name: "Games", ParentKey: "entertainment", Type: strPtr("expense"), SortOrder: intPtr(62)},
		{Key: "ent_travel", Name: "Travel", ParentKey: "entertainment", Type: strPtr("expense"), SortOrder: intPtr(63)},

		{Key: "edu_courses", Name: "Courses", ParentKey: "education", Type: strPtr("expense"), SortOrder: intPtr(71)},
		{Key: "edu_books", Name: "Books", ParentKey: "education", Type: strPtr("expense"), SortOrder: intPtr(72)},

		{Key: "other_misc", Name: "Misc", ParentKey: "other_expense", Type: strPtr("expense"), SortOrder: intPtr(91)},

		// Income parent + children
		{Key: "income", Name: "Income", Type: strPtr("income"), SortOrder: intPtr(5)},
		{Key: "income_salary", Name: "Salary", ParentKey: "income", Type: strPtr("income"), SortOrder: intPtr(6)},
		{Key: "income_bonus", Name: "Bonus", ParentKey: "income", Type: strPtr("income"), SortOrder: intPtr(7)},
		{Key: "income_other", Name: "Other income", ParentKey: "income", Type: strPtr("income"), SortOrder: intPtr(8)},
	}
}
