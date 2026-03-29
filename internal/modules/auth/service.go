package auth

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// SignupRequest contains signup parameters.
type SignupRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// SigninRequest contains signin parameters.
type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful auth.
type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
	User        domain.User `json:"user"`
}

// Service handles authentication business logic.
type Service struct {
	userRepo domain.UserRepository
	cfg      *config.Config
	s3       *storage.S3Client
}

// NewService creates a new auth service.
func NewService(userRepo domain.UserRepository, cfg *config.Config, s3 *storage.S3Client) *Service {
	return &Service{
		userRepo: userRepo,
		cfg:      cfg,
		s3:       s3,
	}
}

func defaultUserSettings() map[string]any {
	return map[string]any{
		"locale":           "vi-VN",
		"default_currency": "VND",
		"number_format":    "vi-VN",
		"month_start_day":  1,
		"week_start_day":   1,
		"timezone":         "Asia/Ho_Chi_Minh",
		"rotating_savings_term": "hui",
	}
}

// Signup creates a new user account.
func (s *Service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)
	password := req.Password

	if email == "" && phone == "" {
		return nil, apperrors.Validation("email or phone is required", nil)
	}
	if len(password) < 8 {
		return nil, apperrors.Validation("password must be at least 8 characters", nil)
	}

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

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			return nil, apperrors.Wrap(apperrors.KindConflict, "user already exists", err)
		}
		return nil, err
	}

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

// Signin authenticates a user.
func (s *Service) Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error) {
	login := strings.TrimSpace(req.Login)
	password := req.Password

	if login == "" || password == "" {
		return nil, apperrors.Validation("login and password required", nil)
	}

	var user *domain.UserWithPassword
	var err error

	if strings.Contains(login, "@") {
		user, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
	} else {
		user, err = s.userRepo.FindUserByPhone(ctx, login)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "invalid credentials", apperrors.ErrInvalidCredentials)
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperrors.Wrap(apperrors.KindUnauthorized, "invalid credentials", apperrors.ErrInvalidCredentials)
	}

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

// Refresh issues a new access token for the current user.
func (s *Service) Refresh(ctx context.Context, userID string) (*AuthResponse, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "user not found", err)
		}
		return nil, err
	}

	token, expiresIn, err := s.generateToken(*user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        *user,
	}, nil
}

// GetMe returns the current user.
func (s *Service) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "user not found", err)
		}
		return nil, err
	}
	return user, nil
}

// UpdateMySettings updates user settings.
func (s *Service) UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error) {
	if patch == nil {
		patch = map[string]any{}
	}

	// Sanitize settings
	s.sanitizeSettings(patch)

	updated, err := s.userRepo.UpdateUserSettings(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "user not found", err)
		}
		return nil, err
	}

	if updated.Settings == nil {
		updated.Settings = defaultUserSettings()
	}
	return updated, nil
}

// UploadAvatar uploads a profile image to SeaweedFS and updates the user's avatar_url.
func (s *Service) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (*domain.User, error) {
	if s.s3 == nil {
		return nil, apperrors.Wrap(apperrors.KindNotFound, "storage not configured", nil)
	}

	objectKey, err := s.s3.UploadAvatar(ctx, userID, file)
	if err != nil {
		return nil, err
	}

	// Build proxied URL pointing to our /media/... endpoint.
	avatarURL := s.s3.AvatarURL(s.cfg.S3PublicBaseURL, objectKey)

	updated, err := s.userRepo.UpdateUserProfile(ctx, userID, domain.UpdateUserParams{
		AvatarURL: &avatarURL,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "user not found", err)
		}
		return nil, err
	}
	return updated, nil
}

