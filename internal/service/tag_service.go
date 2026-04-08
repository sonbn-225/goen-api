package service
 
import (
	"context"
	"errors"
	"strings"
 
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)
 
type TagService struct {
	repo interfaces.TagRepository
}
 
func NewTagService(repo interfaces.TagRepository) *TagService {
	return &TagService{repo: repo}
}
 
func (s *TagService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateTagRequest) (*dto.TagResponse, error) {
	nameVI := utils.NormalizeOptionalString(req.NameVI)
	nameEN := utils.NormalizeOptionalString(req.NameEN)
	if nameVI == nil && nameEN == nil {
		return nil, errors.New("at least one name is required")
	}
 
	color := utils.NormalizeOptionalString(req.Color)
 
	t := entity.Tag{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID: userID,
		NameVI: nameVI,
		NameEN: nameEN,
		Color:  color,
	}
 
	if err := s.repo.CreateTagTx(ctx, nil, userID, t); err != nil {
		return nil, err
	}
 
	it, err := s.repo.GetTag(ctx, userID, t.ID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
 
	resp := dto.NewTagResponse(*it)
	return &resp, nil
}
 
func (s *TagService) Get(ctx context.Context, userID, tagID uuid.UUID) (*dto.TagResponse, error) {
	it, err := s.repo.GetTag(ctx, userID, tagID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
 
	resp := dto.NewTagResponse(*it)
	return &resp, nil
}
 
func (s *TagService) List(ctx context.Context, userID uuid.UUID) ([]dto.TagResponse, error) {
	items, err := s.repo.ListTags(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewTagResponses(items), nil
}
 
func (s *TagService) GetOrCreateByName(ctx context.Context, userID uuid.UUID, name, langHint string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return uuid.Nil, errors.New("tag name cannot be empty")
	}
 
	tags, err := s.repo.ListTags(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
 
	normalizedSearch := strings.ToLower(name)
	for _, t := range tags {
		if t.NameVI != nil && strings.ToLower(*t.NameVI) == normalizedSearch {
			return t.ID, nil
		}
		if t.NameEN != nil && strings.ToLower(*t.NameEN) == normalizedSearch {
			return t.ID, nil
		}
	}
 
	// Create new tag
	id := utils.NewID()
	t := entity.Tag{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: id,
			},
		},
		UserID: userID,
	}
 
	if langHint == "vi" {
		t.NameVI = &name
	} else {
		t.NameEN = &name
	}
 
	if err := s.repo.CreateTagTx(ctx, nil, userID, t); err != nil {
		return uuid.Nil, err
	}
 
	return id, nil
}
