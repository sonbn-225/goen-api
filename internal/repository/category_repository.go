package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/category"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetByID(ctx context.Context, userID, categoryID string) (*category.Category, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "category", "operation", "get_by_id", "user_id", userID, "category_id", categoryID)
	row := r.db.QueryRow(ctx, `
		SELECT id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = $1 AND deleted_at IS NULL
	`, categoryID)

	item, err := scanCategory(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_category_get_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_category_get_succeeded")
	return item, nil
}

func (r *CategoryRepository) ListByUser(ctx context.Context, userID string) ([]category.Category, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "category", "operation", "list_by_user", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL
		  AND is_active = true
		ORDER BY COALESCE(sort_order, 0) ASC, id ASC
	`)
	if err != nil {
		logger.Error("repo_category_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]category.Category, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			logger.Error("repo_category_list_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_category_list_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_category_list_succeeded", "count", len(items))
	return items, nil
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner) (*category.Category, error) {
	var item category.Category
	var parentID sql.NullString
	var categoryType sql.NullString
	var sortOrder sql.NullInt32
	var icon sql.NullString
	var color sql.NullString
	var deletedAt sql.NullTime

	err := scanner.Scan(
		&item.ID,
		&parentID,
		&categoryType,
		&sortOrder,
		&item.IsActive,
		&icon,
		&color,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		item.ParentCategoryID = &parentID.String
	}
	if categoryType.Valid {
		item.Type = &categoryType.String
	}
	if sortOrder.Valid {
		v := int(sortOrder.Int32)
		item.SortOrder = &v
	}
	if icon.Valid {
		item.Icon = &icon.String
	}
	if color.Valid {
		item.Color = &color.String
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}

	return &item, nil
}
