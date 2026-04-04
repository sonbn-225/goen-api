package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo       UserRepository
	hasher     PasswordHasher
	issuer     TokenIssuer
	expiresInS int
}

var _ Service = (*service)(nil)

func NewService(repo UserRepository, hasher PasswordHasher, issuer TokenIssuer, accessTTLMinutes int) Service {
	expiresIn := accessTTLMinutes * 60
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	return &service{repo: repo, hasher: hasher, issuer: issuer, expiresInS: expiresIn}
}

func (s *service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "signup")
	logger.Info("auth_signup_started", "has_email", strings.TrimSpace(req.Email) != "", "has_phone", strings.TrimSpace(req.Phone) != "", "username", strings.TrimSpace(req.Username))

	email := normalizeEmail(req.Email)
	phone := strings.TrimSpace(req.Phone)
	password := req.Password
	username := strings.ToLower(strings.TrimSpace(req.Username))
	if username == "" && email != "" {
		username = strings.Split(email, "@")[0]
	}

	if email == "" && phone == "" {
		logger.Warn("auth_signup_failed", "reason", "email or phone is required")
		return nil, apperrors.New(apperrors.KindValidation, "email or phone is required")
	}
	if username == "" {
		return nil, apperrors.New(apperrors.KindValidation, "username is required")
	}
	if len(password) < 8 {
		logger.Warn("auth_signup_failed", logx.MaskAttrs("reason", "weak password", "password", password)...)
		return nil, apperrors.New(apperrors.KindValidation, "password must be at least 8 characters")
	}

	if email != "" {
		existingByEmail, err := s.repo.FindUserByEmail(ctx, email)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.KindInternal, "failed to check existing email", err)
		}
		if existingByEmail != nil {
			return nil, apperrors.New(apperrors.KindConflict, "email already exists")
		}
	}

	if phone != "" {
		existingByPhone, err := s.repo.FindUserByPhone(ctx, phone)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.KindInternal, "failed to check existing phone", err)
		}
		if existingByPhone != nil {
			return nil, apperrors.New(apperrors.KindConflict, "phone already exists")
		}
	}

	existingByUsername, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to check existing username", err)
	}
	if existingByUsername != nil {
		return nil, apperrors.New(apperrors.KindConflict, "username already exists")
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to hash password", err)
	}

	now := time.Now().UTC()
	var emailPtr, phonePtr, displayNamePtr *string
	if email != "" {
		emailPtr = &email
	}
	if phone != "" {
		phonePtr = &phone
	}
	if d := strings.TrimSpace(req.DisplayName); d != "" {
		displayNamePtr = &d
	}

	newUser := UserWithPassword{
		User: User{
			ID:          uuid.NewString(),
			Username:    username,
			Email:       emailPtr,
			Phone:       phonePtr,
			DisplayName: displayNamePtr,
			Settings:    defaultUserSettings(),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		PasswordHash: hash,
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		if isUniqueViolation(err) {
			logger.Warn("auth_signup_conflict", "reason", "duplicate email/phone/username")
			return nil, apperrors.New(apperrors.KindConflict, "email, phone, or username already exists")
		}
		logger.Error("auth_signup_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to create user", err)
	}

	token, err := s.issuer.Issue(newUser.ID)
	if err != nil {
		logger.Error("auth_signup_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to issue token", err)
	}

	logger.Info("auth_signup_succeeded", "user_id", newUser.ID)

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.expiresInS,
		User:        newUser.User,
	}, nil
}

func (s *service) Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "signin")
	login := strings.TrimSpace(req.Login)
	logger.Info("auth_signin_started", "login", login)
	if login == "" || req.Password == "" {
		logger.Warn("auth_signin_failed", logx.MaskAttrs("reason", "login and password are required", "password", req.Password)...)
		return nil, apperrors.New(apperrors.KindValidation, "login and password are required")
	}

	var (
		userWithPassword *UserWithPassword
		err              error
	)
	if strings.Contains(login, "@") {
		userWithPassword, err = s.repo.FindUserByEmail(ctx, normalizeEmail(login))
	} else if isPhoneLike(login) {
		userWithPassword, err = s.repo.FindUserByPhone(ctx, login)
	} else {
		userWithPassword, err = s.repo.FindUserByUsername(ctx, strings.ToLower(login))
	}
	if err != nil {
		logger.Error("auth_signin_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load user", err)
	}
	if userWithPassword == nil {
		logger.Warn("auth_signin_failed", "reason", "invalid credentials")
		return nil, apperrors.New(apperrors.KindUnauth, "invalid credentials")
	}

	if err := s.hasher.Compare(userWithPassword.PasswordHash, req.Password); err != nil {
		logger.Warn("auth_signin_failed", logx.MaskAttrs("reason", "invalid credentials", "password", req.Password)...)
		return nil, apperrors.New(apperrors.KindUnauth, "invalid credentials")
	}

	token, err := s.issuer.Issue(userWithPassword.ID)
	if err != nil {
		logger.Error("auth_signin_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to issue token", err)
	}

	logger.Info("auth_signin_succeeded", "user_id", userWithPassword.ID)

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.expiresInS,
		User:        userWithPassword.User,
	}, nil
}

func (s *service) Refresh(ctx context.Context, userID string) (*AuthResponse, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "refresh", "user_id", userID)
	logger.Info("auth_refresh_started")

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		logger.Error("auth_refresh_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load user", err)
	}
	if user == nil {
		logger.Warn("auth_refresh_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindUnauth, "user not found")
	}

	token, err := s.issuer.Issue(userID)
	if err != nil {
		logger.Error("auth_refresh_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to issue token", err)
	}
	logger.Info("auth_refresh_succeeded")

	return &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.expiresInS,
		User:        *user,
	}, nil
}

