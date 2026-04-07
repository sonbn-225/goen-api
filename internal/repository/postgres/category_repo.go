package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type CategoryRepo struct {
	BaseRepo
}

func NewCategoryRepo(db *database.Postgres) *CategoryRepo {
	return &CategoryRepo{BaseRepo: *NewBaseRepo(db)}
}

func (r *CategoryRepo) GetCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) (*entity.Category, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, key, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = $1 AND deleted_at IS NULL
	`, categoryID)

	var c entity.Category
	if err := row.Scan(
		&c.ID,
		&c.Key,
		&c.ParentCategoryID,
		&c.Type,
		&c.SortOrder,
		&c.IsActive,
		&c.Icon,
		&c.Color,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &c, nil
}

func (r *CategoryRepo) ListCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, key, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY COALESCE(sort_order, 0) ASC, key ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Category, 0)
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(
			&c.ID,
			&c.Key,
			&c.ParentCategoryID,
			&c.Type,
			&c.SortOrder,
			&c.IsActive,
			&c.Icon,
			&c.Color,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
