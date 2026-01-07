package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type CategoryRepo struct {
	db *Postgres
}

func NewCategoryRepo(db *Postgres) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) CreateCategory(ctx context.Context, userID string, c domain.Category) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	if strings.TrimSpace(userID) == "" {
		return errors.New("userID is required")
	}
	if c.UserID == nil || strings.TrimSpace(*c.UserID) == "" {
		uid := userID
		c.UserID = &uid
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO categories (
			id, user_id, name, parent_category_id, type, sort_order, is_active, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		c.ID,
		c.UserID,
		c.Name,
		c.ParentCategoryID,
		c.Type,
		c.SortOrder,
		c.IsActive,
		c.CreatedAt,
		c.UpdatedAt,
		c.DeletedAt,
	)
	return err
}

func (r *CategoryRepo) GetCategory(ctx context.Context, userID string, categoryID string) (*domain.Category, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, user_id, name, parent_category_id, type, sort_order, is_active, created_at, updated_at, deleted_at
		FROM categories
		WHERE id = $1 AND (user_id = $2 OR user_id IS NULL) AND deleted_at IS NULL
	`, categoryID, userID)

	var c domain.Category
	var userIDNull sql.NullString
	var parentIDNull sql.NullString
	var typeNull sql.NullString
	var deletedAtNull sql.NullTime
	if err := row.Scan(
		&c.ID,
		&userIDNull,
		&c.Name,
		&parentIDNull,
		&typeNull,
		&c.SortOrder,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&deletedAtNull,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	if userIDNull.Valid {
		c.UserID = &userIDNull.String
	}
	if parentIDNull.Valid {
		c.ParentCategoryID = &parentIDNull.String
	}
	if typeNull.Valid {
		c.Type = &typeNull.String
	}
	if deletedAtNull.Valid {
		c.DeletedAt = &deletedAtNull.Time
	}

	return &c, nil
}

func (r *CategoryRepo) ListCategories(ctx context.Context, userID string) ([]domain.Category, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, user_id, name, parent_category_id, type, sort_order, is_active, created_at, updated_at, deleted_at
		FROM categories
		WHERE (user_id = $1 OR user_id IS NULL) AND deleted_at IS NULL
		ORDER BY COALESCE(sort_order, 0) ASC, name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Category, 0)
	for rows.Next() {
		var c domain.Category
		var userIDNull sql.NullString
		var parentIDNull sql.NullString
		var typeNull sql.NullString
		var deletedAtNull sql.NullTime
		if err := rows.Scan(
			&c.ID,
			&userIDNull,
			&c.Name,
			&parentIDNull,
			&typeNull,
			&c.SortOrder,
			&c.IsActive,
			&c.CreatedAt,
			&c.UpdatedAt,
			&deletedAtNull,
		); err != nil {
			return nil, err
		}
		if userIDNull.Valid {
			c.UserID = &userIDNull.String
		}
		if parentIDNull.Valid {
			c.ParentCategoryID = &parentIDNull.String
		}
		if typeNull.Valid {
			c.Type = &typeNull.String
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
