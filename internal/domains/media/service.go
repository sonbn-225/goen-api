package media

import (
	"context"
	"io"
	"strings"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type service struct {
	storage Storage
}

var _ Service = (*service)(nil)

func NewService(storage Storage) Service {
	return &service{storage: storage}
}

func (s *service) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectInfo, error) {
	if s.storage == nil {
		return nil, ObjectInfo{}, apperrors.New(apperrors.KindInternal, "storage not configured")
	}
	if strings.TrimSpace(bucket) == "" || strings.TrimSpace(key) == "" {
		return nil, ObjectInfo{}, apperrors.New(apperrors.KindValidation, "invalid path")
	}

	obj, info, err := s.storage.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, ObjectInfo{}, apperrors.New(apperrors.KindNotFound, "media not found")
	}
	return obj, info, nil
}
