package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type RotatingSavingsRepository interface {
	CreateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	GetRotatingGroup(ctx context.Context, userID, groupID uuid.UUID) (*entity.RotatingSavingsGroup, error)
	ListRotatingGroups(ctx context.Context, userID uuid.UUID) ([]entity.RotatingSavingsGroup, error)
	UpdateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error
	DeleteRotatingGroup(ctx context.Context, userID, groupID uuid.UUID) error

	CreateContribution(ctx context.Context, c entity.RotatingSavingsContribution) error
	CreateContributionTx(ctx context.Context, tx pgx.Tx, c entity.RotatingSavingsContribution) error
	GetContributions(ctx context.Context, groupID uuid.UUID) ([]entity.RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, contributionID uuid.UUID) error

	AddAuditLog(ctx context.Context, log entity.RotatingSavingsAuditLog) error
	AddAuditLogTx(ctx context.Context, tx pgx.Tx, log entity.RotatingSavingsAuditLog) error
	GetAuditLogs(ctx context.Context, groupID uuid.UUID) ([]entity.RotatingSavingsAuditLog, error)
}

type RotatingSavingsService interface {
	CreateGroup(ctx context.Context, userID uuid.UUID, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupResponse, error)
	GetGroupDetail(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupDetailResponse, error)
	UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error
	ListGroups(ctx context.Context, userID uuid.UUID) ([]dto.RotatingSavingsGroupSummary, error)

	CreateContribution(ctx context.Context, userID, groupID uuid.UUID, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error)
	DeleteContribution(ctx context.Context, userID, groupID, contributionID uuid.UUID) error
}

