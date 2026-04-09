package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type AuditRepository interface {
	RecordTx(ctx context.Context, tx pgx.Tx, log entity.AuditLog) error
	ListTx(ctx context.Context, tx pgx.Tx, filter entity.AuditLogFilter) ([]entity.AuditLog, error)
	DeleteOldLogs(ctx context.Context, tx pgx.Tx, before time.Time) (int64, error)
}

type AuditService interface {
	// Record simplifies recording an event with automatic diff calculation if oldObj and newObj are provided.
	Record(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, accountID *uuid.UUID, resourceType entity.AuditResourceType, action entity.AuditAction, resourceID uuid.UUID, oldObj any, newObj any) error
	
	// List returns audit logs based on filters.
	List(ctx context.Context, userID uuid.UUID, filter entity.AuditLogFilter) ([]entity.AuditLog, error)
}
