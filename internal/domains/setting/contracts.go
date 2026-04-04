package setting

import "github.com/sonbn-225/goen-api-v2/internal/domains/auth"

// HTTP contract models used by handlers and API docs.

type PatchSettingsRequest map[string]any

type UserResponse = auth.User
