package setting

import (
	"context"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type service struct {
	repo auth.UserRepository
}

var _ Service = (*service)(nil)

func NewService(repo auth.UserRepository) Service {
	return &service{repo: repo}
}

func (s *service) UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "setting", "operation", "update_settings", "user_id", userID)
	logger.Info("setting_update_settings_started")

	if patch == nil {
		patch = map[string]any{}
	}

	updated, err := s.repo.UpdateUserSettings(ctx, userID, patch)
	if err != nil {
		logger.Error("setting_update_settings_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update settings", err)
	}
	if updated == nil {
		logger.Warn("setting_update_settings_failed", "reason", "user not found")
		return nil, apperrors.New(apperrors.KindNotFound, "user not found")
	}
	logger.Info("setting_update_settings_succeeded")
	return updated, nil
}
