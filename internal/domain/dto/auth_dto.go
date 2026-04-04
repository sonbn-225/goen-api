package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type SignupRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
	User        entity.User `json:"user"`
}
