package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

// Ensure UserAdapter implements domain.UserRepository
// Note: We'll create a UserRepo struct to implement this interface explicitly
// and avoid direct coupling if possible, but for now we extend existing logic.

type UserRepo struct {
	db *Postgres
}

func NewUserRepo(db *Postgres) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, u domain.UserWithPassword) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	settingsJSON, err := json.Marshal(u.Settings)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, phone, display_name, settings, status, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9)
	`, u.ID, u.Email, u.Phone, u.DisplayName, settingsJSON, u.Status, u.PasswordHash, u.CreatedAt, u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return apperrors.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindUserByEmail(ctx context.Context, email string) (*domain.UserWithPassword, error) {
	return r.findOneUser(ctx, "email = $1", email)
}

func (r *UserRepo) FindUserByPhone(ctx context.Context, phone string) (*domain.UserWithPassword, error) {
	return r.findOneUser(ctx, "phone = $1", phone)
}

func (r *UserRepo) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, settings, status, created_at, updated_at
		FROM users
		WHERE id = $1`, id)

	var u domain.User
	var settingsJSON []byte
	err = row.Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.DisplayName,
		&settingsJSON,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if len(settingsJSON) > 0 {
		var v any
		if err := json.Unmarshal(settingsJSON, &v); err == nil {
			u.Settings = v
		}
	}
	return &u, nil
}

func (r *UserRepo) findOneUser(ctx context.Context, whereClause string, args ...any) (*domain.UserWithPassword, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, settings, status, password_hash, created_at, updated_at
		FROM users
		WHERE `+whereClause, args...)

	var u domain.UserWithPassword
	var settingsJSON []byte
	err = row.Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.DisplayName,
		&settingsJSON,
		&u.Status,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if len(settingsJSON) > 0 {
		var v any
		if err := json.Unmarshal(settingsJSON, &v); err == nil {
			u.Settings = v
		}
	}
	return &u, nil
}

func (r *UserRepo) UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		UPDATE users
		SET settings = COALESCE(settings, '{}'::jsonb) || $1::jsonb,
		    updated_at = NOW()
		WHERE id = $2
		RETURNING id, email, phone, display_name, settings, status, created_at, updated_at
	`, patchJSON, userID)

	var u domain.User
	var settingsJSON []byte
	if err := row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &settingsJSON, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if len(settingsJSON) > 0 {
		var v any
		if err := json.Unmarshal(settingsJSON, &v); err == nil {
			u.Settings = v
		}
	}
	return &u, nil
}

// EnsureUsersSchema creates the users table if it doesn't exist.
// This is used at startup for simplicity in this project.
// DEPRECATED: Use goose migrations instead.
func EnsureUsersSchema(ctx context.Context, db *Postgres) error {
	return nil
}
