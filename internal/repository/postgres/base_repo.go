package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

// BaseRepo provides common functionality for PostgreSQL repositories.
type BaseRepo struct {
	db *database.Postgres
}

// NewBaseRepo creates a new BaseRepo instance.
func NewBaseRepo(db *database.Postgres) *BaseRepo {
	return &BaseRepo{db: db}
}

// SoftDelete performs a soft delete on the specified table for the given ID.
// It sets the deleted_at timestamp and optionally the updated_by field.
func (r *BaseRepo) SoftDelete(ctx context.Context, table string, id uuid.UUID, userID *uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	now := utils.Now()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = $1", table)
	args := []any{now}

	if userID != nil {
		query += ", updated_by = $2 WHERE id = $3 AND deleted_at IS NULL"
		args = append(args, *userID, id)
	} else {
		query += " WHERE id = $2 AND deleted_at IS NULL"
		args = append(args, id)
	}

	commandTag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("%s not found or already deleted", table)
	}

	return nil
}

// SoftDeleteByField performs a soft delete based on a specific field match.
func (r *BaseRepo) SoftDeleteByField(ctx context.Context, table string, field string, value any, userID *uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	now := utils.Now()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = $1", table)
	args := []any{now}

	argCount := 2
	if userID != nil {
		query += fmt.Sprintf(", updated_by = $%d", argCount)
		args = append(args, *userID)
		argCount++
	}

	query += fmt.Sprintf(" WHERE %s = $%d AND deleted_at IS NULL", field, argCount)
	args = append(args, value)

	_, err = pool.Exec(ctx, query, args...)
	return err
}

// UpdateTimestamps returns the current time for updated_at and optionally for created_at.
func (r *BaseRepo) UpdateTimestamps() time.Time {
	return utils.Now()
}
