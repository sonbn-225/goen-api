package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type RefreshTokenRepo struct {
	db *database.Postgres
}

func NewRefreshTokenRepo(db *database.Postgres) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// --- Nhóm 1: Truy vấn Token (Read-only) ---

func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, created_at, updated_at
		FROM refresh_tokens WHERE token = $1
	`, token)

	var t entity.RefreshToken
	err = row.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &t, nil
}

// --- Nhóm 2: Thao tác Token (Transactional) ---

func (r *RefreshTokenRepo) CreateTx(ctx context.Context, tx pgx.Tx, t *entity.RefreshToken) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, t.ID, t.UserID, t.Token, t.ExpiresAt, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) DeleteByTokenTx(ctx context.Context, tx pgx.Tx, token string) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, "DELETE FROM refresh_tokens WHERE token = $1", token)
	return err
}

func (r *RefreshTokenRepo) DeleteAllByUserIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}
