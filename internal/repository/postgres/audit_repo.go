package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type AuditRepo struct {
	BaseRepo
}

func NewAuditRepo(db *database.Postgres) *AuditRepo {
	return &AuditRepo{BaseRepo: *NewBaseRepo(db)}
}

func (r *AuditRepo) RecordTx(ctx context.Context, tx pgx.Tx, l entity.AuditLog) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	metadataJSON, _ := json.Marshal(l.Metadata)

	_, err = q.Exec(ctx, `
		INSERT INTO audit_logs (
			id, occurred_at, actor_user_id, account_id, resource_type, resource_id, action, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, l.ID, l.OccurredAt, l.ActorUserID, l.AccountID, l.ResourceType, l.ResourceID, l.Action, metadataJSON)
	
	return err
}

func (r *AuditRepo) ListTx(ctx context.Context, tx pgx.Tx, filter entity.AuditLogFilter) ([]entity.AuditLog, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, occurred_at, actor_user_id, account_id, resource_type, resource_id, action, metadata
		FROM audit_logs
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if filter.ActorUserID != nil {
		query += fmt.Sprintf(" AND actor_user_id = $%d", argIdx)
		args = append(args, *filter.ActorUserID)
		argIdx++
	}
	if filter.AccountID != nil {
		query += fmt.Sprintf(" AND account_id = $%d", argIdx)
		args = append(args, *filter.AccountID)
		argIdx++
	}
	if filter.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, *filter.ResourceType)
		argIdx++
	}
	if filter.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, *filter.ResourceID)
		argIdx++
	}
	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filter.Action)
		argIdx++
	}

	query += " ORDER BY occurred_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.AuditLog
	for rows.Next() {
		var l entity.AuditLog
		var metadataJSON []byte
		err := rows.Scan(
			&l.ID, &l.OccurredAt, &l.ActorUserID, &l.AccountID, &l.ResourceType, &l.ResourceID, &l.Action, &metadataJSON,
		)
		if err != nil {
			return nil, err
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &l.Metadata)
		}
		logs = append(logs, l)
	}
	
	return logs, nil
}

func (r *AuditRepo) DeleteOldLogs(ctx context.Context, tx pgx.Tx, before time.Time) (int64, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return 0, err
	}

	sql := `DELETE FROM audit_logs WHERE occurred_at < $1`
	
	res, err := q.Exec(ctx, sql, before)
	if err != nil {
		return 0, err
	}
	
	return res.RowsAffected(), nil
}