// UpdateMyProfile patches display_name, email, or phone.
func (s *Service) UpdateMyProfile(ctx context.Context, userID string, displayName, email, phone *string) (*domain.User, error) {
	params := domain.UpdateUserParams{}

	if displayName != nil {
		v := strings.TrimSpace(*displayName)
		params.DisplayName = &v
	}

	if email != nil {
		v := strings.ToLower(strings.TrimSpace(*email))
		if v != "" && !strings.Contains(v, "@") {
			return nil, apperrors.Validation("invalid email format", nil)
		}
		params.Email = &v
	}

	if phone != nil {
		v := strings.TrimSpace(*phone)
		params.Phone = &v
	}

	updated, err := s.userRepo.UpdateUserProfile(ctx, userID, params)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindUnauthorized, "user not found", err)
		}
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			return nil, apperrors.Wrap(apperrors.KindConflict, "email or phone already in use", err)
		}
		return nil, err
	}
	return updated, nil
}

// ChangePassword updates the user's password after verifying the current one.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return apperrors.Validation("new password must be at least 8 characters", nil)
	}
	// Check password strength: at least 1 uppercase, 1 lowercase, 1 digit
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, char := range newPassword {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return apperrors.Validation("password must contain uppercase, lowercase, and digits", nil)
	}

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// We need the hash, so find user with password
	var uWithPass *domain.UserWithPassword
	if user.Email != nil {
		uWithPass, err = s.userRepo.FindUserByEmail(ctx, *user.Email)
	} else if user.Phone != nil {
		uWithPass, err = s.userRepo.FindUserByPhone(ctx, *user.Phone)
	} else {
		return errors.New("cannot identify user for password change")
	}

	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(uWithPass.PasswordHash), []byte(currentPassword)); err != nil {
		return apperrors.Wrap(apperrors.KindUnauthorized, "invalid current password", apperrors.ErrInvalidCredentials)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashStr := string(newHash)
	_, err = s.userRepo.UpdateUserProfile(ctx, userID, domain.UpdateUserParams{
		PasswordHash: &hashStr,
	})
	return err
}

func (s *Service) sanitizeSettings(patch map[string]any) {
	if v, ok := patch["default_currency"]; ok {
		if str, ok := v.(string); ok {
			patch["default_currency"] = strings.ToUpper(strings.TrimSpace(str))
		} else {
			delete(patch, "default_currency")
		}
	}
	if v, ok := patch["locale"]; ok {
		if str, ok := v.(string); ok {
			patch["locale"] = strings.TrimSpace(str)
		} else {
			delete(patch, "locale")
		}
	}
	if v, ok := patch["number_format"]; ok {
		if str, ok := v.(string); ok {
			patch["number_format"] = strings.TrimSpace(str)
		} else {
			delete(patch, "number_format")
		}
	}
	if v, ok := patch["timezone"]; ok {
		if str, ok := v.(string); ok {
			patch["timezone"] = strings.TrimSpace(str)
		} else {
			delete(patch, "timezone")
		}
	}
	if v, ok := patch["month_start_day"]; ok {
		if n, ok := v.(float64); ok {
			day := int(n)
			if day >= 1 && day <= 28 {
				patch["month_start_day"] = day
			} else {
				delete(patch, "month_start_day")
			}
		} else {
			delete(patch, "month_start_day")
		}
	}
	if v, ok := patch["week_start_day"]; ok {
		if n, ok := v.(float64); ok {
			day := int(n)
			if day >= 1 && day <= 7 {
				patch["week_start_day"] = day
			} else {
				delete(patch, "week_start_day")
			}
		} else {
			delete(patch, "week_start_day")
		}
	}
	if v, ok := patch["rotating_savings_term"]; ok {
		if str, ok := v.(string); ok {
			val := strings.ToLower(strings.TrimSpace(str))
			if val == "hui" || val == "ho" || val == "phuong" {
				patch["rotating_savings_term"] = val
			} else {
				delete(patch, "rotating_savings_term")
			}
		} else {
			delete(patch, "rotating_savings_term")
		}
	}
}

func (s *Service) generateToken(user domain.User) (string, int, error) {
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

