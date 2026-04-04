package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type CategoryRepo struct {
	db *database.Postgres
}

func NewCategoryRepo(db *database.Postgres) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) GetCategory(ctx context.Context, userID, categoryID string) (*entity.Category, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = $1 AND deleted_at IS NULL
	`, categoryID)

	var c entity.Category
	var parentIDNull sql.NullString
	var typeNull sql.NullString
	var iconNull sql.NullString
	var colorNull sql.NullString
	var deletedAtNull sql.NullTime
	if err := row.Scan(
		&c.ID,
		&parentIDNull,
		&typeNull,
		&c.SortOrder,
		&c.IsActive,
		&iconNull,
		&colorNull,
		&c.CreatedAt,
		&c.UpdatedAt,
		&deletedAtNull,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	if parentIDNull.Valid {
		c.ParentCategoryID = &parentIDNull.String
	}
	if typeNull.Valid {
		c.Type = &typeNull.String
	}
	if iconNull.Valid {
		c.Icon = &iconNull.String
	}
	if colorNull.Valid {
		c.Color = &colorNull.String
	}
	if deletedAtNull.Valid {
		c.DeletedAt = &deletedAtNull.Time
	}

	return &c, nil
}

func (r *CategoryRepo) ListCategories(ctx context.Context, userID string) ([]entity.Category, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, parent_category_id, type, sort_order, is_active, icon, color, created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY COALESCE(sort_order, 0) ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Category, 0)
	for rows.Next() {
		var c entity.Category
		var parentIDNull sql.NullString
		var typeNull sql.NullString
		var iconNull sql.NullString
		var colorNull sql.NullString
		var deletedAtNull sql.NullTime
		if err := rows.Scan(
			&c.ID,
			&parentIDNull,
			&typeNull,
			&c.SortOrder,
			&c.IsActive,
			&iconNull,
			&colorNull,
			&c.CreatedAt,
			&c.UpdatedAt,
			&deletedAtNull,
		); err != nil {
			return nil, err
		}
		if parentIDNull.Valid {
			c.ParentCategoryID = &parentIDNull.String
		}
		if typeNull.Valid {
			c.Type = &typeNull.String
		}
		if iconNull.Valid {
			c.Icon = &iconNull.String
		}
		if colorNull.Valid {
			c.Color = &colorNull.String
		}
		if deletedAtNull.Valid {
			c.DeletedAt = &deletedAtNull.Time
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
