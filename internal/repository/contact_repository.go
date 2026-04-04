package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/contact"
)

type ContactRepository struct {
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, userID string, input contact.Contact) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "contact", "operation", "create", "user_id", userID, "contact_id", input.ID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO contacts (id, user_id, name, email, phone, avatar_url, linked_user_id, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, input.ID, input.UserID, input.Name, input.Email, input.Phone, input.AvatarURL, input.LinkedUserID, input.Notes, input.CreatedAt, input.UpdatedAt)
	if err != nil {
		logger.Error("repo_contact_create_failed", "error", err)
		return err
	}
	logger.Info("repo_contact_create_succeeded")
	return nil
}

func (r *ContactRepository) GetByID(ctx context.Context, userID, contactID string) (*contact.Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "contact", "operation", "get_by_id", "user_id", userID, "contact_id", contactID)
	row := r.db.QueryRow(ctx, `
		SELECT
			c.id,
			c.user_id,
			c.name,
			c.email,
			c.phone,
			c.avatar_url,
			c.linked_user_id,
			c.notes,
			c.created_at,
			c.updated_at,
			c.deleted_at,
			u.display_name,
			u.avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1
		  AND c.id = $2
		  AND c.deleted_at IS NULL
	`, userID, contactID)

	item, err := scanContact(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_contact_get_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_contact_get_succeeded")
	return item, nil
}

func (r *ContactRepository) ListByUser(ctx context.Context, userID string) ([]contact.Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "contact", "operation", "list_by_user", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			c.id,
			c.user_id,
			c.name,
			c.email,
			c.phone,
			c.avatar_url,
			c.linked_user_id,
			c.notes,
			c.created_at,
			c.updated_at,
			c.deleted_at,
			u.display_name,
			u.avatar_url
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1
		  AND c.deleted_at IS NULL
		ORDER BY c.name ASC
	`, userID)
	if err != nil {
		logger.Error("repo_contact_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]contact.Contact, 0)
	for rows.Next() {
		item, err := scanContact(rows)
		if err != nil {
			logger.Error("repo_contact_list_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_contact_list_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_contact_list_succeeded", "count", len(items))
	return items, nil
}

func (r *ContactRepository) Update(ctx context.Context, userID string, input contact.Contact) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "contact", "operation", "update", "user_id", userID, "contact_id", input.ID)
	_, err := r.db.Exec(ctx, `
		UPDATE contacts
		SET name = $1,
			email = $2,
			phone = $3,
			avatar_url = $4,
			linked_user_id = $5,
			notes = $6,
			updated_at = $7
		WHERE user_id = $8
		  AND id = $9
		  AND deleted_at IS NULL
	`, input.Name, input.Email, input.Phone, input.AvatarURL, input.LinkedUserID, input.Notes, input.UpdatedAt, userID, input.ID)
	if err != nil {
		logger.Error("repo_contact_update_failed", "error", err)
		return err
	}
	logger.Info("repo_contact_update_succeeded")
	return nil
}

func (r *ContactRepository) Delete(ctx context.Context, userID, contactID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "contact", "operation", "delete", "user_id", userID, "contact_id", contactID)
	_, err := r.db.Exec(ctx, `
		UPDATE contacts
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
		  AND id = $2
		  AND deleted_at IS NULL
	`, userID, contactID)
	if err != nil {
		logger.Error("repo_contact_delete_failed", "error", err)
		return err
	}
	logger.Info("repo_contact_delete_succeeded")
	return nil
}

func (r *ContactRepository) FindLinkedUserByEmail(ctx context.Context, email string) (*contact.LinkedUser, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, display_name, avatar_url
		FROM users
		WHERE lower(email) = lower($1)
	`, email)

	item, err := scanLinkedUser(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *ContactRepository) FindLinkedUserByPhone(ctx context.Context, phone string) (*contact.LinkedUser, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, display_name, avatar_url
		FROM users
		WHERE phone = $1
	`, phone)

	item, err := scanLinkedUser(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

type contactScanner interface {
	Scan(dest ...any) error
}

func scanContact(scanner contactScanner) (*contact.Contact, error) {
	var item contact.Contact
	var email sql.NullString
	var phone sql.NullString
	var avatarURL sql.NullString
	var linkedUserID sql.NullString
	var notes sql.NullString
	var deletedAt sql.NullTime
	var linkedDisplayName sql.NullString
	var linkedAvatarURL sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&email,
		&phone,
		&avatarURL,
		&linkedUserID,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
		&linkedDisplayName,
		&linkedAvatarURL,
	)
	if err != nil {
		return nil, err
	}

	if email.Valid {
		item.Email = &email.String
	}
	if phone.Valid {
		item.Phone = &phone.String
	}
	if avatarURL.Valid {
		item.AvatarURL = &avatarURL.String
	}
	if linkedUserID.Valid {
		item.LinkedUserID = &linkedUserID.String
	}
	if notes.Valid {
		item.Notes = &notes.String
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	if linkedDisplayName.Valid {
		item.LinkedDisplayName = &linkedDisplayName.String
	}
	if linkedAvatarURL.Valid {
		item.LinkedAvatarURL = &linkedAvatarURL.String
	}

	return &item, nil
}

func scanLinkedUser(scanner contactScanner) (*contact.LinkedUser, error) {
	var item contact.LinkedUser
	var displayName sql.NullString
	var avatarURL sql.NullString

	err := scanner.Scan(&item.ID, &displayName, &avatarURL)
	if err != nil {
		return nil, err
	}

	if displayName.Valid {
		item.DisplayName = &displayName.String
	}
	if avatarURL.Valid {
		item.AvatarURL = &avatarURL.String
	}

	return &item, nil
}
