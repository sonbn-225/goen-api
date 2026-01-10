package services

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type DebtService interface {
	Create(ctx context.Context, userID string, req CreateDebtRequest) (*domain.Debt, error)
	Get(ctx context.Context, userID string, debtID string) (*domain.Debt, error)
	List(ctx context.Context, userID string) ([]domain.Debt, error)

	CreatePayment(ctx context.Context, userID string, debtID string, req CreateDebtPaymentRequest) (*domain.DebtPaymentLink, error)
	ListPayments(ctx context.Context, userID string, debtID string) ([]domain.DebtPaymentLink, error)
	ListPaymentsByTransaction(ctx context.Context, userID string, transactionID string) ([]domain.DebtPaymentLink, error)

	CreateInstallment(ctx context.Context, userID string, debtID string, req CreateDebtInstallmentRequest) (*domain.DebtInstallment, error)
	ListInstallments(ctx context.Context, userID string, debtID string) ([]domain.DebtInstallment, error)
}

type CreateDebtRequest struct {
	ClientID     *string `json:"client_id,omitempty"`
	AccountID    string  `json:"account_id"`
	Direction    string  `json:"direction"`
	Name         *string `json:"name,omitempty"`
	Principal    string  `json:"principal"`
	StartDate    string  `json:"start_date"`
	DueDate      string  `json:"due_date"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type CreateDebtPaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
}

type CreateDebtInstallmentRequest struct {
	InstallmentNo int     `json:"installment_no"`
	DueDate       string  `json:"due_date"`
	AmountDue     string  `json:"amount_due"`
	AmountPaid    *string `json:"amount_paid,omitempty"`
	Status        *string `json:"status,omitempty"`
}

type debtService struct {
	txService TransactionService
	repo      domain.DebtRepository
}

func NewDebtService(txService TransactionService, repo domain.DebtRepository) DebtService {
	return &debtService{txService: txService, repo: repo}
}

func (s *debtService) Create(ctx context.Context, userID string, req CreateDebtRequest) (*domain.Debt, error) {
	direction := strings.TrimSpace(req.Direction)
	if direction != "borrowed" && direction != "lent" {
		return nil, ValidationError("direction is invalid", nil)
	}

	principal := strings.TrimSpace(req.Principal)
	if principal == "" {
		return nil, ValidationError("principal is required", nil)
	}
	if !isValidDecimal(principal) {
		return nil, ValidationError("principal must be a decimal string", nil)
	}

	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, ValidationError("account_id is required", map[string]any{"field": "account_id"})
	}

	startDate := strings.TrimSpace(req.StartDate)
	if startDate == "" {
		return nil, ValidationError("start_date is required", nil)
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, ValidationError("start_date must be YYYY-MM-DD", nil)
	}

	dueDate := strings.TrimSpace(req.DueDate)
	if dueDate == "" {
		return nil, ValidationError("due_date is required", nil)
	}
	dueT, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return nil, ValidationError("due_date must be YYYY-MM-DD", nil)
	}
	startT, _ := time.Parse("2006-01-02", startDate)
	if dueT.Before(startT) {
		return nil, ValidationError("due_date must be >= start_date", nil)
	}

	var interestRate *string
	if req.InterestRate != nil {
		v := strings.TrimSpace(*req.InterestRate)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, ValidationError("interest_rate must be a decimal string", nil)
			}
			interestRate = &v
		}
	}

	var interestRule *string
	if req.InterestRule != nil {
		v := strings.TrimSpace(*req.InterestRule)
		if v != "" {
			if v != "interest_first" && v != "principal_first" {
				return nil, ValidationError("interest_rule is invalid", nil)
			}
			interestRule = &v
		}
	}

	status := "active"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "overdue" && v != "closed" {
				return nil, ValidationError("status is invalid", nil)
			}
			status = v
		}
	}

	now := time.Now().UTC()
	name := normalizeOptionalString(req.Name)
	clientID := normalizeOptionalString(req.ClientID)

	debt := domain.Debt{
		ID:                   uuid.NewString(),
		ClientID:             clientID,
		UserID:               userID,
		AccountID:            &accountID,
		Direction:            direction,
		Name:                 name,
		Principal:            principal,
		StartDate:            startDate,
		DueDate:              dueDate,
		InterestRate:         interestRate,
		InterestRule:         interestRule,
		OutstandingPrincipal: principal,
		AccruedInterest:      "0",
		Status:               status,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.repo.CreateDebt(ctx, debt); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, NotFoundErrorWithCause("account not found", map[string]any{"field": "account_id"}, err)
		}
		if errors.Is(err, domain.ErrAccountForbidden) {
			return nil, ForbiddenErrorWithCause("account forbidden", map[string]any{"field": "account_id"}, err)
		}
		return nil, err
	}
	return s.repo.GetDebt(ctx, userID, debt.ID)
}

func (s *debtService) Get(ctx context.Context, userID string, debtID string) (*domain.Debt, error) {
	item, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}
	return item, nil
}

func (s *debtService) List(ctx context.Context, userID string) ([]domain.Debt, error) {
	return s.repo.ListDebts(ctx, userID)
}

func (s *debtService) CreatePayment(ctx context.Context, userID string, debtID string, req CreateDebtPaymentRequest) (*domain.DebtPaymentLink, error) {
	transactionID := strings.TrimSpace(req.TransactionID)
	if transactionID == "" {
		return nil, ValidationError("transaction_id is required", nil)
	}

	debt, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}

	tx, err := s.txService.Get(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}

	// Direction mapping (recommended):
	// - borrowed: repayment is expense
	// - lent: collection is income
	if debt.Direction == "borrowed" && tx.Type != "expense" {
		return nil, ValidationError("transaction.type must be expense for borrowed debt payment", nil)
	}
	if debt.Direction == "lent" && tx.Type != "income" {
		return nil, ValidationError("transaction.type must be income for lent debt collection", nil)
	}

	amountRat, ok := new(big.Rat).SetString(tx.Amount)
	if !ok {
		return nil, ValidationError("transaction amount is invalid", nil)
	}

	outstandingRat, ok := new(big.Rat).SetString(debt.OutstandingPrincipal)
	if !ok {
		return nil, ValidationError("outstanding_principal is invalid", nil)
	}

	accruedRat, ok := new(big.Rat).SetString(debt.AccruedInterest)
	if !ok {
		return nil, ValidationError("accrued_interest is invalid", nil)
	}

	var principalPaidRat *big.Rat
	var interestPaidRat *big.Rat

	if req.PrincipalPaid != nil || req.InterestPaid != nil {
		if req.PrincipalPaid != nil {
			v := strings.TrimSpace(*req.PrincipalPaid)
			if v != "" {
				if !isValidDecimal(v) {
					return nil, ValidationError("principal_paid must be a decimal string", nil)
				}
				p, ok := new(big.Rat).SetString(v)
				if !ok {
					return nil, ValidationError("principal_paid must be a decimal string", nil)
				}
				principalPaidRat = p
			}
		}
		if req.InterestPaid != nil {
			v := strings.TrimSpace(*req.InterestPaid)
			if v != "" {
				if !isValidDecimal(v) {
					return nil, ValidationError("interest_paid must be a decimal string", nil)
				}
				i, ok := new(big.Rat).SetString(v)
				if !ok {
					return nil, ValidationError("interest_paid must be a decimal string", nil)
				}
				interestPaidRat = i
			}
		}

		if principalPaidRat == nil {
			principalPaidRat = big.NewRat(0, 1)
		}
		if interestPaidRat == nil {
			interestPaidRat = big.NewRat(0, 1)
		}

		sum := new(big.Rat).Add(principalPaidRat, interestPaidRat)
		if sum.Cmp(amountRat) != 0 {
			return nil, ValidationError("principal_paid + interest_paid must equal transaction.amount", nil)
		}
	} else {
		// Auto allocate based on interest_rule
		interestFirst := true
		if debt.InterestRule != nil && *debt.InterestRule == "principal_first" {
			interestFirst = false
		}

		remaining := new(big.Rat).Set(amountRat)
		principalPaidRat = big.NewRat(0, 1)
		interestPaidRat = big.NewRat(0, 1)

		if interestFirst {
			interestPaidRat = minRat(remaining, accruedRat)
			remaining.Sub(remaining, interestPaidRat)
			principalPaidRat = minRat(remaining, outstandingRat)
		} else {
			principalPaidRat = minRat(remaining, outstandingRat)
			remaining.Sub(remaining, principalPaidRat)
			interestPaidRat = minRat(remaining, accruedRat)
		}

		// If payment exceeds total due, reject
		totalDue := new(big.Rat).Add(outstandingRat, accruedRat)
		if amountRat.Cmp(totalDue) > 0 {
			return nil, ValidationError("payment exceeds total due", nil)
		}
	}

	newOutstanding := new(big.Rat).Sub(outstandingRat, principalPaidRat)
	newAccrued := new(big.Rat).Sub(accruedRat, interestPaidRat)
	if newOutstanding.Sign() < 0 || newAccrued.Sign() < 0 {
		return nil, ValidationError("outstanding cannot become negative", nil)
	}

	status := debt.Status
	var closedAt *time.Time
	if newOutstanding.Sign() == 0 && newAccrued.Sign() == 0 {
		status = "closed"
		t := time.Now().UTC()
		closedAt = &t
	}

	pStr := ratToDecimalString(principalPaidRat)
	iStr := ratToDecimalString(interestPaidRat)

	link := domain.DebtPaymentLink{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		TransactionID: transactionID,
		PrincipalPaid: &pStr,
		InterestPaid:  &iStr,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreatePaymentLink(ctx, userID, link, ratToDecimalString(newOutstanding), ratToDecimalString(newAccrued), status, closedAt); err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}
	return &link, nil
}

func (s *debtService) ListPayments(ctx context.Context, userID string, debtID string) ([]domain.DebtPaymentLink, error) {
	items, err := s.repo.ListPaymentLinks(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}
	return items, nil
}

func (s *debtService) ListPaymentsByTransaction(ctx context.Context, userID string, transactionID string) ([]domain.DebtPaymentLink, error) {
	transactionID = strings.TrimSpace(transactionID)
	if transactionID == "" {
		return nil, ValidationError("transaction_id is required", nil)
	}

	// Ensure the user has access to this transaction.
	if _, err := s.txService.Get(ctx, userID, transactionID); err != nil {
		return nil, err
	}

	return s.repo.ListPaymentLinksByTransaction(ctx, userID, transactionID)
}

func (s *debtService) CreateInstallment(ctx context.Context, userID string, debtID string, req CreateDebtInstallmentRequest) (*domain.DebtInstallment, error) {
	if req.InstallmentNo <= 0 {
		return nil, ValidationError("installment_no must be > 0", nil)
	}
	dueDate := strings.TrimSpace(req.DueDate)
	if dueDate == "" {
		return nil, ValidationError("due_date is required", nil)
	}
	if _, err := time.Parse("2006-01-02", dueDate); err != nil {
		return nil, ValidationError("due_date must be YYYY-MM-DD", nil)
	}

	amountDue := strings.TrimSpace(req.AmountDue)
	if amountDue == "" {
		return nil, ValidationError("amount_due is required", nil)
	}
	if !isValidDecimal(amountDue) {
		return nil, ValidationError("amount_due must be a decimal string", nil)
	}

	amountPaid := "0"
	if req.AmountPaid != nil {
		v := strings.TrimSpace(*req.AmountPaid)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, ValidationError("amount_paid must be a decimal string", nil)
			}
			amountPaid = v
		}
	}

	status := "pending"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "pending" && v != "paid" && v != "overdue" {
				return nil, ValidationError("status is invalid", nil)
			}
			status = v
		}
	}

	inst := domain.DebtInstallment{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		InstallmentNo: req.InstallmentNo,
		DueDate:       dueDate,
		AmountDue:     amountDue,
		AmountPaid:    amountPaid,
		Status:        status,
	}

	if err := s.repo.CreateInstallment(ctx, userID, inst); err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}
	return &inst, nil
}

func (s *debtService) ListInstallments(ctx context.Context, userID string, debtID string) ([]domain.DebtInstallment, error) {
	items, err := s.repo.ListInstallments(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, domain.ErrDebtNotFound) {
			return nil, NotFoundErrorWithCause("debt not found", nil, err)
		}
		return nil, err
	}
	return items, nil
}

func minRat(a *big.Rat, b *big.Rat) *big.Rat {
	if a.Cmp(b) <= 0 {
		return new(big.Rat).Set(a)
	}
	return new(big.Rat).Set(b)
}

func ratToDecimalString(r *big.Rat) string {
	// Use 2 decimal places to match numeric(18,2)
	return r.FloatString(2)
}
