package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

// withTx executes a function within a database transaction.
// It handles begin, commit, and rollback automatically.
func withTx(ctx context.Context, pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// requireAccountPermission checks if a user has permission to access an account.
// If requireWrite is true, only owner/editor permissions are allowed.
func requireAccountPermission(ctx context.Context, dbtx pgx.Tx, userID, accountID string, requireWrite bool) error {
	var one int
	permClause := ""
	if requireWrite {
		permClause = " AND ua.permission IN ('owner','editor')"
	}
	err := dbtx.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1 AND ua.account_id = $2 AND ua.status = 'active'`+permClause+`
	`, userID, accountID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrTransactionForbidden
		}
		return err
	}
	return nil
}

// requireAccountActive checks if an account is active (not closed).
func requireAccountActive(ctx context.Context, dbtx pgx.Tx, accountID string) error {
	var status string
	err := dbtx.QueryRow(ctx, `
		SELECT status
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL
	`, accountID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrAccountNotFound
		}
		return err
	}
	if status != "active" {
		return apperrors.ErrAccountClosed
	}
	return nil
}

// getAccountCurrency returns the currency code for an account.
func getAccountCurrency(ctx context.Context, dbtx pgx.Tx, accountID string) (string, error) {
	var currency string
	err := dbtx.QueryRow(ctx, `
		SELECT currency
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL
	`, accountID).Scan(&currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrAccountNotFound
		}
		return "", err
	}
	return currency, nil
}

// requireAccountOwnerForAccount checks if a user is the owner of an account.
func requireAccountOwnerForAccount(ctx context.Context, tx pgx.Tx, actorUserID string, accountID string) error {
	var one int
	err := tx.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1 AND ua.account_id = $2 AND ua.status = 'active' AND ua.permission = 'owner'
	`, actorUserID, accountID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrAccountForbidden
		}
		return err
	}
	return nil
}

// computeExchangeRate returns to_amount / from_amount rounded to 8 decimals.
// Returns nil when from_amount is zero (cannot compute).
func computeExchangeRate(fromAmount string, toAmount string) (*string, error) {
	fromRat, ok := new(big.Rat).SetString(strings.TrimSpace(fromAmount))
	if !ok {
		return nil, apperrors.ErrInvalidDecimalAmount
	}
	toRat, ok := new(big.Rat).SetString(strings.TrimSpace(toAmount))
	if !ok {
		return nil, apperrors.ErrInvalidDecimalAmount
	}
	if fromRat.Cmp(new(big.Rat)) == 0 {
		return nil, nil
	}
	rate := new(big.Rat).Quo(toRat, fromRat)
	v := rate.FloatString(8)
	return &v, nil
}

// encodeCursor creates a cursor string from occurred_at and id.
func encodeCursor(occurredAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", occurredAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// decodeCursor parses a cursor string back into occurred_at and id.
func decodeCursor(cursor *string) (*time.Time, *string, error) {
	if cursor == nil {
		return nil, nil, nil
	}
	c := strings.TrimSpace(*cursor)
	if c == "" {
		return nil, nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return nil, nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, nil, apperrors.ErrInvalidCursor
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, nil, err
	}
	id := parts[1]
	return &ts, &id, nil
}

// normalizeOptionalString trims whitespace from optional string pointers.
// Returns nil if the string is nil or empty after trimming.
func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// nullTimeToDatePtr converts sql.NullTime to a date string pointer (YYYY-MM-DD).
func nullTimeToDatePtr(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	v := nt.Time.UTC().Format("2006-01-02")
	return &v
}

// insertAuditEvent is a helper to insert an audit event within a transaction.
// It validates required fields and serializes diff to JSON.
func insertAuditEvent(ctx context.Context, dbtx pgx.Tx, accountID string, actorUserID string, action string, entityType string, entityID string, occurredAt time.Time, diff any) error {
	if accountID == "" || actorUserID == "" || action == "" || entityType == "" || entityID == "" {
		return nil
	}

	var diffJSON []byte
	if diff != nil {
		b, err := json.Marshal(diff)
		if err == nil {
			diffJSON = b
		}
	}

	id := uuid.NewString()

	_, err := dbtx.Exec(ctx, `
		INSERT INTO audit_events (id, account_id, actor_user_id, action, entity_type, entity_id, occurred_at, diff)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, id, accountID, actorUserID, action, entityType, entityID, occurredAt, diffJSON)
	return err
}
