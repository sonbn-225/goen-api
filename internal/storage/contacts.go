package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type ContactRepo struct {
	db *Postgres
}

func NewContactRepo(db *Postgres) *ContactRepo {
	return &ContactRepo{db: db}
}

func (r *ContactRepo) CreateContact(ctx context.Context, c domain.Contact) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO contacts (id, user_id, name, email, phone, avatar_url, linked_user_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, c.ID, c.UserID, c.Name, c.Email, c.Phone, c.AvatarURL, c.LinkedUserID, c.Notes, c.CreatedAt, c.UpdatedAt)

	return err
}

func (r *ContactRepo) GetContact(ctx context.Context, userID, contactID string) (*domain.Contact, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT c.id, c.user_id, c.name, c.email, c.phone, c.avatar_url, c.linked_user_id, c.notes, c.created_at, c.updated_at,
		       u.display_name as linked_display_name, u.avatar_url as linked_avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1 AND c.id = $2 AND c.deleted_at IS NULL
	`, userID, contactID)

	var c domain.Contact
	err = row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.AvatarURL, &c.LinkedUserID, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
		&c.LinkedDisplayName, &c.LinkedAvatarURL,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrContactNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *ContactRepo) ListContacts(ctx context.Context, userID string) ([]domain.Contact, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT c.id, c.user_id, c.name, c.email, c.phone, c.avatar_url, c.linked_user_id, c.notes, c.created_at, c.updated_at,
		       u.display_name as linked_display_name, u.avatar_url as linked_avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name ASC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Contact, 0)
	for rows.Next() {
		var c domain.Contact
		err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.AvatarURL, &c.LinkedUserID, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
			&c.LinkedDisplayName, &c.LinkedAvatarURL,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}

	return items, nil
}

func (r *ContactRepo) UpdateContact(ctx context.Context, userID string, c domain.Contact) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE contacts 
		SET name = $1, email = $2, phone = $3, avatar_url = $4, linked_user_id = $5, notes = $6, updated_at = $7
		WHERE user_id = $8 AND id = $9 AND deleted_at IS NULL
	`, c.Name, c.Email, c.Phone, c.AvatarURL, c.LinkedUserID, c.Notes, c.UpdatedAt, userID, c.ID)

	return err
}

func (r *ContactRepo) DeleteContact(ctx context.Context, userID, contactID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE contacts SET deleted_at = NOW() WHERE user_id = $1 AND id = $2
	`, userID, contactID)

	return err
}

func (r *ContactRepo) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)

	var u domain.User
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *ContactRepo) FindUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, email, phone, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE phone = $1
	`, phone)

	var u domain.User
	err = row.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.AvatarURL, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}
