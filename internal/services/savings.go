package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type SavingsService interface {
	CreateInstrument(ctx context.Context, userID string, req CreateSavingsInstrumentRequest) (*domain.SavingsInstrument, error)
	GetInstrument(ctx context.Context, userID string, savingsInstrumentID string) (*domain.SavingsInstrument, error)
	ListInstruments(ctx context.Context, userID string) ([]domain.SavingsInstrument, error)
}

type CreateSavingsInstrumentRequest struct {
	SavingsAccountID string  `json:"savings_account_id"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        *bool   `json:"auto_renew,omitempty"`
	AccruedInterest  *string `json:"accrued_interest,omitempty"`
	Status           *string `json:"status,omitempty"`
}

type savingsService struct {
	accounts domain.AccountRepository
	repo     domain.SavingsRepository
}

func NewSavingsService(accounts domain.AccountRepository, repo domain.SavingsRepository) SavingsService {
	return &savingsService{accounts: accounts, repo: repo}
}

func (s *savingsService) CreateInstrument(ctx context.Context, userID string, req CreateSavingsInstrumentRequest) (*domain.SavingsInstrument, error) {
	savingsAccountID := strings.TrimSpace(req.SavingsAccountID)
	if savingsAccountID == "" {
		return nil, errors.New("savings_account_id is required")
	}

	principal := strings.TrimSpace(req.Principal)
	if principal == "" {
		return nil, errors.New("principal is required")
	}
	if !isValidDecimal(principal) {
		return nil, errors.New("principal must be a decimal string")
	}

	interestRate := normalizeOptionalString(req.InterestRate)
	if interestRate != nil && !isValidDecimal(*interestRate) {
		return nil, errors.New("interest_rate must be a decimal string")
	}

	accruedInterest := normalizeOptionalString(req.AccruedInterest)
	accrued := "0"
	if accruedInterest != nil {
		if !isValidDecimal(*accruedInterest) {
			return nil, errors.New("accrued_interest must be a decimal string")
		}
		accrued = *accruedInterest
	}

	if req.TermMonths != nil && *req.TermMonths < 0 {
		return nil, errors.New("term_months must be >= 0")
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
				return nil, errors.New("status is invalid")
			}
			status = v
		}
	}

	acc, err := s.accounts.GetAccountForUser(ctx, userID, savingsAccountID)
	if err != nil {
		return nil, err
	}
	if acc.AccountType != "savings" {
		return nil, errors.New("savings_account_id must be an account of type savings")
	}
	if acc.ParentAccountID == nil {
		return nil, errors.New("savings account must have parent_account_id")
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

	created, err := s.repo.GetSavingsInstrument(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func normalizeOptionalDate(s *string) (*string, error) {
	v := normalizeOptionalString(s)
	if v == nil {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", *v); err != nil {
		return nil, errors.New("date must be YYYY-MM-DD")
	}
	return v, nil
}

func (s *savingsService) GetInstrument(ctx context.Context, userID string, savingsInstrumentID string) (*domain.SavingsInstrument, error) {
	return s.repo.GetSavingsInstrument(ctx, userID, savingsInstrumentID)
}

func (s *savingsService) ListInstruments(ctx context.Context, userID string) ([]domain.SavingsInstrument, error) {
	return s.repo.ListSavingsInstruments(ctx, userID)
}
