package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, phone, display_name, status, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, u.ID, u.Email, u.Phone, u.DisplayName, u.Status, u.PasswordHash, u.CreatedAt, u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrUserAlreadyExists
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
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, status, created_at, updated_at
		FROM users
		WHERE id = $1`, id)

	var u domain.User
	err = row.Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.DisplayName,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) findOneUser(ctx context.Context, whereClause string, args ...any) (*domain.UserWithPassword, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, status, password_hash, created_at, updated_at
		FROM users
		WHERE `+whereClause, args...)

	var u domain.UserWithPassword
	err = row.Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.DisplayName,
		&u.Status,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// EnsureUsersSchema creates the users table if it doesn't exist.
// This is used at startup for simplicity in this project.
// DEPRECATED: Use goose migrations instead.
func EnsureUsersSchema(ctx context.Context, db *Postgres) error {
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
