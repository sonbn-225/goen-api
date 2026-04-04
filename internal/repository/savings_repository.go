package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/savings"
)

type SavingsRepository struct {
	db *pgxpool.Pool
}

var _ savings.Repository = (*SavingsRepository)(nil)

func NewSavingsRepository(db *pgxpool.Pool) *SavingsRepository {
	return &SavingsRepository{db: db}
}

func (r *SavingsRepository) GetAccountForUser(ctx context.Context, userID, accountID string) (*savings.AccountRef, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "get_account_for_user", "user_id", userID, "account_id", accountID)
	row := r.db.QueryRow(ctx, `
		SELECT
			a.id,
			a.name,
			a.account_type::text,
			a.currency,
			a.parent_account_id
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		  AND a.id = $2
	`, userID, accountID)

	var item savings.AccountRef
	if err := row.Scan(&item.ID, &item.Name, &item.Type, &item.Currency, &item.ParentAccountID); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_savings_get_account_for_user_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_savings_get_account_for_user_succeeded")
	return &item, nil
}

func (r *SavingsRepository) CreateLinkedSavingsAccount(ctx context.Context, userID, parentAccountID, accountName, currency string) (*savings.AccountRef, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "create_linked_savings_account", "user_id", userID, "parent_account_id", parentAccountID)
	now := time.Now().UTC()
	accountID := uuid.NewString()

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error("repo_savings_create_linked_account_failed", "error", err)
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	cmd, err := tx.Exec(ctx, `
		INSERT INTO accounts (
			id,
			name,
			account_type,
			currency,
			parent_account_id,
			created_at,
			updated_at,
			created_by,
			updated_by
		)
		SELECT $1, $2, 'savings'::account_type, $3, $4, $5, $6, $7, $8
		WHERE EXISTS (
			SELECT 1
			FROM user_accounts ua
			JOIN accounts a ON a.id = ua.account_id
			WHERE ua.user_id = $7
			  AND ua.account_id = $4
			  AND ua.status = 'active'
			  AND a.deleted_at IS NULL
		)
	`, accountID, accountName, currency, parentAccountID, now, now, userID, userID)
	if err != nil {
		logger.Error("repo_savings_create_linked_account_failed", "error", err)
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		logger.Warn("repo_savings_create_linked_account_failed", "reason", "parent account not found")
		return nil, apperrors.New(apperrors.KindNotFound, "parent account not found")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_accounts (
			id,
			account_id,
			user_id,
			permission,
			status,
			created_at,
			updated_at,
			created_by,
			updated_by
		) VALUES ($1, $2, $3, 'owner', 'active', $4, $5, $6, $7)
	`, uuid.NewString(), accountID, userID, now, now, userID, userID)
	if err != nil {
		logger.Error("repo_savings_create_linked_account_failed", "error", err)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("repo_savings_create_linked_account_failed", "error", err)
		return nil, err
	}
	committed = true

	item := &savings.AccountRef{
		ID:              accountID,
		Name:            accountName,
		Type:            "savings",
		Currency:        currency,
		ParentAccountID: &parentAccountID,
	}
	logger.Info("repo_savings_create_linked_account_succeeded", "account_id", accountID)
	return item, nil
}

func (r *SavingsRepository) DeleteAccountForUser(ctx context.Context, userID, accountID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "delete_account_for_user", "user_id", userID, "account_id", accountID)
	cmd, err := r.db.Exec(ctx, `
		DELETE FROM accounts a
		USING user_accounts ua
		WHERE a.id = $2
		  AND ua.account_id = a.id
		  AND ua.user_id = $1
		  AND ua.status = 'active'
	`, userID, accountID)
	if err != nil {
		logger.Error("repo_savings_delete_account_for_user_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "account not found")
	}

	logger.Info("repo_savings_delete_account_for_user_succeeded")
	return nil
}

func (r *SavingsRepository) CreateSavingsInstrument(ctx context.Context, userID string, item savings.SavingsInstrument) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "create_savings_instrument", "user_id", userID, "instrument_id", item.ID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO savings_instruments (
			id,
			savings_account_id,
			parent_account_id,
			principal,
			interest_rate,
			term_months,
			start_date,
			maturity_date,
			auto_renew,
			accrued_interest,
			status,
			closed_at,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::savings_instrument_status,$12,$13,$14)
	`,
		item.ID,
		item.SavingsAccountID,
		item.ParentAccountID,
		item.Principal,
		item.InterestRate,
		item.TermMonths,
		item.StartDate,
		item.MaturityDate,
		item.AutoRenew,
		item.AccruedInterest,
		item.Status,
		item.ClosedAt,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_savings_create_savings_instrument_failed", "error", err)
		return err
	}

	logger.Info("repo_savings_create_savings_instrument_succeeded")
	return nil
}

