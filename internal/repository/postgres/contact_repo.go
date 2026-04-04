package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type ContactRepo struct {
	db *database.Postgres
}

func NewContactRepo(db *database.Postgres) *ContactRepo {
	return &ContactRepo{db: db}
}

func (r *ContactRepo) CreateContact(ctx context.Context, c entity.Contact) error {
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

func (r *ContactRepo) GetContact(ctx context.Context, userID, contactID string) (*entity.Contact, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT c.id, c.user_id, c.name, c.email, c.phone, c.avatar_url, c.linked_user_id, c.notes, c.created_at, c.updated_at, c.deleted_at,
		       u.display_name AS linked_display_name, u.avatar_url AS linked_avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
	`, contactID, userID)

	var c entity.Contact
	err = row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.AvatarURL, &c.LinkedUserID, &c.Notes, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&c.LinkedDisplayName, &c.LinkedAvatarURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *ContactRepo) ListContacts(ctx context.Context, userID string) ([]entity.Contact, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT c.id, c.user_id, c.name, c.email, c.phone, c.avatar_url, c.linked_user_id, c.notes, c.created_at, c.updated_at, c.deleted_at,
		       u.display_name AS linked_display_name, u.avatar_url AS linked_avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Contact
	for rows.Next() {
		var c entity.Contact
		err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.AvatarURL, &c.LinkedUserID, &c.Notes, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
			&c.LinkedDisplayName, &c.LinkedAvatarURL,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}

func (r *ContactRepo) UpdateContact(ctx context.Context, userID string, c entity.Contact) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE contacts
		SET name = $1, email = $2, phone = $3, avatar_url = $4, linked_user_id = $5, notes = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9
	`, c.Name, c.Email, c.Phone, c.AvatarURL, c.LinkedUserID, c.Notes, c.UpdatedAt, c.ID, userID)
	return err
}

func (r *ContactRepo) DeleteContact(ctx context.Context, userID, contactID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE contacts SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3
	`, time.Now().UTC(), contactID, userID)
	return err
}

func (r *ContactRepo) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, "SELECT id, email, display_name, avatar_url FROM users WHERE email = $1", email)
	var u entity.User
	err = row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}
	return &u, nil
}

func (r *ContactRepo) FindUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, "SELECT id, email, display_name, avatar_url FROM users WHERE phone = $1", phone)
	var u entity.User
	err = row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}
	return &u, nil
}
