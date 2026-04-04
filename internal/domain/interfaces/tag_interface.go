package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type TagRepository interface {
	CreateTag(ctx context.Context, userID string, tag entity.Tag) error
	GetTag(ctx context.Context, userID string, tagID string) (*entity.Tag, error)
	ListTags(ctx context.Context, userID string) ([]entity.Tag, error)
}

type TagService interface {
	Create(ctx context.Context, userID string, req dto.CreateTagRequest) (*entity.Tag, error)
	Get(ctx context.Context, userID, tagID string) (*entity.Tag, error)
	List(ctx context.Context, userID string) ([]entity.Tag, error)
	GetOrCreateByName(ctx context.Context, userID, name, langHint string) (string, error)
}