func (r *SavingsRepository) GetSavingsInstrument(ctx context.Context, userID, instrumentID string) (*savings.SavingsInstrument, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "get_savings_instrument", "user_id", userID, "instrument_id", instrumentID)
	row := r.db.QueryRow(ctx, `
		SELECT
			si.id,
			si.savings_account_id,
			si.parent_account_id,
			si.principal::text,
			CASE WHEN si.interest_rate IS NULL THEN NULL ELSE si.interest_rate::text END,
			si.term_months,
			CASE WHEN si.start_date IS NULL THEN NULL ELSE to_char(si.start_date, 'YYYY-MM-DD') END,
			CASE WHEN si.maturity_date IS NULL THEN NULL ELSE to_char(si.maturity_date, 'YYYY-MM-DD') END,
			si.auto_renew,
			si.accrued_interest::text,
			si.status::text,
			si.closed_at,
			si.created_at,
			si.updated_at
		FROM savings_instruments si
		JOIN accounts a ON a.id = si.savings_account_id
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE si.id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
	`, instrumentID, userID)

	var item savings.SavingsInstrument
	var interestNull sql.NullString
	var termNull sql.NullInt32
	var startNull sql.NullString
	var maturityNull sql.NullString
	var closedNull sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.SavingsAccountID,
		&item.ParentAccountID,
		&item.Principal,
		&interestNull,
		&termNull,
		&startNull,
		&maturityNull,
		&item.AutoRenew,
		&item.AccruedInterest,
		&item.Status,
		&closedNull,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_savings_get_savings_instrument_failed", "error", err)
		return nil, err
	}

	if interestNull.Valid {
		item.InterestRate = &interestNull.String
	}
	if termNull.Valid {
		v := int(termNull.Int32)
		item.TermMonths = &v
	}
	if startNull.Valid {
		item.StartDate = &startNull.String
	}
	if maturityNull.Valid {
		item.MaturityDate = &maturityNull.String
	}
	if closedNull.Valid {
		v := closedNull.Time
		item.ClosedAt = &v
	}

	logger.Info("repo_savings_get_savings_instrument_succeeded")
	return &item, nil
}

func (r *SavingsRepository) ListSavingsInstruments(ctx context.Context, userID string) ([]savings.SavingsInstrument, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "list_savings_instruments", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			si.id,
			si.savings_account_id,
			si.parent_account_id,
			si.principal::text,
			CASE WHEN si.interest_rate IS NULL THEN NULL ELSE si.interest_rate::text END,
			si.term_months,
			CASE WHEN si.start_date IS NULL THEN NULL ELSE to_char(si.start_date, 'YYYY-MM-DD') END,
			CASE WHEN si.maturity_date IS NULL THEN NULL ELSE to_char(si.maturity_date, 'YYYY-MM-DD') END,
			si.auto_renew,
			si.accrued_interest::text,
			si.status::text,
			si.closed_at,
			si.created_at,
			si.updated_at
		FROM savings_instruments si
		JOIN accounts a ON a.id = si.savings_account_id
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		ORDER BY si.created_at DESC, si.id DESC
	`, userID)
	if err != nil {
		logger.Error("repo_savings_list_savings_instruments_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]savings.SavingsInstrument, 0)
	for rows.Next() {
		var item savings.SavingsInstrument
		var interestNull sql.NullString
		var termNull sql.NullInt32
		var startNull sql.NullString
		var maturityNull sql.NullString
		var closedNull sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.SavingsAccountID,
			&item.ParentAccountID,
			&item.Principal,
			&interestNull,
			&termNull,
			&startNull,
			&maturityNull,
			&item.AutoRenew,
			&item.AccruedInterest,
			&item.Status,
			&closedNull,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			logger.Error("repo_savings_list_savings_instruments_failed", "error", err)
			return nil, err
		}

		if interestNull.Valid {
			item.InterestRate = &interestNull.String
		}
		if termNull.Valid {
			v := int(termNull.Int32)
			item.TermMonths = &v
		}
		if startNull.Valid {
			item.StartDate = &startNull.String
		}
		if maturityNull.Valid {
			item.MaturityDate = &maturityNull.String
		}
		if closedNull.Valid {
			v := closedNull.Time
			item.ClosedAt = &v
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_savings_list_savings_instruments_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_savings_list_savings_instruments_succeeded", "count", len(out))
	return out, nil
}

func (r *SavingsRepository) UpdateSavingsInstrument(ctx context.Context, userID string, item savings.SavingsInstrument) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "update_savings_instrument", "user_id", userID, "instrument_id", item.ID)
	cmd, err := r.db.Exec(ctx, `
		UPDATE savings_instruments si
		SET
			principal = $3,
			interest_rate = $4,
			term_months = $5,
			start_date = $6,
			maturity_date = $7,
			auto_renew = $8,
			accrued_interest = $9,
			status = $10::savings_instrument_status,
			closed_at = $11,
			updated_at = $12
		FROM user_accounts ua
		WHERE ua.account_id = si.savings_account_id
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		  AND si.id = $1
	`,
		item.ID,
		userID,
		item.Principal,
		item.InterestRate,
		item.TermMonths,
		item.StartDate,
		item.MaturityDate,
		item.AutoRenew,
		item.AccruedInterest,
		item.Status,
		item.ClosedAt,
		item.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_savings_update_savings_instrument_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "savings instrument not found")
	}

	logger.Info("repo_savings_update_savings_instrument_succeeded")
	return nil
}

func (r *SavingsRepository) DeleteSavingsInstrument(ctx context.Context, userID, instrumentID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "savings", "operation", "delete_savings_instrument", "user_id", userID, "instrument_id", instrumentID)
	cmd, err := r.db.Exec(ctx, `
		DELETE FROM savings_instruments si
		USING user_accounts ua
		WHERE ua.account_id = si.savings_account_id
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		  AND si.id = $1
	`, instrumentID, userID)
	if err != nil {
		logger.Error("repo_savings_delete_savings_instrument_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "savings instrument not found")
	}

	logger.Info("repo_savings_delete_savings_instrument_succeeded")
	return nil
}
