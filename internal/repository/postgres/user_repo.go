package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type UserRepo struct {
	db *database.Postgres
}

func NewUserRepo(db *database.Postgres) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUserWithRefreshToken(ctx context.Context, u entity.UserWithPassword, refreshToken entity.RefreshToken) error {
	settingsJSON, err := json.Marshal(u.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	defaultCurrency := "VND"
	if settings, ok := u.Settings.(map[string]any); ok {
		if raw, ok := settings["default_currency"]; ok {
			if cur, ok := raw.(string); ok {
				normalized := strings.ToUpper(strings.TrimSpace(cur))
				if len(normalized) == 3 {
					defaultCurrency = normalized
				}
			}
		}
	}

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Insert user
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, email, phone, display_name, avatar_url, username, settings, status, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, u.ID, u.Email, u.Phone, u.DisplayName, u.AvatarURL, u.Username, settingsJSON, u.Status, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return errors.New("user already exists")
			}
			return fmt.Errorf("failed to insert user: %w", err)
		}

		// 2. Create initial cash account
		cashAccountID := uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO accounts (id, name, account_type, currency, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, cashAccountID, "cash_account_name", "cash", defaultCurrency, "active", u.CreatedAt, u.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create cash account: %w", err)
		}

		// 3. Link user to account
		_, err = tx.Exec(ctx, `
			INSERT INTO user_accounts (id, account_id, user_id, permission, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.New(), cashAccountID, u.ID, "owner", "active", u.CreatedAt, u.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to link user to account: %w", err)
		}

		// 4. Create bootstrap refresh token in the same transaction.
		_, err = tx.Exec(ctx, `
			INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, refreshToken.ID, u.ID, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.CreatedAt, refreshToken.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert refresh token: %w", err)
		}

		return nil
	})
}

func (r *UserRepo) FindUserByEmail(ctx context.Context, email string) (*entity.UserWithPassword, error) {
	return r.findOneUser(ctx, "email = $1", email)
}

func (r *UserRepo) FindUserByPhone(ctx context.Context, phone string) (*entity.UserWithPassword, error) {
	return r.findOneUser(ctx, "phone = $1", phone)
}

func (r *UserRepo) FindUserByUsername(ctx context.Context, username string) (*entity.UserWithPassword, error) {
	return r.findOneUser(ctx, "username = $1", username)
}

func (r *UserRepo) FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, avatar_url, username, settings, status, created_at, updated_at
		FROM users WHERE id = $1
	`, id)

	var u entity.User
	var settingsJSON []byte
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Username, &settingsJSON, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &u.Settings)
	}
	return &u, nil
}

func (r *UserRepo) UpdateUserSettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*entity.User, error) {
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
		RETURNING id, email, phone, display_name, avatar_url, username, settings, status, created_at, updated_at
	`, patchJSON, userID)

	var u entity.User
	var settingsJSON []byte
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Username, &settingsJSON, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &u.Settings)
	}
	return &u, nil
}

func (r *UserRepo) UpdateUserProfile(ctx context.Context, userID uuid.UUID, params entity.UpdateUserParams) (*entity.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		UPDATE users
		SET display_name  = COALESCE($1, display_name),
		    avatar_url    = COALESCE($2, avatar_url),
		    email         = COALESCE($3, email),
		    phone         = COALESCE($4, phone),
		    username      = COALESCE($5, username),
		    password_hash = COALESCE($6, password_hash),
		    updated_at    = NOW()
		WHERE id = $7
		RETURNING id, email, phone, display_name, avatar_url, username, settings, status, created_at, updated_at
	`, params.DisplayName, params.AvatarURL, params.Email, params.Phone, params.Username, params.PasswordHash, userID)

	var u entity.User
	var settingsJSON []byte
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Username, &settingsJSON, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("user already exists")
		}
		return nil, err
	}

	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &u.Settings)
	}
	return &u, nil
}

func (r *UserRepo) findOneUser(ctx context.Context, where string, args ...any) (*entity.UserWithPassword, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, email, phone, display_name, avatar_url, username, settings, status, password_hash, created_at, updated_at
		FROM users WHERE %s
	`, where)

	row := pool.QueryRow(ctx, query, args...)

	var u entity.UserWithPassword
	var settingsJSON []byte
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Username, &settingsJSON, &u.Status, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &u.Settings)
	}
	return &u, nil
}
