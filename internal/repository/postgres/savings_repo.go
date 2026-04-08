package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type SavingsRepo struct {
	db *database.Postgres
}

func NewSavingsRepo(db *database.Postgres) *SavingsRepo {
	return &SavingsRepo{db: db}
}

func (r *SavingsRepo) GetSavings(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Savings, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var s entity.Savings
	var settingsJSON []byte
	err = pool.QueryRow(ctx, `
		SELECT
			a.id, a.name, a.parent_account_id, a.status, a.settings, a.created_at, a.updated_at,
			COALESCE((
				SELECT SUM(t.amount)
				FROM transactions t
				WHERE t.to_account_id = a.id AND t.type = 'income' AND t.status = 'posted' AND t.deleted_at IS NULL
			), 0)::text as accrued_interest
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE a.id = $1 AND ua.user_id = $2 AND a.account_type = 'savings' AND a.deleted_at IS NULL AND ua.status = 'active'
	`, id, userID).Scan(
		&s.ID, &s.Name, &s.ParentAccountID, &s.Status, &settingsJSON, &s.CreatedAt, &s.UpdatedAt, &s.AccruedInterest,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("savings not found")
		}
		return nil, err
	}

	r.mapSettingsToSavings(&s, settingsJSON)
	s.SavingsAccountID = s.ID // In the new model, the Savings ID is the Account ID
	return &s, nil
}

func (r *SavingsRepo) ListSavings(ctx context.Context, userID uuid.UUID) ([]entity.Savings, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			a.id, a.name, a.parent_account_id, a.status, a.settings, a.created_at, a.updated_at,
			COALESCE((
				SELECT SUM(t.amount)
				FROM transactions t
				WHERE t.to_account_id = a.id AND t.type = 'income' AND t.status = 'posted' AND t.deleted_at IS NULL
			), 0)::text as accrued_interest
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ua.user_id = $1 AND a.account_type = 'savings' AND a.deleted_at IS NULL AND ua.status = 'active'
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Savings
	for rows.Next() {
		var s entity.Savings
		var settingsJSON []byte
		if err := rows.Scan(
			&s.ID, &s.Name, &s.ParentAccountID, &s.Status, &settingsJSON, &s.CreatedAt, &s.UpdatedAt, &s.AccruedInterest,
		); err != nil {
			return nil, err
		}
		r.mapSettingsToSavings(&s, settingsJSON)
		s.SavingsAccountID = s.ID
		out = append(out, s)
	}
	return out, nil
}

func (r *SavingsRepo) UpdateSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, s entity.Savings) error {
	settings := entity.AccountSettings{
		Savings: &entity.SavingsSettings{
			Principal:    s.Principal,
			InterestRate: s.InterestRate,
			TermMonths:   s.TermMonths,
			StartDate:    s.StartDate,
			MaturityDate: s.MaturityDate,
			AutoRenew:    s.AutoRenew,
		},
	}
	settingsJSON, _ := json.Marshal(settings)

	_, err := tx.Exec(ctx, `
		UPDATE accounts a
		SET name = $3, status = $4, parent_account_id = $5, settings = $6, updated_at = NOW()
		FROM user_accounts ua
		WHERE ua.account_id = a.id AND ua.user_id = $2 AND a.id = $1 AND a.account_type = 'savings'
	`, s.ID, userID, s.Name, s.Status, s.ParentAccountID, settingsJSON)
	return err
}

func (r *SavingsRepo) DeleteSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE accounts a
		SET deleted_at = NOW(), status = $3, updated_at = NOW()
		FROM user_accounts ua
		WHERE ua.account_id = a.id AND ua.user_id = $2 AND a.id = $1 AND a.account_type = 'savings'
	`, id, userID, entity.AccountStatusDeleted)
	return err
}

func (r *SavingsRepo) mapSettingsToSavings(s *entity.Savings, settingsJSON []byte) {
	if len(settingsJSON) == 0 {
		return
	}
	var settings entity.AccountSettings
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return
	}
	if settings.Savings != nil {
		s.Principal = settings.Savings.Principal
		s.InterestRate = settings.Savings.InterestRate
		s.TermMonths = settings.Savings.TermMonths
		s.StartDate = settings.Savings.StartDate
		s.MaturityDate = settings.Savings.MaturityDate
		s.AutoRenew = settings.Savings.AutoRenew
	}
}
