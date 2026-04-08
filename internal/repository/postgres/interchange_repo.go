package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type InterchangeRepo struct {
	db *database.Postgres
}

func NewInterchangeRepo(db *database.Postgres) *InterchangeRepo {
	return &InterchangeRepo{db: db}
}

// --- Nhóm 1: Truy vấn dữ liệu chờ (Read-only Optimized) ---

func (r *InterchangeRepo) ListStagedImports(ctx context.Context, userID uuid.UUID, resourceType string) ([]entity.StagedImport, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, resource_type, source, external_id, data, metadata, status, created_at, updated_at
		FROM staged_imports
		WHERE user_id = $1 AND resource_type = $2
		ORDER BY created_at DESC, id DESC
	`, userID, resourceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.StagedImport
	for rows.Next() {
		var si entity.StagedImport
		var dataB, metaB []byte
		err := rows.Scan(
			&si.ID, &si.ResourceType, &si.Source, &si.ExternalID,
			&dataB, &metaB, &si.Status, &si.CreatedAt, &si.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(dataB, &si.Data)
		_ = json.Unmarshal(metaB, &si.Metadata)
		results = append(results, si)
	}
	return results, nil
}

func (r *InterchangeRepo) GetStagedImport(ctx context.Context, userID, id uuid.UUID) (*entity.StagedImport, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var si entity.StagedImport
	var dataB, metaB []byte
	err = pool.QueryRow(ctx, `
		SELECT id, resource_type, source, external_id, data, metadata, status, created_at, updated_at
		FROM staged_imports
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&si.ID, &si.ResourceType, &si.Source, &si.ExternalID,
		&dataB, &metaB, &si.Status, &si.CreatedAt, &si.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(dataB, &si.Data)
	_ = json.Unmarshal(metaB, &si.Metadata)
	return &si, nil
}

func (r *InterchangeRepo) ListImportRules(ctx context.Context, userID uuid.UUID, resourceType string) ([]entity.StagedImportRule, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, resource_type, rule_type, match_key, match_value, mapped_id, created_at, updated_at
		FROM staged_import_rules
		WHERE user_id = $1 AND resource_type = $2
		ORDER BY rule_type ASC, match_value ASC
	`, userID, resourceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.StagedImportRule
	for rows.Next() {
		var rule entity.StagedImportRule
		if err := rows.Scan(
			&rule.ID, &rule.ResourceType, &rule.RuleType, &rule.MatchKey, &rule.MatchValue, &rule.MappedID,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, rule)
	}
	return results, nil
}

// --- Nhóm 2: Thao tác ghi & Nhất quán (Transactional) ---

func (r *InterchangeRepo) UpsertStagedImportsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, items []entity.StagedImportCreate) ([]entity.StagedImport, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	now := utils.Now()
	created := make([]entity.StagedImport, 0, len(items))

	for _, item := range items {
		id := uuid.New()
		dataBytes, _ := json.Marshal(item.Data)
		metaBytes, _ := json.Marshal(item.Metadata)

		_, err := q.Exec(ctx, `
			INSERT INTO staged_imports (
				id, user_id, resource_type, source, external_id, data, metadata, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, id, userID, item.ResourceType, item.Source, item.ExternalID, dataBytes, metaBytes, "pending", now, now)
		if err != nil {
			return nil, err
		}

		created = append(created, entity.StagedImport{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{ID: id},
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			UserID:       userID,
			ResourceType: item.ResourceType,
			Source:       item.Source,
			ExternalID:   item.ExternalID,
			Data:         item.Data,
			Metadata:     item.Metadata,
			Status:       "pending",
		})
	}
	return created, nil
}

func (r *InterchangeRepo) PatchStagedImportTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID, patch entity.StagedImportPatch) (*entity.StagedImport, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	now := utils.Now()
	var currentMetaB []byte
	err = q.QueryRow(ctx, `SELECT metadata FROM staged_imports WHERE id = $1 AND user_id = $2`, id, userID).Scan(&currentMetaB)
	if err != nil {
		return nil, err
	}

	meta := make(map[string]any)
	_ = json.Unmarshal(currentMetaB, &meta)

	for k, v := range patch.Metadata {
		meta[k] = v
	}

	newMetaB, _ := json.Marshal(meta)
	statusClause := ""
	args := []any{newMetaB, now, id, userID}
	if patch.Status != nil {
		args = append(args, *patch.Status)
		statusClause = ", status = $5"
	}

	var si entity.StagedImport
	var dataB, metaB []byte
	query := `
		UPDATE staged_imports
		SET metadata = $1, updated_at = $2 ` + statusClause + `
		WHERE id = $3 AND user_id = $4
		RETURNING id, resource_type, source, external_id, data, metadata, status, created_at, updated_at
	`
	err = q.QueryRow(ctx, query, args...).Scan(
		&si.ID, &si.ResourceType, &si.Source, &si.ExternalID,
		&dataB, &metaB, &si.Status, &si.CreatedAt, &si.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(dataB, &si.Data)
	_ = json.Unmarshal(metaB, &si.Metadata)
	return &si, nil
}

func (r *InterchangeRepo) DeleteStagedImportTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	ct, err := q.Exec(ctx, `DELETE FROM staged_imports WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("staged import not found")
	}
	return nil
}

func (r *InterchangeRepo) DeleteAllStagedImportsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, resourceType string) (int64, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return 0, err
	}

	ct, err := q.Exec(ctx, `DELETE FROM staged_imports WHERE user_id = $1 AND resource_type = $2`, userID, resourceType)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (r *InterchangeRepo) UpsertImportRulesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, rules []entity.StagedImportRuleUpsert) ([]entity.StagedImportRule, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	now := utils.Now()
	out := make([]entity.StagedImportRule, 0, len(rules))

	for _, rule := range rules {
		id := uuid.New()
		var outRule entity.StagedImportRule
		err := q.QueryRow(ctx, `
			INSERT INTO staged_import_rules (
				id, user_id, resource_type, rule_type, match_key, match_value, mapped_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (user_id, resource_type, rule_type, match_key, match_value) 
			DO UPDATE SET
				mapped_id = EXCLUDED.mapped_id,
				updated_at = EXCLUDED.updated_at
			RETURNING id, resource_type, rule_type, match_key, match_value, mapped_id, created_at, updated_at
		`, id, userID, rule.ResourceType, rule.RuleType, rule.MatchKey, rule.MatchValue, rule.MappedID, now, now).Scan(
			&outRule.ID, &outRule.ResourceType, &outRule.RuleType, &outRule.MatchKey, &outRule.MatchValue, &outRule.MappedID, &outRule.CreatedAt, &outRule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, outRule)
	}
	return out, nil
}

func (r *InterchangeRepo) DeleteImportRuleTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	ct, err := q.Exec(ctx, `DELETE FROM staged_import_rules WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("rule not found")
	}
	return nil
}
