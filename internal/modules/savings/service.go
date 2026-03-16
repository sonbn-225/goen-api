package savings

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/i18n"
	"github.com/sonbn-225/goen-api/internal/modules/account"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
)

// CreateRequest contains savings instrument create parameters.
type CreateRequest struct {
	Name             *string `json:"name,omitempty"`
	SavingsAccountID *string `json:"savings_account_id,omitempty"`
	ParentAccountID  *string `json:"parent_account_id,omitempty"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        *bool   `json:"auto_renew,omitempty"`
	AccruedInterest  *string `json:"accrued_interest,omitempty"`
	Status           *string `json:"status,omitempty"`
}

// PatchRequest contains savings instrument patch parameters.
type PatchRequest struct {
	Principal       *string `json:"principal,omitempty"`
	InterestRate    *string `json:"interest_rate,omitempty"`
	TermMonths      *int    `json:"term_months,omitempty"`
	StartDate       *string `json:"start_date,omitempty"`
	MaturityDate    *string `json:"maturity_date,omitempty"`
	AutoRenew       *bool   `json:"auto_renew,omitempty"`
	AccruedInterest *string `json:"accrued_interest,omitempty"`
	Status          *string `json:"status,omitempty"`
}

// Service handles savings business logic.
type Service struct {
	accounts AccountServiceInterface
	tx       TransactionServiceInterface
	repo     domain.SavingsRepository
}

// NewService creates a new savings service.
func NewService(accounts AccountServiceInterface, tx TransactionServiceInterface, repo domain.SavingsRepository) *Service {
	return &Service{accounts: accounts, tx: tx, repo: repo}
}

// Create creates a new savings instrument.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*domain.SavingsInstrument, error) {
	var savingsAccountID string
	parentAccountID := normalizeOptionalString(req.ParentAccountID)
	name := normalizeOptionalString(req.Name)
	if req.SavingsAccountID != nil {
		savingsAccountID = strings.TrimSpace(*req.SavingsAccountID)
	}
	if savingsAccountID == "" && parentAccountID == nil {
		return nil, apperrors.Validation("either savings_account_id or parent_account_id is required", nil)
	}

	principal := strings.TrimSpace(req.Principal)
	if principal == "" {
		return nil, apperrors.Validation("principal is required", map[string]any{"field": "principal"})
	}
	if !isValidDecimal(principal) {
		return nil, apperrors.Validation("principal must be a decimal string", map[string]any{"field": "principal"})
	}

	interestRate := normalizeOptionalString(req.InterestRate)
	if interestRate != nil && !isValidDecimal(*interestRate) {
		return nil, apperrors.Validation("interest_rate must be a decimal string", map[string]any{"field": "interest_rate"})
	}

	accruedInterest := normalizeOptionalString(req.AccruedInterest)
	accrued := "0"
	if accruedInterest != nil {
		if !isValidDecimal(*accruedInterest) {
			return nil, apperrors.Validation("accrued_interest must be a decimal string", map[string]any{"field": "accrued_interest"})
		}
		accrued = *accruedInterest
	}

	if req.TermMonths != nil && *req.TermMonths < 0 {
		return nil, apperrors.Validation("term_months must be >= 0", map[string]any{"field": "term_months"})
	}

	startDate, err := normalizeOptionalDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	maturityDate, err := normalizeOptionalDate(req.MaturityDate)
	if err != nil {
		return nil, err
	}

	status := "active"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "matured" && v != "closed" {
				return nil, apperrors.Validation("status is invalid", map[string]any{"field": "status"})
			}
			status = v
		}
	}

	var acc *domain.Account
	autoCreatedAccount := false
	lang := httpx.LangFromContext(ctx)
	if savingsAccountID == "" {
		parent, err := s.accounts.Get(ctx, userID, *parentAccountID)
		if err != nil {
			return nil, err
		}
		if parent.AccountType != "bank" && parent.AccountType != "wallet" {
			return nil, apperrors.Validation("parent account must be bank or wallet", map[string]any{"field": "parent_account_id"})
		}

		accName := i18n.T(lang, "savings_name")
		if name != nil {
			accName = *name
		}

		createReq := account.CreateAccountRequest{
			Name:            accName,
			AccountType:     "savings",
			Currency:        parent.Currency,
			ParentAccountID: parentAccountID,
		}
		createdAcc, err := s.accounts.Create(ctx, userID, createReq)
		if err != nil {
			return nil, err
		}
		savingsAccountID = createdAcc.ID
		acc = createdAcc
		autoCreatedAccount = true
	} else {
		cur, err := s.accounts.Get(ctx, userID, savingsAccountID)
		if err != nil {
			return nil, err
		}
		acc = cur
	}

	if acc.AccountType != "savings" {
		return nil, apperrors.Validation("savings_account_id must be an account of type savings", map[string]any{"field": "savings_account_id"})
	}
	if acc.ParentAccountID == nil {
		return nil, apperrors.Validation("savings account must have parent_account_id", map[string]any{"field": "savings_account_id"})
	}

	autoRenew := false
	if req.AutoRenew != nil {
		autoRenew = *req.AutoRenew
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	item := domain.SavingsInstrument{
		ID:               id,
		SavingsAccountID: savingsAccountID,
		ParentAccountID:  *acc.ParentAccountID,
		Principal:        principal,
		InterestRate:     interestRate,
		TermMonths:       req.TermMonths,
		StartDate:        startDate,
		MaturityDate:     maturityDate,
		AutoRenew:        autoRenew,
		AccruedInterest:  accrued,
		Status:           status,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateSavingsInstrument(ctx, userID, item); err != nil {
		return nil, err
	}

	if autoCreatedAccount {
		if amt, ok := new(big.Rat).SetString(principal); ok && amt.Cmp(new(big.Rat)) > 0 {
			occurredDate := now.Format("2006-01-02")
			if startDate != nil && strings.TrimSpace(*startDate) != "" {
				occurredDate = strings.TrimSpace(*startDate)
			}

			desc := i18n.T(lang, "savings_deposit")
			if acc != nil {
				name := strings.TrimSpace(acc.Name)
				if name != "" {
					desc = i18n.T(lang, "savings_deposit") + ": " + name
				}
			}

			fromID := *acc.ParentAccountID
			toID := acc.ID
			txReq := transaction.CreateRequest{
				Type:          "transfer",
				OccurredDate:  &occurredDate,
				Amount:        principal,
				Description:   &desc,
				FromAccountID: &fromID,
				ToAccountID:   &toID,
			}
			if _, err := s.tx.Create(ctx, userID, txReq); err != nil {
				_ = s.repo.DeleteSavingsInstrument(ctx, userID, id)
				_ = s.accounts.Delete(ctx, userID, acc.ID)
				return nil, err
			}
		}
	}

	created, err := s.repo.GetSavingsInstrument(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Get retrieves a savings instrument by ID.
func (s *Service) Get(ctx context.Context, userID, savingsInstrumentID string) (*domain.SavingsInstrument, error) {
	item, err := s.repo.GetSavingsInstrument(ctx, userID, savingsInstrumentID)
	if err != nil {
		if strings.TrimSpace(savingsInstrumentID) == "" {
			return nil, apperrors.Validation("instrumentId is required", map[string]any{"field": "instrumentId"})
		}
		if errors.Is(err, apperrors.ErrSavingsInstrumentNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "savings instrument not found", err)
		}
		return nil, err
	}
	return item, nil
}

// List returns all savings instruments for a user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.SavingsInstrument, error) {
	return s.repo.ListSavingsInstruments(ctx, userID)
}

// Patch updates a savings instrument.
func (s *Service) Patch(ctx context.Context, userID, savingsInstrumentID string, req PatchRequest) (*domain.SavingsInstrument, error) {
	cur, err := s.repo.GetSavingsInstrument(ctx, userID, savingsInstrumentID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSavingsInstrumentNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "savings instrument not found", err)
		}
		return nil, err
	}

	if req.Principal != nil {
		principal := strings.TrimSpace(*req.Principal)
		if principal == "" {
			return nil, apperrors.Validation("principal is required", map[string]any{"field": "principal"})
		}
		if !isValidDecimal(principal) {
			return nil, apperrors.Validation("principal must be a decimal string", map[string]any{"field": "principal"})
		}
		cur.Principal = principal
	}

	if req.InterestRate != nil {
		v := strings.TrimSpace(*req.InterestRate)
		if v == "" {
			cur.InterestRate = nil
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("interest_rate must be a decimal string", map[string]any{"field": "interest_rate"})
			}
			cur.InterestRate = &v
		}
	}

	if req.AccruedInterest != nil {
		v := strings.TrimSpace(*req.AccruedInterest)
		if v == "" {
			cur.AccruedInterest = "0"
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("accrued_interest must be a decimal string", map[string]any{"field": "accrued_interest"})
			}
			cur.AccruedInterest = v
		}
	}

	if req.TermMonths != nil {
		if *req.TermMonths < 0 {
			return nil, apperrors.Validation("term_months must be >= 0", map[string]any{"field": "term_months"})
		}
		if *req.TermMonths == 0 {
			cur.TermMonths = nil
		} else {
			v := *req.TermMonths
			cur.TermMonths = &v
		}
	}

	if req.StartDate != nil {
		v := strings.TrimSpace(*req.StartDate)
		if v == "" {
			cur.StartDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return nil, apperrors.Validation("date must be YYYY-MM-DD", map[string]any{"field": "start_date"})
			}
			cur.StartDate = &v
		}
	}

	if req.MaturityDate != nil {
		v := strings.TrimSpace(*req.MaturityDate)
		if v == "" {
			cur.MaturityDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return nil, apperrors.Validation("date must be YYYY-MM-DD", map[string]any{"field": "maturity_date"})
			}
			cur.MaturityDate = &v
		}
	}

	if req.AutoRenew != nil {
		cur.AutoRenew = *req.AutoRenew
	}

	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "matured" && v != "closed" {
				return nil, apperrors.Validation("status is invalid", map[string]any{"field": "status"})
			}
			cur.Status = v
			if v == "closed" {
				now := time.Now().UTC()
				cur.ClosedAt = &now
			} else {
				cur.ClosedAt = nil
			}
		}
	}

	cur.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateSavingsInstrument(ctx, userID, *cur); err != nil {
		if errors.Is(err, apperrors.ErrSavingsInstrumentNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "savings instrument not found", err)
		}
		return nil, err
	}

	item, err := s.repo.GetSavingsInstrument(ctx, userID, savingsInstrumentID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSavingsInstrumentNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "savings instrument not found", err)
		}
		return nil, err
	}
	return item, nil
}

// Delete deletes a savings instrument.
func (s *Service) Delete(ctx context.Context, userID, savingsInstrumentID string) error {
	err := s.repo.DeleteSavingsInstrument(ctx, userID, savingsInstrumentID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSavingsInstrumentNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "savings instrument not found", err)
		}
		return err
	}
	return nil
}

func normalizeOptionalDate(s *string) (*string, error) {
	v := normalizeOptionalString(s)
	if v == nil {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", *v); err != nil {
		return nil, apperrors.Validation("date must be YYYY-MM-DD", nil)
	}
	return v, nil
}

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

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}

