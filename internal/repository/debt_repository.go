package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/debt"
)

type DebtRepository struct {
	db *pgxpool.Pool
}

func NewDebtRepository(db *pgxpool.Pool) *DebtRepository {
	return &DebtRepository{db: db}
}

func (r *DebtRepository) Create(ctx context.Context, userID string, input debt.Debt) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "create", "user_id", userID, "debt_id", input.ID)

	if input.AccountID == nil || strings.TrimSpace(*input.AccountID) == "" {
		return apperrors.New(apperrors.KindValidation, "account_id is required")
	}

	commandTag, err := r.db.Exec(ctx, `
		INSERT INTO debts (
			id,
			client_id,
			user_id,
			account_id,
			direction,
			name,
			contact_id,
			principal,
			start_date,
			due_date,
			interest_rate,
			interest_rule,
			outstanding_principal,
			accrued_interest,
			status,
			closed_at,
			created_at,
			updated_at
		)
		SELECT
			$1,
			$2,
			$3,
			$4,
			$5::debt_direction,
			$6,
			$7,
			$8::numeric,
			$9::date,
			$10::date,
			$11::numeric,
			$12::debt_interest_rule,
			$13::numeric,
			$14::numeric,
			$15::debt_status,
			$16,
			$17,
			$18
		WHERE EXISTS (
			SELECT 1
			FROM user_accounts ua
			WHERE ua.user_id = $3
			  AND ua.account_id = $4
			  AND ua.status = 'active'
		)
	`,
		input.ID,
		input.ClientID,
		userID,
		input.AccountID,
		input.Direction,
		input.Name,
		input.ContactID,
		input.Principal,
		input.StartDate,
		input.DueDate,
		input.InterestRate,
		input.InterestRule,
		input.OutstandingPrincipal,
		input.AccruedInterest,
		input.Status,
		input.ClosedAt,
		input.CreatedAt,
		input.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_debt_create_failed", "error", err)
		return err
	}

	if commandTag.RowsAffected() == 0 {
		var exists bool
		err := r.db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM accounts
				WHERE id = $1 AND deleted_at IS NULL
			)
		`, *input.AccountID).Scan(&exists)
		if err != nil {
			logger.Error("repo_debt_create_failed", "error", err)
			return err
		}
		if !exists {
			return apperrors.New(apperrors.KindNotFound, "account not found")
		}
		return apperrors.New(apperrors.KindForbidden, "account does not belong to user")
	}

	logger.Info("repo_debt_create_succeeded")
	return nil
}

func (r *DebtRepository) GetByID(ctx context.Context, userID, debtID string) (*debt.Debt, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "get_by_id", "user_id", userID, "debt_id", debtID)

	row := r.db.QueryRow(ctx, `
		SELECT
			d.id,
			d.client_id,
			d.user_id,
			d.account_id,
			d.direction::text,
			d.name,
			d.contact_id,
			COALESCE(u.display_name, c.name) AS contact_name,
			COALESCE(u.avatar_url, c.avatar_url) AS contact_avatar_url,
			d.principal::text,
			a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'),
			to_char(d.due_date, 'YYYY-MM-DD'),
			CASE WHEN d.interest_rate IS NULL THEN NULL ELSE d.interest_rate::text END,
			CASE WHEN d.interest_rule IS NULL THEN NULL ELSE d.interest_rule::text END,
			d.outstanding_principal::text,
			d.accrued_interest::text,
			d.status::text,
			d.closed_at,
			d.created_at,
			d.updated_at
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		LEFT JOIN contacts c ON c.id = d.contact_id AND c.deleted_at IS NULL
		LEFT JOIN users u ON u.id = c.linked_user_id
		WHERE d.user_id = $1
		  AND d.id = $2
	`, userID, debtID)

	item, err := scanDebt(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_debt_get_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_debt_get_succeeded")
	return item, nil
}

func (r *DebtRepository) ListByUser(ctx context.Context, userID string) ([]debt.Debt, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "list_by_user", "user_id", userID)

	rows, err := r.db.Query(ctx, `
		SELECT
			d.id,
			d.client_id,
			d.user_id,
			d.account_id,
			d.direction::text,
			d.name,
			d.contact_id,
			COALESCE(u.display_name, c.name) AS contact_name,
			COALESCE(u.avatar_url, c.avatar_url) AS contact_avatar_url,
			d.principal::text,
			a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'),
			to_char(d.due_date, 'YYYY-MM-DD'),
			CASE WHEN d.interest_rate IS NULL THEN NULL ELSE d.interest_rate::text END,
			CASE WHEN d.interest_rule IS NULL THEN NULL ELSE d.interest_rule::text END,
			d.outstanding_principal::text,
			d.accrued_interest::text,
			d.status::text,
			d.closed_at,
			d.created_at,
			d.updated_at
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		LEFT JOIN contacts c ON c.id = d.contact_id AND c.deleted_at IS NULL
		LEFT JOIN users u ON u.id = c.linked_user_id
		WHERE d.user_id = $1
		ORDER BY d.due_date ASC, d.id ASC
	`, userID)
	if err != nil {
		logger.Error("repo_debt_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]debt.Debt, 0)
	for rows.Next() {
		item, err := scanDebt(rows)
		if err != nil {
			logger.Error("repo_debt_list_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_debt_list_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_debt_list_succeeded", "count", len(items))
	return items, nil
}

func (r *DebtRepository) CreatePaymentLink(ctx context.Context, userID string, input debt.DebtPaymentLink, update debt.DebtUpdate) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "create_payment_link", "user_id", userID, "debt_id", input.DebtID, "transaction_id", input.TransactionID)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.Error("repo_debt_create_payment_link_failed", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT TRUE FROM debts WHERE id = $1 AND user_id = $2`, input.DebtID, userID).Scan(&exists); err != nil {
		if isNoRows(err) {
			return apperrors.New(apperrors.KindNotFound, "debt not found")
		}
		logger.Error("repo_debt_create_payment_link_failed", "error", err)
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO debt_payment_links (id, debt_id, transaction_id, principal_paid, interest_paid, created_at)
		VALUES ($1, $2, $3, $4::numeric, $5::numeric, $6)
	`, input.ID, input.DebtID, input.TransactionID, input.PrincipalPaid, input.InterestPaid, input.CreatedAt)
	if err != nil {
		logger.Error("repo_debt_create_payment_link_failed", "error", err)
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE debts
		SET principal = $1::numeric,
			outstanding_principal = $2::numeric,
			accrued_interest = $3::numeric,
			status = $4::debt_status,
			closed_at = $5,
			updated_at = $6
		WHERE id = $7
		  AND user_id = $8
	`, update.Principal, update.OutstandingPrincipal, update.AccruedInterest, update.Status, update.ClosedAt, update.UpdatedAt, input.DebtID, userID)
	if err != nil {
		logger.Error("repo_debt_create_payment_link_failed", "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("repo_debt_create_payment_link_failed", "error", err)
		return err
	}

	logger.Info("repo_debt_create_payment_link_succeeded")
	return nil
}

