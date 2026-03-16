package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type AuditRepo struct {
	db *Postgres
}

func NewAuditRepo(db *Postgres) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) ListAuditEventsForAccount(ctx context.Context, userID string, accountID string, limit int) ([]domain.AuditEvent, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Permission check (viewer/editor/owner)
	var one int
	if err := pool.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1 AND ua.account_id = $2 AND ua.status = 'active'
	`, userID, accountID).Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrAuditForbidden
		}
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT ae.id, ae.account_id, ae.actor_user_id, ae.action, ae.entity_type, ae.entity_id, ae.occurred_at, ae.diff
		FROM audit_events ae
		WHERE ae.account_id = $1
		ORDER BY ae.occurred_at DESC
		LIMIT $2
	`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var it domain.AuditEvent
		var diffRaw []byte
		var occurredAt time.Time
		if err := rows.Scan(&it.ID, &it.AccountID, &it.ActorUserID, &it.Action, &it.EntityType, &it.EntityID, &occurredAt, &diffRaw); err != nil {
			return nil, err
		}
		it.OccurredAt = occurredAt
		if len(diffRaw) > 0 {
			var diff map[string]any
			if err := json.Unmarshal(diffRaw, &diff); err == nil {
				it.Diff = diff
			}
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

