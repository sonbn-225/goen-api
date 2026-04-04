package category

import (
	"context"
	"strings"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Get(ctx context.Context, userID, categoryID string) (*Category, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "category", "operation", "get")
	logger.Info("category_get_started", "user_id", userID, "category_id", categoryID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	id := strings.TrimSpace(categoryID)
	if id == "" {
		return nil, apperrors.New(apperrors.KindValidation, "categoryId is required")
	}

	item, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		logger.Error("category_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get category", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "category not found")
	}

	logger.Info("category_get_succeeded", "category_id", item.ID)
	return item, nil
}

func (s *service) List(ctx context.Context, userID string, txType string) ([]Category, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "category", "operation", "list")
	logger.Info("category_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("category_list_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list categories", err)
	}

	normalizedType := strings.ToLower(strings.TrimSpace(txType))
	if normalizedType == "" {
		logger.Info("category_list_succeeded", "count", len(items))
		return items, nil
	}

	if normalizedType != "income" && normalizedType != "expense" && normalizedType != "both" {
		return nil, apperrors.New(apperrors.KindValidation, "type must be one of income, expense, both")
	}

	filtered := make([]Category, 0, len(items))
	for _, item := range items {
		if item.Type == nil {
			filtered = append(filtered, item)
			continue
		}

		catType := strings.ToLower(strings.TrimSpace(*item.Type))
		switch normalizedType {
		case "income":
			if catType == "income" || catType == "both" {
				filtered = append(filtered, item)
			}
		case "expense":
			if catType == "expense" || catType == "both" {
				filtered = append(filtered, item)
			}
		case "both":
			if catType == "both" {
				filtered = append(filtered, item)
			}
		}
	}

	logger.Info("category_list_succeeded", "count", len(filtered), "type", normalizedType)
	return filtered, nil
}
