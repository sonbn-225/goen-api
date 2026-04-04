package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/tag"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, userID string, input tag.Tag) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "tag", "operation", "create", "user_id", userID, "tag_id", input.ID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO tags (id, user_id, name_vi, name_en, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, input.ID, input.UserID, input.NameVI, input.NameEN, input.Color, input.CreatedAt, input.UpdatedAt)
	if err != nil {
		logger.Error("repo_tag_create_failed", "error", err)
		return err
	}
	logger.Info("repo_tag_create_succeeded")
	return nil
}

func (r *TagRepository) GetByID(ctx context.Context, userID, tagID string) (*tag.Tag, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "tag", "operation", "get_by_id", "user_id", userID, "tag_id", tagID)
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, name_vi, name_en, color, created_at, updated_at
		FROM tags
		WHERE id = $1 AND user_id = $2
	`, tagID, userID)

	item, err := scanTag(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_tag_get_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_tag_get_succeeded")
	return item, nil
}

func (r *TagRepository) ListByUser(ctx context.Context, userID string) ([]tag.Tag, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "tag", "operation", "list_by_user", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name_vi, name_en, color, created_at, updated_at
		FROM tags
		WHERE user_id = $1
		ORDER BY COALESCE(name_vi, name_en) ASC
	`, userID)
	if err != nil {
		logger.Error("repo_tag_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]tag.Tag, 0)
	for rows.Next() {
		item, err := scanTag(rows)
		if err != nil {
			logger.Error("repo_tag_list_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_tag_list_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_tag_list_succeeded", "count", len(items))
	return items, nil
}

type tagScanner interface {
	Scan(dest ...any) error
}

func scanTag(scanner tagScanner) (*tag.Tag, error) {
	var item tag.Tag
	var nameVI sql.NullString
	var nameEN sql.NullString
	var color sql.NullString

	err := scanner.Scan(&item.ID, &item.UserID, &nameVI, &nameEN, &color, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if nameVI.Valid {
		item.NameVI = &nameVI.String
	}
	if nameEN.Valid {
		item.NameEN = &nameEN.String
	}
	if color.Valid {
		item.Color = &color.String
	}

	return &item, nil
}
