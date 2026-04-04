package dto

type SignupRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type SigninRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         UserResponse `json:"user"`
}
