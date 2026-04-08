package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func (r *DebtRepo) ListPublicParticipants(ctx context.Context, userID uuid.UUID) ([]string, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Find the contact names of all debts where Direction=Lent and outstanding_principal > 0
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT COALESCE(u.display_name, c.name) 
		FROM debts d
		LEFT JOIN contacts c ON d.contact_id = c.id
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE d.user_id = $1 
		  AND d.deleted_at IS NULL 
		  AND d.direction = 'lent' 
		  AND d.outstanding_principal > 0
		  AND COALESCE(u.display_name, c.name) IS NOT NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (r *DebtRepo) ListPublicDebtsByParticipant(ctx context.Context, userID uuid.UUID, participantName string) ([]entity.PublicDebt, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT d.id, to_char(d.created_at, 'YYYY-MM-DD'), d.outstanding_principal::text, d.status
		FROM debts d
		LEFT JOIN contacts c ON d.contact_id = c.id
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE d.user_id = $1 
		  AND d.deleted_at IS NULL
		  AND d.direction = 'lent'
		  AND d.outstanding_principal > 0
		  AND COALESCE(u.display_name, c.name) = $2
	`, userID, participantName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []entity.PublicDebt
	for rows.Next() {
		var d entity.PublicDebt
		if err := rows.Scan(&d.ID, &d.CreatedAt, &d.ShareAmount, &d.Status); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, nil
}
