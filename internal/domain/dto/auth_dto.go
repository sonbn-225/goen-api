package dto

// SignupRequest is the payload for user registration.
// Used in: AuthHandler, AuthService
// SignupRequest is the payload for user registration.
// Used in: AuthHandler, AuthService
type SignupRequest struct {
	Email       string `json:"email"`        // User's email address
	Phone       string `json:"phone"`        // User's phone number
	DisplayName string `json:"display_name"` // User's name shown in UI
	Username    string `json:"username"`     // Unique login ID
	Password    string `json:"password"`     // Plain text password (hashed before storage)
}

// SigninRequest is the payload for user login.
// Used in: AuthHandler, AuthService
type SigninRequest struct {
	Login    string `json:"login" binding:"required"`    // Username or email
	Password string `json:"password" binding:"required"` // User's password
}

// RefreshRequest is the payload for refreshing an access token.
// Used in: AuthHandler, AuthService
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // The valid refresh token
}

// AuthResponse represents the authentication result containing tokens.
// Used in: AuthHandler, AuthService
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`  // JWT access token for API authorization
	RefreshToken string       `json:"refresh_token"` // Token used to get a new access token
	TokenType    string       `json:"token_type"`    // Type of token (e.g., "Bearer")
	ExpiresIn    int          `json:"expires_in"`    // Access token duration in seconds
	User         UserResponse `json:"user"`          // Authenticated user profile information
}
