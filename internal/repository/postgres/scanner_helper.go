package postgres

import (
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// RowScanner is an interface that matches both pgx.Row and pgx.Rows.
type RowScanner interface {
	Scan(dest ...any) error
}

// ScanAccount transforms a raw database row into an Account entity.
func ScanAccount(s RowScanner) (*entity.Account, error) {
	var a entity.Account
	var settingsJSON []byte
	err := s.Scan(
		&a.ID, &a.Name, &a.AccountNumber, &a.AccountType, &a.Currency,
		&a.ParentAccountID, &a.Status, &settingsJSON, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt,
		&a.DeletedAt, &a.Balance,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &a.Settings)
	}
	return &a, nil
}

// ScanTransaction transforms a raw database row into a Transaction entity.
func ScanTransaction(s RowScanner, includeCategoryIDs bool) (*entity.Transaction, error) {
	var t entity.Transaction
	dest := []any{
		&t.ID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate,
		&t.Amount, &t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID, &t.AccountName, &t.FromAccountID, &t.ToAccountID,
		&t.ExchangeRate, &t.AccountCurrency, &t.FromCurrency, &t.ToCurrency, &t.Status,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	}
	if includeCategoryIDs {
		dest = append(dest, &t.CategoryIDs)
	}

	err := s.Scan(dest...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ScanContact transforms a raw database row into a Contact entity.
func ScanContact(s RowScanner) (*entity.Contact, error) {
	var c entity.Contact
	err := s.Scan(
		&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.AvatarURL, &c.LinkedUserID, &c.Notes, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&c.LinkedDisplayName, &c.LinkedAvatarURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