func (r *DebtRepository) ListPaymentLinks(ctx context.Context, userID, debtID string) ([]debt.DebtPaymentLink, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "list_payment_links", "user_id", userID, "debt_id", debtID)

	existing, err := r.GetByID(ctx, userID, debtID)
	if err != nil {
		logger.Error("repo_debt_list_payment_links_failed", "error", err)
		return nil, err
	}
	if existing == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "debt not found")
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			debt_id,
			transaction_id,
			CASE WHEN principal_paid IS NULL THEN NULL ELSE principal_paid::text END,
			CASE WHEN interest_paid IS NULL THEN NULL ELSE interest_paid::text END,
			created_at
		FROM debt_payment_links
		WHERE debt_id = $1
		ORDER BY created_at DESC, id DESC
	`, debtID)
	if err != nil {
		logger.Error("repo_debt_list_payment_links_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]debt.DebtPaymentLink, 0)
	for rows.Next() {
		item, err := scanDebtPaymentLink(rows)
		if err != nil {
			logger.Error("repo_debt_list_payment_links_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_debt_list_payment_links_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_debt_list_payment_links_succeeded", "count", len(items))
	return items, nil
}

func (r *DebtRepository) ListPaymentLinksByTransaction(ctx context.Context, userID, transactionID string) ([]debt.DebtPaymentLink, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "list_payment_links_by_transaction", "user_id", userID, "transaction_id", transactionID)

	rows, err := r.db.Query(ctx, `
		SELECT
			l.id,
			l.debt_id,
			l.transaction_id,
			CASE WHEN l.principal_paid IS NULL THEN NULL ELSE l.principal_paid::text END,
			CASE WHEN l.interest_paid IS NULL THEN NULL ELSE l.interest_paid::text END,
			l.created_at
		FROM debt_payment_links l
		JOIN debts d ON d.id = l.debt_id
		JOIN transactions t ON t.id = l.transaction_id
		WHERE l.transaction_id = $2
		  AND d.user_id = $1
		  AND t.deleted_at IS NULL
		  AND `+accessibleTransactionCondition("$1")+`
		ORDER BY l.created_at DESC, l.id DESC
	`, userID, transactionID)
	if err != nil {
		logger.Error("repo_debt_list_payment_links_by_transaction_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]debt.DebtPaymentLink, 0)
	for rows.Next() {
		item, err := scanDebtPaymentLink(rows)
		if err != nil {
			logger.Error("repo_debt_list_payment_links_by_transaction_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_debt_list_payment_links_by_transaction_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_debt_list_payment_links_by_transaction_succeeded", "count", len(items))
	return items, nil
}

func (r *DebtRepository) CreateInstallment(ctx context.Context, userID string, input debt.DebtInstallment) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "create_installment", "user_id", userID, "debt_id", input.DebtID, "installment_id", input.ID)

	existing, err := r.GetByID(ctx, userID, input.DebtID)
	if err != nil {
		logger.Error("repo_debt_create_installment_failed", "error", err)
		return err
	}
	if existing == nil {
		return apperrors.New(apperrors.KindNotFound, "debt not found")
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO debt_installments (id, debt_id, installment_no, due_date, amount_due, amount_paid, status)
		VALUES ($1, $2, $3, $4::date, $5::numeric, $6::numeric, $7::debt_installment_status)
	`, input.ID, input.DebtID, input.InstallmentNo, input.DueDate, input.AmountDue, input.AmountPaid, input.Status)
	if err != nil {
		logger.Error("repo_debt_create_installment_failed", "error", err)
		return err
	}

	logger.Info("repo_debt_create_installment_succeeded")
	return nil
}

