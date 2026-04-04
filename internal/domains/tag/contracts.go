package tag

// HTTP contract models used by handlers and API docs.

type CreateTagRequest struct {
	NameVI *string `json:"name_vi,omitempty"`
	NameEN *string `json:"name_en,omitempty"`
	Color  *string `json:"color,omitempty"`
}
