package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	rotatingsavings "github.com/sonbn-225/goen-api-v2/internal/domains/rotating_savings"
)

type RotatingSavingsRepository struct {
	db *pgxpool.Pool
}

var _ rotatingsavings.Repository = (*RotatingSavingsRepository)(nil)

func NewRotatingSavingsRepository(db *pgxpool.Pool) *RotatingSavingsRepository {
	return &RotatingSavingsRepository{db: db}
}

func (r *RotatingSavingsRepository) CreateGroup(ctx context.Context, group rotatingsavings.RotatingSavingsGroup) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "create_group", "group_id", group.ID, "user_id", group.UserID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO rotating_savings_groups (
			id,
			user_id,
			account_id,
			name,
			member_count,
			user_slots,
			contribution_amount,
			payout_cycle_no,
			fixed_interest_amount,
			cycle_frequency,
			start_date,
			status,
			created_at,
			updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,
			$10::rotating_savings_cycle_frequency,
			$11,
			$12::rotating_savings_group_status,
			$13,
			$14
		)
	`,
		group.ID,
		group.UserID,
		group.AccountID,
		group.Name,
		group.MemberCount,
		group.UserSlots,
		group.ContributionAmount,
		group.PayoutCycleNo,
		group.FixedInterestAmount,
		group.CycleFrequency,
		group.StartDate,
		group.Status,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_rotating_savings_create_group_failed", "error", err)
		return err
	}
	logger.Info("repo_rotating_savings_create_group_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) GetGroup(ctx context.Context, userID, groupID string) (*rotatingsavings.RotatingSavingsGroup, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "get_group", "group_id", groupID, "user_id", userID)
	row := r.db.QueryRow(ctx, `
		SELECT
			g.id,
			g.user_id,
			g.account_id,
			g.name,
			a.currency,
			g.member_count,
			g.user_slots,
			g.contribution_amount,
			g.payout_cycle_no,
			g.fixed_interest_amount,
			g.cycle_frequency::text,
			to_char(g.start_date, 'YYYY-MM-DD'),
			g.status::text,
			g.created_at,
			g.updated_at
		FROM rotating_savings_groups g
		JOIN accounts a ON a.id = g.account_id
		WHERE g.id = $1
		  AND g.user_id = $2
	`, groupID, userID)

	var item rotatingsavings.RotatingSavingsGroup
	var payoutCycleNull sql.NullInt32
	var fixedInterestNull sql.NullFloat64
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.AccountID,
		&item.Name,
		&item.Currency,
		&item.MemberCount,
		&item.UserSlots,
		&item.ContributionAmount,
		&payoutCycleNull,
		&fixedInterestNull,
		&item.CycleFrequency,
		&item.StartDate,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_rotating_savings_get_group_failed", "error", err)
		return nil, err
	}

	if payoutCycleNull.Valid {
		v := int(payoutCycleNull.Int32)
		item.PayoutCycleNo = &v
	}
	if fixedInterestNull.Valid {
		v := fixedInterestNull.Float64
		item.FixedInterestAmount = &v
	}

	logger.Info("repo_rotating_savings_get_group_succeeded")
	return &item, nil
}

func (r *RotatingSavingsRepository) UpdateGroup(ctx context.Context, group rotatingsavings.RotatingSavingsGroup) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "update_group", "group_id", group.ID, "user_id", group.UserID)
	cmd, err := r.db.Exec(ctx, `
		UPDATE rotating_savings_groups
		SET
			account_id = $3,
			name = $4,
			member_count = $5,
			user_slots = $6,
			contribution_amount = $7,
			payout_cycle_no = $8,
			fixed_interest_amount = $9,
			cycle_frequency = $10::rotating_savings_cycle_frequency,
			start_date = $11,
			status = $12::rotating_savings_group_status,
			updated_at = $13
		WHERE id = $1
		  AND user_id = $2
	`,
		group.ID,
		group.UserID,
		group.AccountID,
		group.Name,
		group.MemberCount,
		group.UserSlots,
		group.ContributionAmount,
		group.PayoutCycleNo,
		group.FixedInterestAmount,
		group.CycleFrequency,
		group.StartDate,
		group.Status,
		group.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_rotating_savings_update_group_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}

	logger.Info("repo_rotating_savings_update_group_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) DeleteGroup(ctx context.Context, userID, groupID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "delete_group", "group_id", groupID, "user_id", userID)
	cmd, err := r.db.Exec(ctx, `
		DELETE FROM rotating_savings_groups
		WHERE id = $1
		  AND user_id = $2
	`, groupID, userID)
	if err != nil {
		logger.Error("repo_rotating_savings_delete_group_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}

	logger.Info("repo_rotating_savings_delete_group_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) ListGroups(ctx context.Context, userID string) ([]rotatingsavings.RotatingSavingsGroup, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "list_groups", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			g.id,
			g.user_id,
			g.account_id,
			g.name,
			a.currency,
			g.member_count,
			g.user_slots,
			g.contribution_amount,
			g.payout_cycle_no,
			g.fixed_interest_amount,
			g.cycle_frequency::text,
			to_char(g.start_date, 'YYYY-MM-DD'),
			g.status::text,
			g.created_at,
			g.updated_at
		FROM rotating_savings_groups g
		JOIN accounts a ON a.id = g.account_id
		WHERE g.user_id = $1
		ORDER BY g.start_date DESC, g.id DESC
	`, userID)
	if err != nil {
		logger.Error("repo_rotating_savings_list_groups_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]rotatingsavings.RotatingSavingsGroup, 0)
	for rows.Next() {
		var item rotatingsavings.RotatingSavingsGroup
		var payoutCycleNull sql.NullInt32
		var fixedInterestNull sql.NullFloat64
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.AccountID,
			&item.Name,
			&item.Currency,
			&item.MemberCount,
			&item.UserSlots,
			&item.ContributionAmount,
			&payoutCycleNull,
			&fixedInterestNull,
			&item.CycleFrequency,
			&item.StartDate,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			logger.Error("repo_rotating_savings_list_groups_failed", "error", err)
			return nil, err
		}
		if payoutCycleNull.Valid {
			v := int(payoutCycleNull.Int32)
			item.PayoutCycleNo = &v
		}
		if fixedInterestNull.Valid {
			v := fixedInterestNull.Float64
			item.FixedInterestAmount = &v
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_rotating_savings_list_groups_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_rotating_savings_list_groups_succeeded", "count", len(items))
	return items, nil
}

func (r *RotatingSavingsRepository) CreateContribution(ctx context.Context, contribution rotatingsavings.RotatingSavingsContribution) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "create_contribution", "contribution_id", contribution.ID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO rotating_savings_contributions (
			id,
			group_id,
			transaction_id,
			kind,
			cycle_no,
			due_date,
			amount,
			slots_taken,
			collected_fee_per_slot,
			occurred_at,
			note,
			created_at
		) VALUES (
			$1,$2,$3,$4::rotating_savings_contribution_kind,$5,$6,$7,$8,$9,$10,$11,$12
		)
	`,
		contribution.ID,
		contribution.GroupID,
		contribution.TransactionID,
		contribution.Kind,
		contribution.CycleNo,
		contribution.DueDate,
		contribution.Amount,
		contribution.SlotsTaken,
		contribution.CollectedFeePerSlot,
		contribution.OccurredAt,
		contribution.Note,
		contribution.CreatedAt,
	)
	if err != nil {
		logger.Error("repo_rotating_savings_create_contribution_failed", "error", err)
		return err
	}
	logger.Info("repo_rotating_savings_create_contribution_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) GetContribution(ctx context.Context, userID, contributionID string) (*rotatingsavings.RotatingSavingsContribution, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "get_contribution", "contribution_id", contributionID, "user_id", userID)
	row := r.db.QueryRow(ctx, `
		SELECT
			c.id,
			c.group_id,
			c.transaction_id,
			c.kind::text,
			c.cycle_no,
			CASE WHEN c.due_date IS NULL THEN NULL ELSE to_char(c.due_date, 'YYYY-MM-DD') END,
			c.amount,
			c.slots_taken,
			c.collected_fee_per_slot,
			c.occurred_at,
			c.note,
			c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.id = $1
		  AND g.user_id = $2
	`, contributionID, userID)

	var item rotatingsavings.RotatingSavingsContribution
	var cycleNoNull sql.NullInt32
	var dueDateNull sql.NullString
	var noteNull sql.NullString
	if err := row.Scan(
		&item.ID,
		&item.GroupID,
		&item.TransactionID,
		&item.Kind,
		&cycleNoNull,
		&dueDateNull,
		&item.Amount,
		&item.SlotsTaken,
		&item.CollectedFeePerSlot,
		&item.OccurredAt,
		&noteNull,
		&item.CreatedAt,
	); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_rotating_savings_get_contribution_failed", "error", err)
		return nil, err
	}

	if cycleNoNull.Valid {
		v := int(cycleNoNull.Int32)
		item.CycleNo = &v
	}
	if dueDateNull.Valid {
		v := dueDateNull.String
		item.DueDate = &v
	}
	if noteNull.Valid {
		v := noteNull.String
		item.Note = &v
	}
	logger.Info("repo_rotating_savings_get_contribution_succeeded")
	return &item, nil
}

func (r *RotatingSavingsRepository) ListContributions(ctx context.Context, userID, groupID string) ([]rotatingsavings.RotatingSavingsContribution, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "list_contributions", "group_id", groupID, "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			c.id,
			c.group_id,
			c.transaction_id,
			c.kind::text,
			c.cycle_no,
			CASE WHEN c.due_date IS NULL THEN NULL ELSE to_char(c.due_date, 'YYYY-MM-DD') END,
			c.amount,
			c.slots_taken,
			c.collected_fee_per_slot,
			c.occurred_at,
			c.note,
			c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.group_id = $1
		  AND g.user_id = $2
		ORDER BY c.occurred_at DESC, c.id DESC
	`, groupID, userID)
	if err != nil {
		logger.Error("repo_rotating_savings_list_contributions_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]rotatingsavings.RotatingSavingsContribution, 0)
	for rows.Next() {
		var item rotatingsavings.RotatingSavingsContribution
		var cycleNoNull sql.NullInt32
		var dueDateNull sql.NullString
		var noteNull sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.GroupID,
			&item.TransactionID,
			&item.Kind,
			&cycleNoNull,
			&dueDateNull,
			&item.Amount,
			&item.SlotsTaken,
			&item.CollectedFeePerSlot,
			&item.OccurredAt,
			&noteNull,
			&item.CreatedAt,
		); err != nil {
			logger.Error("repo_rotating_savings_list_contributions_failed", "error", err)
			return nil, err
		}
		if cycleNoNull.Valid {
			v := int(cycleNoNull.Int32)
			item.CycleNo = &v
		}
		if dueDateNull.Valid {
			v := dueDateNull.String
			item.DueDate = &v
		}
		if noteNull.Valid {
			v := noteNull.String
			item.Note = &v
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_rotating_savings_list_contributions_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_rotating_savings_list_contributions_succeeded", "count", len(items))
	return items, nil
}

func (r *RotatingSavingsRepository) DeleteContribution(ctx context.Context, userID, contributionID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "delete_contribution", "contribution_id", contributionID, "user_id", userID)
	cmd, err := r.db.Exec(ctx, `
		DELETE FROM rotating_savings_contributions c
		USING rotating_savings_groups g
		WHERE c.id = $1
		  AND g.id = c.group_id
		  AND g.user_id = $2
	`, contributionID, userID)
	if err != nil {
		logger.Error("repo_rotating_savings_delete_contribution_failed", "error", err)
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindNotFound, "rotating savings contribution not found")
	}
	logger.Info("repo_rotating_savings_delete_contribution_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) CreateAuditLog(ctx context.Context, log rotatingsavings.RotatingSavingsAuditLog) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "create_audit_log", "log_id", log.ID, "user_id", log.UserID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO rotating_savings_audit_logs (
			id,
			user_id,
			group_id,
			action,
			details,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`,
		log.ID,
		log.UserID,
		log.GroupID,
		log.Action,
		log.Details,
		log.CreatedAt,
	)
	if err != nil {
		logger.Error("repo_rotating_savings_create_audit_log_failed", "error", err)
		return err
	}
	logger.Info("repo_rotating_savings_create_audit_log_succeeded")
	return nil
}

func (r *RotatingSavingsRepository) ListAuditLogs(ctx context.Context, userID, groupID string) ([]rotatingsavings.RotatingSavingsAuditLog, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "list_audit_logs", "group_id", groupID, "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			group_id,
			action,
			details,
			created_at
		FROM rotating_savings_audit_logs
		WHERE group_id = $1
		  AND user_id = $2
		ORDER BY created_at DESC, id DESC
	`, groupID, userID)
	if err != nil {
		logger.Error("repo_rotating_savings_list_audit_logs_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]rotatingsavings.RotatingSavingsAuditLog, 0)
	for rows.Next() {
		var item rotatingsavings.RotatingSavingsAuditLog
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.GroupID,
			&item.Action,
			&item.Details,
			&item.CreatedAt,
		); err != nil {
			logger.Error("repo_rotating_savings_list_audit_logs_failed", "error", err)
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_rotating_savings_list_audit_logs_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_rotating_savings_list_audit_logs_succeeded", "count", len(items))
	return items, nil
}

func (r *RotatingSavingsRepository) SoftDeleteTransactionForUser(ctx context.Context, userID, transactionID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "rotating_savings", "operation", "soft_delete_transaction_for_user", "transaction_id", transactionID, "user_id", userID)
	_, err := r.db.Exec(ctx, `
		UPDATE transactions
		SET deleted_at = $3,
			updated_at = $3,
			updated_by = $2
		WHERE id = $1
		  AND created_by = $2
	`, transactionID, userID, time.Now().UTC())
	if err != nil {
		logger.Error("repo_rotating_savings_soft_delete_transaction_failed", "error", err)
		return err
	}
	logger.Info("repo_rotating_savings_soft_delete_transaction_succeeded")
	return nil
}