func (s *service) GetMe(ctx context.Context, userID string) (*User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "get_me", "user_id", userID)
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		logger.Error("auth_get_me_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load user", err)
	}
	if user == nil {
		logger.Warn("auth_get_me_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindUnauth, "user not found")
	}
	if user.Settings == nil {
		user.Settings = defaultUserSettings()
	}
	logger.Info("auth_get_me_succeeded")
	return user, nil
}

func (s *service) UpdateMyProfile(ctx context.Context, userID string, input UpdateProfileInput) (*User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "update_profile", "user_id", userID)
	logger.Info("auth_update_profile_started")

	if input.Email != nil {
		n := normalizeEmail(*input.Email)
		if n != "" && !strings.Contains(n, "@") {
			return nil, apperrors.New(apperrors.KindValidation, "invalid email format")
		}
		input.Email = &n
	}
	if input.Username != nil {
		n := strings.ToLower(strings.TrimSpace(*input.Username))
		if n == "" {
			return nil, apperrors.New(apperrors.KindValidation, "username cannot be empty")
		}
		input.Username = &n
	}
	if input.Phone != nil {
		n := strings.TrimSpace(*input.Phone)
		input.Phone = &n
	}
	if input.DisplayName != nil {
		n := strings.TrimSpace(*input.DisplayName)
		input.DisplayName = &n
	}

	updated, err := s.repo.UpdateUserProfile(ctx, userID, input)
	if err != nil {
		if isUniqueViolation(err) {
			logger.Warn("auth_update_profile_failed", "reason", "duplicate email/phone/username")
			return nil, apperrors.New(apperrors.KindConflict, "email, phone, or username already exists")
		}
		logger.Error("auth_update_profile_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update profile", err)
	}
	if updated == nil {
		logger.Warn("auth_update_profile_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}
	logger.Info("auth_update_profile_succeeded")
	return updated, nil
}

func (s *service) UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "update_settings", "user_id", userID)
	logger.Info("auth_update_settings_started")
	if patch != nil {
		logger.Info("auth_update_settings_payload", logx.MaskAttrs("patch", patch)...)
	}

	if patch == nil {
		patch = map[string]any{}
	}
	updated, err := s.repo.UpdateUserSettings(ctx, userID, patch)
	if err != nil {
		logger.Error("auth_update_settings_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update settings", err)
	}
	if updated == nil {
		logger.Warn("auth_update_settings_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}
	logger.Info("auth_update_settings_succeeded")
	return updated, nil
}

func (s *service) UpdateMyAvatar(ctx context.Context, userID, avatarURL string) (*User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "update_avatar", "user_id", userID)
	logger.Info("auth_update_avatar_started")

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return nil, apperrors.New(apperrors.KindValidation, "avatar_url is required")
	}

	updated, err := s.repo.UpdateAvatarURL(ctx, userID, avatarURL)
	if err != nil {
		logger.Error("auth_update_avatar_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update avatar", err)
	}
	if updated == nil {
		logger.Warn("auth_update_avatar_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}

	logger.Info("auth_update_avatar_succeeded", "avatar_url", avatarURL)
	return updated, nil
}

func (s *service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "auth", "operation", "change_password", "user_id", userID)
	logger.Info("auth_change_password_started", logx.MaskAttrs("current_password", currentPassword, "new_password", newPassword)...)

	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
		logger.Warn("auth_change_password_failed", "reason", "required fields missing")
		return apperrors.New(apperrors.KindValidation, "required fields missing")
	}
	if len(newPassword) < 8 {
		return apperrors.New(apperrors.KindValidation, "new password must be at least 8 characters")
	}
	if !hasUpperLowerDigit(newPassword) {
		return apperrors.New(apperrors.KindValidation, "password must contain uppercase, lowercase, and digits")
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return apperrors.Wrap(apperrors.KindInternal, "failed to load user", err)
	}
	if user == nil {
		logger.Warn("auth_change_password_failed", "reason", "user not found")
		return apperrors.New(apperrors.KindNotFound, "user not found")
	}

	var withPassword *UserWithPassword
	if user.Email != nil && *user.Email != "" {
		withPassword, err = s.repo.FindUserByEmail(ctx, *user.Email)
	} else if user.Phone != nil && *user.Phone != "" {
		withPassword, err = s.repo.FindUserByPhone(ctx, *user.Phone)
	} else {
		withPassword, err = s.repo.FindUserByUsername(ctx, user.Username)
	}
	if err != nil {
		logger.Error("auth_change_password_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to load credentials", err)
	}
	if withPassword == nil {
		logger.Warn("auth_change_password_failed", "reason", "user credentials not found")
		return apperrors.New(apperrors.KindNotFound, "user credentials not found")
	}

	if err := s.hasher.Compare(withPassword.PasswordHash, currentPassword); err != nil {
		logger.Warn("auth_change_password_failed", "reason", "invalid current password")
		return apperrors.New(apperrors.KindUnauth, "current password is invalid")
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return apperrors.Wrap(apperrors.KindInternal, "failed to hash new password", err)
	}
	if err := s.repo.UpdatePasswordHash(ctx, userID, hash); err != nil {
		logger.Error("auth_change_password_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to update password", err)
	}
	logger.Info("auth_change_password_succeeded")
	return nil
}

func defaultUserSettings() map[string]any {
	return map[string]any{
		"locale":           "en-US",
		"default_currency": "VND",
		"timezone":         "Asia/Ho_Chi_Minh",
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isPhoneLike(login string) bool {
	if login == "" {
		return false
	}
	if login[0] == '+' {
		return true
	}
	for _, ch := range login {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func hasUpperLowerDigit(password string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		} else if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
