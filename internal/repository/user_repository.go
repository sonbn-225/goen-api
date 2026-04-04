package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user auth.UserWithPassword) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "create_user", "user_id", user.ID)
	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		logger.Error("repo_create_user_failed", "error", err)
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO users (
			id,
			username,
			email,
			phone,
			display_name,
			status,
			password_hash,
			settings,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'active', $6, $7::jsonb, $8, $9
		)
	`,
		user.ID,
		strings.ToLower(user.Username),
		nullableString(user.Email),
		nullableString(user.Phone),
		nullableString(user.DisplayName),
		user.PasswordHash,
		string(settingsJSON),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_create_user_failed", "error", err)
		return err
	}
	logger.Info("repo_create_user_succeeded")
	return err
}

func (r *UserRepository) FindUserByID(ctx context.Context, userID string) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "find_user_by_id", "user_id", userID)
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, display_name, avatar_url, settings, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	result, err := scanUser(row)
	if err != nil {
		logger.Error("repo_find_user_by_id_failed", "error", err)
		return nil, err
	}
	return result, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*auth.UserWithPassword, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "find_user_by_email")
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, display_name, avatar_url, settings, created_at, updated_at, password_hash
		FROM users
		WHERE lower(email) = lower($1)
	`, email)
	result, err := scanUserWithPassword(row)
	if err != nil {
		logger.Error("repo_find_user_by_email_failed", "error", err)
		return nil, err
	}
	return result, nil
}

func (r *UserRepository) FindUserByPhone(ctx context.Context, phone string) (*auth.UserWithPassword, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "find_user_by_phone")
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, display_name, avatar_url, settings, created_at, updated_at, password_hash
		FROM users
		WHERE phone = $1
	`, phone)
	result, err := scanUserWithPassword(row)
	if err != nil {
		logger.Error("repo_find_user_by_phone_failed", "error", err)
		return nil, err
	}
	return result, nil
}

func (r *UserRepository) FindUserByUsername(ctx context.Context, username string) (*auth.UserWithPassword, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "find_user_by_username")
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, display_name, avatar_url, settings, created_at, updated_at, password_hash
		FROM users
		WHERE lower(username) = lower($1)
	`, username)
	result, err := scanUserWithPassword(row)
	if err != nil {
		logger.Error("repo_find_user_by_username_failed", "error", err)
		return nil, err
	}
	return result, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID string, input auth.UpdateProfileInput) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "update_user_profile", "user_id", userID)
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET
			display_name = COALESCE($2, display_name),
			email = COALESCE($3, email),
			phone = COALESCE($4, phone),
			username = COALESCE($5, username),
			updated_at = $6
		WHERE id = $1
	`,
		userID,
		nullableString(input.DisplayName),
		nullableString(input.Email),
		nullableString(input.Phone),
		nullableString(input.Username),
		time.Now().UTC(),
	)
	if err != nil {
		logger.Error("repo_update_user_profile_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_update_user_profile_succeeded")
	return r.FindUserByID(ctx, userID)
}

func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID, avatarURL string) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "update_avatar_url", "user_id", userID)
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE users
		SET avatar_url = $2,
			updated_at = $3
		WHERE id = $1
	`, userID, avatarURL, time.Now().UTC())
	if err != nil {
		logger.Error("repo_update_avatar_url_failed", "error", err)
		return nil, err
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, nil
	}
	logger.Info("repo_update_avatar_url_succeeded", "avatar_url", avatarURL)
	return r.FindUserByID(ctx, userID)
}

func (r *UserRepository) UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*auth.User, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "update_user_settings", "user_id", userID)
	logger.Info("repo_update_user_settings_payload", logx.MaskAttrs("patch", patch)...)
	patchJSON, err := json.Marshal(patch)
	if err != nil {
		logger.Error("repo_update_user_settings_failed", "error", err)
		return nil, err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE users
		SET settings = COALESCE(settings, '{}'::jsonb) || $2::jsonb,
			updated_at = $3
		WHERE id = $1
	`, userID, string(patchJSON), time.Now().UTC())
	if err != nil {
		logger.Error("repo_update_user_settings_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_update_user_settings_succeeded")
	return r.FindUserByID(ctx, userID)
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "auth", "operation", "update_password_hash", "user_id", userID)
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET password_hash = $2,
			updated_at = $3
		WHERE id = $1
	`, userID, passwordHash, time.Now().UTC())
	if err != nil {
		logger.Error("repo_update_password_hash_failed", "error", err)
		return err
	}
	logger.Info("repo_update_password_hash_succeeded", logx.MaskAttrs("password_hash", passwordHash)...)
	return err
}

func nullableString(v *string) any {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return s
}

type userRow interface {
	Scan(dest ...any) error
}

func scanUser(row userRow) (*auth.User, error) {
	var (
		user         auth.User
		email        *string
		phone        *string
		displayName  *string
		avatarURL    *string
		settingsJSON []byte
	)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&email,
		&phone,
		&displayName,
		&avatarURL,
		&settingsJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	user.Email = email
	user.Phone = phone
	user.DisplayName = displayName
	user.AvatarURL = avatarURL
	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &user.Settings)
	}
	return &user, nil
}

func scanUserWithPassword(row userRow) (*auth.UserWithPassword, error) {
	var (
		user         auth.UserWithPassword
		email        *string
		phone        *string
		displayName  *string
		avatarURL    *string
		settingsJSON []byte
	)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&email,
		&phone,
		&displayName,
		&avatarURL,
		&settingsJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.PasswordHash,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	user.Email = email
	user.Phone = phone
	user.DisplayName = displayName
	user.AvatarURL = avatarURL
	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &user.Settings)
	}
	return &user, nil
}
