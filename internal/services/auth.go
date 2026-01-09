package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error)
	Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error)
}

func defaultUserSettings() map[string]any {
	return map[string]any{
		"locale":           "vi-VN",
		"default_currency": "VND",
		"number_format":    "vi-VN",
		"month_start_day":  1,
		"week_start_day":   1,
		"timezone":         "Asia/Ho_Chi_Minh",
	}
}

type SignupRequest struct {
	Email       string
	Phone       string
	DisplayName string
	Password    string
}

type SigninRequest struct {
	Login    string
	Password string
}

type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
	User        domain.User `json:"user"`
}

type authService struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo domain.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)
	password := req.Password

	// Basic Validation
	if email == "" && phone == "" {
		return nil, errors.New("email or phone is required")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Prepare User Entity
	now := time.Now().UTC()
	userID := uuid.NewString()

	var emailPtr, phonePtr, displayNamePtr *string
	if email != "" {
		emailPtr = &email
	}
	if phone != "" {
		phonePtr = &phone
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName != "" {
		displayNamePtr = &displayName
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := domain.UserWithPassword{
		User: domain.User{
			ID:          userID,
			Email:       emailPtr,
			Phone:       phonePtr,
			DisplayName: displayNamePtr,
			Settings:    defaultUserSettings(),
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		PasswordHash: string(hashBytes),
	}

	// Persist
	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}

	// Generate Token
	token, expiresIn, err := s.generateToken(newUser.User)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        newUser.User,
	}, nil
}

func (s *authService) Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error) {
	login := strings.TrimSpace(req.Login)
	password := req.Password

	if login == "" || password == "" {
		return nil, errors.New("login and password required")
	}

	var user *domain.UserWithPassword
	var err error

	// Try find by email first (simple heuristic: contains @)
	if strings.Contains(login, "@") {
		user, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
	} else {
		user, err = s.userRepo.FindUserByPhone(ctx, login)
	}

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate Token
	token, expiresIn, err := s.generateToken(user.User)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        user.User,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error) {
	if patch == nil {
		patch = map[string]any{}
	}

	// Minimal validation/sanitization for common keys.
	if v, ok := patch["default_currency"]; ok {
		if s, ok := v.(string); ok {
			patch["default_currency"] = strings.ToUpper(strings.TrimSpace(s))
		} else {
			delete(patch, "default_currency")
		}
	}
	if v, ok := patch["locale"]; ok {
		if s, ok := v.(string); ok {
			patch["locale"] = strings.TrimSpace(s)
		} else {
			delete(patch, "locale")
		}
	}
	if v, ok := patch["number_format"]; ok {
		if s, ok := v.(string); ok {
			patch["number_format"] = strings.TrimSpace(s)
		} else {
			delete(patch, "number_format")
		}
	}
	if v, ok := patch["timezone"]; ok {
		if s, ok := v.(string); ok {
			patch["timezone"] = strings.TrimSpace(s)
		} else {
			delete(patch, "timezone")
		}
	}
	if v, ok := patch["month_start_day"]; ok {
		n, ok := v.(float64) // JSON numbers decode as float64
		if !ok {
			delete(patch, "month_start_day")
		} else {
			day := int(n)
			if day < 1 || day > 28 {
				delete(patch, "month_start_day")
			} else {
				patch["month_start_day"] = day
			}
		}
	}
	if v, ok := patch["week_start_day"]; ok {
		n, ok := v.(float64)
		if !ok {
			delete(patch, "week_start_day")
		} else {
			day := int(n)
			if day < 1 || day > 7 {
				delete(patch, "week_start_day")
			} else {
				patch["week_start_day"] = day
			}
		}
	}

	updated, err := s.userRepo.UpdateUserSettings(ctx, userID, patch)
	if err != nil {
		return nil, err
	}

	// Ensure defaults exist for older users.
	if updated.Settings == nil {
		updated.Settings = defaultUserSettings()
	}
	return updated, nil
}

func (s *authService) generateToken(user domain.User) (string, int, error) {
	ttlMinutes := s.cfg.JWTAccessTTLMinutes
	if ttlMinutes <= 0 {
		ttlMinutes = 60
	}
	expiresIn := ttlMinutes * 60
	exp := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"sub": user.ID,
		"iat": time.Now().Unix(),
		"exp": exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn, nil
}
