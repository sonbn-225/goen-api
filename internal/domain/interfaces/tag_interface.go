package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type TagRepository interface {
	CreateTag(ctx context.Context, userID uuid.UUID, tag entity.Tag) error
	GetTag(ctx context.Context, userID uuid.UUID, tagID uuid.UUID) (*entity.Tag, error)
	ListTags(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error)
}

type TagService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTagRequest) (*dto.TagResponse, error)
	Get(ctx context.Context, userID, tagID uuid.UUID) (*dto.TagResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.TagResponse, error)
	GetOrCreateByName(ctx context.Context, userID uuid.UUID, name, langHint string) (uuid.UUID, error)
}