func (r *DebtRepository) ListInstallments(ctx context.Context, userID, debtID string) ([]debt.DebtInstallment, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "debt", "operation", "list_installments", "user_id", userID, "debt_id", debtID)

	existing, err := r.GetByID(ctx, userID, debtID)
	if err != nil {
		logger.Error("repo_debt_list_installments_failed", "error", err)
		return nil, err
	}
	if existing == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "debt not found")
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			debt_id,
			installment_no,
			to_char(due_date, 'YYYY-MM-DD'),
			amount_due::text,
			amount_paid::text,
			status::text
		FROM debt_installments
		WHERE debt_id = $1
		ORDER BY installment_no ASC
	`, debtID)
	if err != nil {
		logger.Error("repo_debt_list_installments_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]debt.DebtInstallment, 0)
	for rows.Next() {
		item, err := scanDebtInstallment(rows)
		if err != nil {
			logger.Error("repo_debt_list_installments_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_debt_list_installments_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_debt_list_installments_succeeded", "count", len(items))
	return items, nil
}

type debtScanner interface {
	Scan(dest ...any) error
}

func scanDebt(scanner debtScanner) (*debt.Debt, error) {
	var item debt.Debt
	var clientID sql.NullString
	var accountID sql.NullString
	var name sql.NullString
	var contactID sql.NullString
	var contactName sql.NullString
	var contactAvatarURL sql.NullString
	var currency sql.NullString
	var interestRate sql.NullString
	var interestRule sql.NullString

	err := scanner.Scan(
		&item.ID,
		&clientID,
		&item.UserID,
		&accountID,
		&item.Direction,
		&name,
		&contactID,
		&contactName,
		&contactAvatarURL,
		&item.Principal,
		&currency,
		&item.StartDate,
		&item.DueDate,
		&interestRate,
		&interestRule,
		&item.OutstandingPrincipal,
		&item.AccruedInterest,
		&item.Status,
		&item.ClosedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if clientID.Valid {
		item.ClientID = &clientID.String
	}
	if accountID.Valid {
		item.AccountID = &accountID.String
	}
	if name.Valid {
		item.Name = &name.String
	}
	if contactID.Valid {
		item.ContactID = &contactID.String
	}
	if contactName.Valid {
		item.ContactName = &contactName.String
	}
	if contactAvatarURL.Valid {
		item.ContactAvatarURL = &contactAvatarURL.String
	}
	if currency.Valid {
		item.Currency = &currency.String
	}
	if interestRate.Valid {
		item.InterestRate = &interestRate.String
	}
	if interestRule.Valid {
		item.InterestRule = &interestRule.String
	}

	return &item, nil
}

func scanDebtPaymentLink(scanner debtScanner) (*debt.DebtPaymentLink, error) {
	var item debt.DebtPaymentLink
	err := scanner.Scan(
		&item.ID,
		&item.DebtID,
		&item.TransactionID,
		&item.PrincipalPaid,
		&item.InterestPaid,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func scanDebtInstallment(scanner debtScanner) (*debt.DebtInstallment, error) {
	var item debt.DebtInstallment
	err := scanner.Scan(
		&item.ID,
		&item.DebtID,
		&item.InstallmentNo,
		&item.DueDate,
		&item.AmountDue,
		&item.AmountPaid,
		&item.Status,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

var _ debt.Repository = (*DebtRepository)(nil)
