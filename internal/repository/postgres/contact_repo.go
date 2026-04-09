package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type ContactRepo struct {
	BaseRepo
}

func NewContactRepo(db *database.Postgres) *ContactRepo {
	return &ContactRepo{BaseRepo: *NewBaseRepo(db)}
}

// --- Nhóm 1: Truy vấn danh bạ (Flexible Tx) ---

func (r *ContactRepo) GetContactTx(ctx context.Context, tx pgx.Tx, userID, contactID uuid.UUID) (*entity.Contact, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
	`, ContactColumnsSQL), contactID, userID)

	return ScanContact(row)
}

func (r *ContactRepo) ListContactsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Contact, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, fmt.Sprintf(`
		SELECT %s
		FROM contacts c
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE c.user_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name ASC
	`, ContactColumnsSQL), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Contact
	for rows.Next() {
		c, err := ScanContact(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *c)
	}
	return results, nil
}

func (r *ContactRepo) FindUserByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*entity.User, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, "SELECT id, email, display_name, avatar_url, created_at, updated_at FROM users WHERE email = $1", email)
	var u entity.User
	err = row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}
	return &u, nil
}

func (r *ContactRepo) FindUserByPhoneTx(ctx context.Context, tx pgx.Tx, phone string) (*entity.User, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, "SELECT id, email, display_name, avatar_url, created_at, updated_at FROM users WHERE phone = $1", phone)
	var u entity.User
	err = row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}
	return &u, nil
}

// --- Nhóm 2: Thao tác ghi & Nhất quán (Transactional) ---

func (r *ContactRepo) CreateContactTx(ctx context.Context, tx pgx.Tx, c entity.Contact) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO contacts (id, user_id, name, email, phone, avatar_url, linked_user_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, c.ID, c.UserID, c.Name, c.Email, c.Phone, c.AvatarURL, c.LinkedUserID, c.Notes, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *ContactRepo) UpdateContactTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, c entity.Contact) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		UPDATE contacts
		SET name = $1, email = $2, phone = $3, avatar_url = $4, linked_user_id = $5, notes = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9 AND deleted_at IS NULL
	`, c.Name, c.Email, c.Phone, c.AvatarURL, c.LinkedUserID, c.Notes, c.UpdatedAt, c.ID, userID)
	return err
}

func (r *ContactRepo) DeleteContactTx(ctx context.Context, tx pgx.Tx, userID, contactID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `UPDATE contacts SET deleted_at = NOW() WHERE id = $1 AND user_id = $2`, contactID, userID)
	return err
}
