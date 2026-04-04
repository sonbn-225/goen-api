package profile

import (
	"context"
	"strings"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type service struct {
	repo         auth.UserRepository
	hasher       auth.PasswordHasher
	avatarClient AvatarStorage
}

var _ Service = (*service)(nil)

func NewService(repo auth.UserRepository, hasher auth.PasswordHasher, avatarClient AvatarStorage) Service {
	return &service{repo: repo, hasher: hasher, avatarClient: avatarClient}
}

func (s *service) GetMe(ctx context.Context, userID string) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "profile", "operation", "get_me", "user_id", userID)
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		logger.Error("profile_get_me_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load user", err)
	}
	if user == nil {
		logger.Warn("profile_get_me_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindUnauth, "user not found")
	}
	if user.Settings == nil {
		user.Settings = defaultUserSettings()
	}
	logger.Info("profile_get_me_succeeded")
	return user, nil
}

func (s *service) UpdateMyProfile(ctx context.Context, userID string, input auth.UpdateProfileInput) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "profile", "operation", "update_profile", "user_id", userID)
	logger.Info("profile_update_profile_started")

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
			logger.Warn("profile_update_profile_failed", "reason", "duplicate email/phone/username")
			return nil, apperrors.New(apperrors.KindConflict, "email, phone, or username already exists")
		}
		logger.Error("profile_update_profile_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update profile", err)
	}
	if updated == nil {
		logger.Warn("profile_update_profile_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}
	logger.Info("profile_update_profile_succeeded")
	return updated, nil
}

func (s *service) UploadAvatar(ctx context.Context, userID, fileName, contentType string, raw []byte) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "profile", "operation", "upload_avatar", "user_id", userID)
	logger.Info("profile_upload_avatar_started")

	if s.avatarClient == nil {
		return nil, apperrors.New(apperrors.KindInternal, "avatar storage not configured")
	}

	avatarURL, err := s.avatarClient.UploadAvatar(ctx, userID, fileName, contentType, raw)
	if err != nil {
		logger.Error("profile_upload_avatar_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to upload avatar", err)
	}

	updated, err := s.repo.UpdateAvatarURL(ctx, userID, avatarURL)
	if err != nil {
		logger.Error("profile_upload_avatar_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update avatar", err)
	}
	if updated == nil {
		logger.Warn("profile_upload_avatar_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}

	logger.Info("profile_upload_avatar_succeeded")
	return updated, nil
}

func (s *service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "profile", "operation", "change_password", "user_id", userID)
	logger.Info("profile_change_password_started", logx.MaskAttrs("current_password", currentPassword, "new_password", newPassword)...)

	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
		logger.Warn("profile_change_password_failed", "reason", "required fields missing")
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
		logger.Warn("profile_change_password_failed", "reason", "user not found")
		return apperrors.New(apperrors.KindNotFound, "user not found")
	}

	var withPassword *auth.UserWithPassword
	if user.Email != nil && *user.Email != "" {
		withPassword, err = s.repo.FindUserByEmail(ctx, *user.Email)
	} else if user.Phone != nil && *user.Phone != "" {
		withPassword, err = s.repo.FindUserByPhone(ctx, *user.Phone)
	} else {
		withPassword, err = s.repo.FindUserByUsername(ctx, user.Username)
	}
	if err != nil {
		logger.Error("profile_change_password_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to load credentials", err)
	}
	if withPassword == nil {
		logger.Warn("profile_change_password_failed", "reason", "user credentials not found")
		return apperrors.New(apperrors.KindNotFound, "user credentials not found")
	}

	if err := s.hasher.Compare(withPassword.PasswordHash, currentPassword); err != nil {
		logger.Warn("profile_change_password_failed", "reason", "invalid current password")
		return apperrors.New(apperrors.KindUnauth, "current password is invalid")
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return apperrors.Wrap(apperrors.KindInternal, "failed to hash new password", err)
	}
	if err := s.repo.UpdatePasswordHash(ctx, userID, hash); err != nil {
		logger.Error("profile_change_password_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to update password", err)
	}
	logger.Info("profile_change_password_succeeded")
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
