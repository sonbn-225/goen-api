package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type RotatingSavingsRepository interface {
	CreateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	GetRotatingGroup(ctx context.Context, userID, groupID string) (*entity.RotatingSavingsGroup, error)
	ListRotatingGroups(ctx context.Context, userID string) ([]entity.RotatingSavingsGroup, error)
	UpdateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	DeleteRotatingGroup(ctx context.Context, userID, groupID string) error

	CreateContribution(ctx context.Context, c entity.RotatingSavingsContribution) error
	GetContributions(ctx context.Context, groupID string) ([]entity.RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, contributionID string) error

	AddAuditLog(ctx context.Context, log entity.RotatingSavingsAuditLog) error
	GetAuditLogs(ctx context.Context, groupID string) ([]entity.RotatingSavingsAuditLog, error)
}

type RotatingSavingsService interface {
	CreateGroup(ctx context.Context, userID string, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	GetGroup(ctx context.Context, userID, groupID string) (*dto.RotatingSavingsGroupResponse, error)
	GetGroupDetail(ctx context.Context, userID, groupID string) (*dto.RotatingSavingsGroupDetailResponse, error)
	UpdateGroup(ctx context.Context, userID, groupID string, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	DeleteGroup(ctx context.Context, userID, groupID string) error
	ListGroups(ctx context.Context, userID string) ([]dto.RotatingSavingsGroupSummary, error)

	CreateContribution(ctx context.Context, userID, groupID string, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error)
	DeleteContribution(ctx context.Context, userID, groupID, contributionID string) error
}
