package postgres

import (
	"context"
	"fmt"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type RefreshTokenRepo struct {
	db *database.Postgres
}

func NewRefreshTokenRepo(db *database.Postgres) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, t *entity.RefreshToken) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, t.ID, t.UserID, t.Token, t.ExpiresAt, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens WHERE token = $1
	`, token)

	var t entity.RefreshToken
	err = row.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &t, nil
}

func (r *RefreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE token = $1", token)
	return err
}

func (r *RefreshTokenRepo) DeleteAllByUserID(ctx context.Context, userID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}
