package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type SavingsRepository interface {
	CreateSavingsInstrument(ctx context.Context, userID string, s entity.SavingsInstrument) error
	GetSavingsInstrument(ctx context.Context, userID, id string) (*entity.SavingsInstrument, error)
	ListSavingsInstruments(ctx context.Context, userID string) ([]entity.SavingsInstrument, error)
	UpdateSavingsInstrument(ctx context.Context, userID string, s entity.SavingsInstrument) error
	DeleteSavingsInstrument(ctx context.Context, userID, id string) error
}

type RotatingSavingsRepository interface {
	CreateGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	GetGroup(ctx context.Context, userID, groupID string) (*entity.RotatingSavingsGroup, error)
	UpdateGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	DeleteGroup(ctx context.Context, userID, groupID string) error
	ListGroups(ctx context.Context, userID string) ([]entity.RotatingSavingsGroup, error)

	CreateContribution(ctx context.Context, c entity.RotatingSavingsContribution) error
	GetContribution(ctx context.Context, userID, id string) (*entity.RotatingSavingsContribution, error)
	ListContributions(ctx context.Context, userID, groupID string) ([]entity.RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, userID, id string) error

	CreateAuditLog(ctx context.Context, log entity.RotatingSavingsAuditLog) error
	ListAuditLogs(ctx context.Context, userID, groupID string) ([]entity.RotatingSavingsAuditLog, error)
}

type SavingsService interface {
	CreateSavingsInstrument(ctx context.Context, userID string, req dto.CreateSavingsInstrumentRequest) (*entity.SavingsInstrument, error)
	GetSavingsInstrument(ctx context.Context, userID, id string) (*entity.SavingsInstrument, error)
	ListSavingsInstruments(ctx context.Context, userID string) ([]entity.SavingsInstrument, error)
	DeleteSavingsInstrument(ctx context.Context, userID, id string) error
}

type RotatingSavingsService interface {
	CreateGroup(ctx context.Context, userID string, req dto.CreateRotatingSavingsGroupRequest) (*entity.RotatingSavingsGroup, error)
	GetGroup(ctx context.Context, userID, groupID string) (*entity.RotatingSavingsGroup, error)
	GetGroupDetail(ctx context.Context, userID, groupID string) (*dto.RotatingSavingsGroupDetailResponse, error)
	UpdateGroup(ctx context.Context, userID, groupID string, req dto.UpdateRotatingSavingsGroupRequest) (*entity.RotatingSavingsGroup, error)
	DeleteGroup(ctx context.Context, userID, groupID string) error
	ListGroups(ctx context.Context, userID string) ([]dto.RotatingSavingsGroupSummary, error)

	CreateContribution(ctx context.Context, userID, groupID string, req dto.RotatingSavingsContributionRequest) (*entity.RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, userID, groupID, contributionID string) error
}
