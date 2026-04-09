package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type TagRepo struct {
	BaseRepo
}

func NewTagRepo(db *database.Postgres) *TagRepo {
	return &TagRepo{BaseRepo: *NewBaseRepo(db)}
}

// --- Nhóm 1: Truy vấn Nhãn (Flexible Tx) ---

func (r *TagRepo) GetTagTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tagID uuid.UUID) (*entity.Tag, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, `
		SELECT id, user_id, name_vi, name_en, color, created_at, updated_at
		FROM tags
		WHERE id = $1 AND user_id = $2
	`, tagID, userID)

	var t entity.Tag
	var nameVINull, nameENNull, colorNull sql.NullString
	if err := row.Scan(&t.ID, &t.UserID, &nameVINull, &nameENNull, &colorNull, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}
	if nameVINull.Valid {
		t.NameVI = &nameVINull.String
	}
	if nameENNull.Valid {
		t.NameEN = &nameENNull.String
	}
	if colorNull.Valid {
		t.Color = &colorNull.String
	}
	return &t, nil
}

func (r *TagRepo) ListTagsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Tag, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, user_id, name_vi, name_en, color, created_at, updated_at
		FROM tags
		WHERE user_id = $1
		ORDER BY COALESCE(name_vi, name_en) ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Tag, 0)
	for rows.Next() {
		var t entity.Tag
		var nameVINull, nameENNull, colorNull sql.NullString
		if err := rows.Scan(&t.ID, &t.UserID, &nameVINull, &nameENNull, &colorNull, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if nameVINull.Valid {
			t.NameVI = &nameVINull.String
		}
		if nameENNull.Valid {
			t.NameEN = &nameENNull.String
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

// --- Nhóm 2: Thao tác ghi & Nhất quán (Transactional) ---

func (r *TagRepo) CreateTagTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Tag) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO tags (id, user_id, name_vi, name_en, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, t.ID, userID, t.NameVI, t.NameEN, t.Color, t.CreatedAt, t.UpdatedAt)
	return err
}
