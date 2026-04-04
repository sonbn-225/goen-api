package profile

// HTTP contract models used by handlers and API docs.

type PatchProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	Username    *string `json:"username"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResult struct {
	Success bool `json:"success"`
}
