package contact

// HTTP contract models used by handlers and API docs.

type CreateContactRequest struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type PatchContactRequest struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}
