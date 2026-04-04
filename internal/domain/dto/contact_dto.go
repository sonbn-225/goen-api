package dto

type CreateContactRequest struct {
	Name      string  `json:"name" binding:"required"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type UpdateContactRequest struct {
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}
