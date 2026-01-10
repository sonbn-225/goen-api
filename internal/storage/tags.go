package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

type TagRepo struct {
	db *Postgres
}

func NewTagRepo(db *Postgres) *TagRepo {
	return &TagRepo{db: db}
}

func (r *TagRepo) CreateTag(ctx context.Context, userID string, t domain.Tag) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	if strings.TrimSpace(userID) == "" {
		return apperrors.ErrUserIDRequired
	}
	if strings.TrimSpace(t.UserID) == "" {
		t.UserID = userID
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO tags (id, user_id, name, color, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, t.ID, t.UserID, t.Name, t.Color, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TagRepo) GetTag(ctx context.Context, userID string, tagID string) (*domain.Tag, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM tags
		WHERE id = $1 AND user_id = $2
	`, tagID, userID)

	var t domain.Tag
	var colorNull sql.NullString
	if err := row.Scan(&t.ID, &t.UserID, &t.Name, &colorNull, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTagNotFound
		}
		return nil, err
	}
	if colorNull.Valid {
		t.Color = &colorNull.String
	}
	return &t, nil
}

func (r *TagRepo) ListTags(ctx context.Context, userID string) ([]domain.Tag, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM tags
		WHERE user_id = $1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Tag, 0)
	for rows.Next() {
		var t domain.Tag
		var colorNull sql.NullString
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &colorNull, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if colorNull.Valid {
			t.Color = &colorNull.String
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
