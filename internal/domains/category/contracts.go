package category

// HTTP contract models used by handlers and API docs.

type ListCategoriesQuery struct {
	Type *string `json:"type,omitempty"`
}
