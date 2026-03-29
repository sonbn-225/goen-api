package debt

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/contact"
)

// CreateRequest contains debt create parameters.
type CreateRequest struct {
	ClientID     *string `json:"client_id,omitempty"`
	AccountID    string  `json:"account_id"`
	Direction    string  `json:"direction"`
	Name         *string `json:"name,omitempty"`
	ContactID    *string `json:"contact_id,omitempty"`
	Principal    string  `json:"principal"`
	StartDate    string  `json:"start_date"`
	DueDate      string  `json:"due_date"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
	Status       *string `json:"status,omitempty"`
}

// CreatePaymentRequest contains debt payment create parameters.
type CreatePaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
}

// CreateInstallmentRequest contains debt installment create parameters.
type CreateInstallmentRequest struct {
	InstallmentNo int     `json:"installment_no"`
	DueDate       string  `json:"due_date"`
	AmountDue     string  `json:"amount_due"`
	AmountPaid    *string `json:"amount_paid,omitempty"`
	Status        *string `json:"status,omitempty"`
}

// Service handles debt business logic.
type Service struct {
	txSvc      TransactionServiceInterface
	contactSvc *contact.Service
	repo       domain.DebtRepository
}

// NewService creates a new debt service.
func NewService(txSvc TransactionServiceInterface, repo domain.DebtRepository, contactSvc *contact.Service) *Service {
	return &Service{txSvc: txSvc, repo: repo, contactSvc: contactSvc}
}

// Create creates a new debt.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*domain.Debt, error) {
	direction := strings.TrimSpace(req.Direction)
	if direction != "borrowed" && direction != "lent" {
		return nil, apperrors.Validation("direction is invalid", nil)
	}

	principal := strings.TrimSpace(req.Principal)
	if principal == "" {
		return nil, apperrors.Validation("principal is required", nil)
	}
	if !isValidDecimal(principal) {
		return nil, apperrors.Validation("principal must be a decimal string", nil)
	}

	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, apperrors.Validation("account_id is required", map[string]any{"field": "account_id"})
	}

	startDate := strings.TrimSpace(req.StartDate)
	if startDate == "" {
		return nil, apperrors.Validation("start_date is required", nil)
	}
	startT, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, apperrors.Validation("start_date must be YYYY-MM-DD", nil)
	}

	dueDate := strings.TrimSpace(req.DueDate)
	if dueDate == "" {
		return nil, apperrors.Validation("due_date is required", nil)
	}
	dueT, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return nil, apperrors.Validation("due_date must be YYYY-MM-DD", nil)
	}
	if dueT.Before(startT) {
		return nil, apperrors.Validation("due_date must be >= start_date", nil)
	}

	var interestRate *string
	if req.InterestRate != nil {
		v := strings.TrimSpace(*req.InterestRate)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("interest_rate must be a decimal string", nil)
			}
			interestRate = &v
		}
	}

	var interestRule *string
	if req.InterestRule != nil {
		v := strings.TrimSpace(*req.InterestRule)
		if v != "" {
			if v != "interest_first" && v != "principal_first" {
				return nil, apperrors.Validation("interest_rule is invalid", nil)
			}
			interestRule = &v
		}
	}

	status := "active"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "overdue" && v != "closed" {
				return nil, apperrors.Validation("status is invalid", nil)
			}
			status = v
		}
	}

	now := time.Now().UTC()
	name := normalizeOptionalString(req.Name)
	contactID := normalizeOptionalString(req.ContactID)

	// Auto-create contact if name exists but contactID is missing
	if contactID == nil && name != nil {
		// 1. Try to find existing contact by name for this user
		contacts, err := s.contactSvc.List(ctx, userID)
		if err == nil {
			for _, c := range contacts {
				if strings.EqualFold(c.Name, *name) {
					id := c.ID
					contactID = &id
					break
				}
			}
		}

		// 2. If still not found, create new contact
		if contactID == nil {
			newContact, err := s.contactSvc.Create(ctx, userID, contact.CreateRequest{
				Name: *name,
			})
			if err == nil {
				contactID = &newContact.ID
			} else {
				// Log error but continue with raw name if needed? 
				// Actually, we should probably return error if we want strict contact linking.
				// But user said "hệ thống sẽ tạo danh bạ tương ứng".
				return nil, err
			}
		}
	}

	clientID := normalizeOptionalString(req.ClientID)

	debt := domain.Debt{
		ID:                   uuid.NewString(),
		ClientID:             clientID,
		UserID:               userID,
		AccountID:            &accountID,
		Direction:            direction,
		Name:                 name,
		ContactID:            contactID,
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
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "account forbidden", err)
		}
		return nil, err
	}
	return s.repo.GetDebt(ctx, userID, debt.ID)
}

// Get retrieves a debt by ID.
func (s *Service) Get(ctx context.Context, userID, debtID string) (*domain.Debt, error) {
	item, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}
	return item, nil
}

// List returns all debts for a user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Debt, error) {
	return s.repo.ListDebts(ctx, userID)
}

// CreatePayment links a transaction to a debt as payment.
func (s *Service) CreatePayment(ctx context.Context, userID, debtID string, req CreatePaymentRequest) (*domain.DebtPaymentLink, error) {
	transactionID := strings.TrimSpace(req.TransactionID)
	if transactionID == "" {
		return nil, apperrors.Validation("transaction_id is required", nil)
	}

	debt, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}

	tx, err := s.txSvc.Get(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}

	amountRat, ok := new(big.Rat).SetString(tx.Amount)
	if !ok {
		return nil, apperrors.Validation("transaction amount is invalid", nil)
	}

	principalRat, ok := new(big.Rat).SetString(debt.Principal)
	if !ok {
		return nil, apperrors.Validation("principal is invalid", nil)
	}

	outstandingRat, ok := new(big.Rat).SetString(debt.OutstandingPrincipal)
	if !ok {
		return nil, apperrors.Validation("outstanding_principal is invalid", nil)
	}

	accruedRat, ok := new(big.Rat).SetString(debt.AccruedInterest)
	if !ok {
		return nil, apperrors.Validation("accrued_interest is invalid", nil)
	}

	// Determine if this is a repayment or a top-up
	isTopUp := false
	if debt.Direction == "borrowed" && tx.Type == "income" {
		isTopUp = true
	} else if debt.Direction == "lent" && tx.Type == "expense" {
		isTopUp = true
	}

	var principalPaidRat *big.Rat
	var interestPaidRat *big.Rat
	var newPrincipal, newOutstanding, newAccrued *big.Rat
	status := debt.Status

	if isTopUp {
		// Logic for borrowing/lending more
		newPrincipal = new(big.Rat).Add(principalRat, amountRat)
		newOutstanding = new(big.Rat).Add(outstandingRat, amountRat)
		newAccrued = accruedRat
		principalPaidRat = big.NewRat(0, 1)
		interestPaidRat = big.NewRat(0, 1)
	} else {
		// Logic for repayment (original logic)
		if debt.Direction == "borrowed" && tx.Type != "expense" {
			return nil, apperrors.Validation("transaction.type must be expense for borrowed debt payment", nil)
		}
		if debt.Direction == "lent" && tx.Type != "income" {
			return nil, apperrors.Validation("transaction.type must be income for lent debt collection", nil)
		}

		if req.PrincipalPaid != nil || req.InterestPaid != nil {
			if req.PrincipalPaid != nil {
				v := strings.TrimSpace(*req.PrincipalPaid)
				if v != "" {
					if !isValidDecimal(v) {
						return nil, apperrors.Validation("principal_paid must be a decimal string", nil)
					}
					p, ok := new(big.Rat).SetString(v)
					if !ok {
						return nil, apperrors.Validation("principal_paid must be a decimal string", nil)
					}
					principalPaidRat = p
				}
			}
			if req.InterestPaid != nil {
				v := strings.TrimSpace(*req.InterestPaid)
				if v != "" {
					if !isValidDecimal(v) {
						return nil, apperrors.Validation("interest_paid must be a decimal string", nil)
					}
					i, ok := new(big.Rat).SetString(v)
					if !ok {
						return nil, apperrors.Validation("interest_paid must be a decimal string", nil)
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
				return nil, apperrors.Validation("principal_paid + interest_paid must equal transaction.amount", nil)
			}
		} else {
			// Auto-allocate logic
			interestFirst := true
			if debt.InterestRule != nil && *debt.InterestRule == "principal_first" {
				interestFirst = false
			}

			remaining := new(big.Rat).Set(amountRat)
			var principalPaidLocal, interestPaidLocal *big.Rat

			if interestFirst {
				interestPaidLocal = minRat(remaining, accruedRat)
				remaining.Sub(remaining, interestPaidLocal)
				principalPaidLocal = minRat(remaining, outstandingRat)
			} else {
				principalPaidLocal = minRat(remaining, outstandingRat)
				remaining.Sub(remaining, principalPaidLocal)
				interestPaidLocal = minRat(remaining, accruedRat)
			}
			principalPaidRat = principalPaidLocal
			interestPaidRat = interestPaidLocal

			totalDue := new(big.Rat).Add(outstandingRat, accruedRat)
			if amountRat.Cmp(totalDue) > 0 {
				return nil, apperrors.Validation("payment exceeds total due", nil)
			}
		}

		newPrincipal = principalRat
		newOutstanding = new(big.Rat).Sub(outstandingRat, principalPaidRat)
		newAccrued = new(big.Rat).Sub(accruedRat, interestPaidRat)

		if newOutstanding.Sign() < 0 || newAccrued.Sign() < 0 {
			return nil, apperrors.Validation("outstanding cannot become negative", nil)
		}

		if newOutstanding.Sign() == 0 && newAccrued.Sign() == 0 {
			status = "closed"
		}
	}

	pStr := ratToDecimalString(principalPaidRat)
	iStr := ratToDecimalString(interestPaidRat)

	now := time.Now().UTC()
	var closedAt *time.Time
	if status == "closed" {
		if debt.Status == "closed" {
			closedAt = debt.ClosedAt
		} else {
			closedAt = &now
		}
	}

	link := domain.DebtPaymentLink{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		TransactionID: transactionID,
		PrincipalPaid: &pStr,
		InterestPaid:  &iStr,
		CreatedAt:     now,
	}

	if err := s.repo.CreatePaymentLink(ctx, userID, link, ratToDecimalString(newPrincipal), ratToDecimalString(newOutstanding), ratToDecimalString(newAccrued), status, closedAt); err != nil {
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}
	return &link, nil
}

// ListPayments returns all payment links for a debt.
func (s *Service) ListPayments(ctx context.Context, userID, debtID string) ([]domain.DebtPaymentLink, error) {
	items, err := s.repo.ListPaymentLinks(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}
	return items, nil
}

// ListPaymentsByTransaction returns payment links for a transaction.
func (s *Service) ListPaymentsByTransaction(ctx context.Context, userID, transactionID string) ([]domain.DebtPaymentLink, error) {
	transactionID = strings.TrimSpace(transactionID)
	if transactionID == "" {
		return nil, apperrors.Validation("transaction_id is required", nil)
	}

	if _, err := s.txSvc.Get(ctx, userID, transactionID); err != nil {
		return nil, err
	}

	return s.repo.ListPaymentLinksByTransaction(ctx, userID, transactionID)
}

// CreateInstallment creates a debt installment.
func (s *Service) CreateInstallment(ctx context.Context, userID, debtID string, req CreateInstallmentRequest) (*domain.DebtInstallment, error) {
	if req.InstallmentNo <= 0 {
		return nil, apperrors.Validation("installment_no must be > 0", nil)
	}
	dueDate := strings.TrimSpace(req.DueDate)
	if dueDate == "" {
		return nil, apperrors.Validation("due_date is required", nil)
	}
	if _, err := time.Parse("2006-01-02", dueDate); err != nil {
		return nil, apperrors.Validation("due_date must be YYYY-MM-DD", nil)
	}

	amountDue := strings.TrimSpace(req.AmountDue)
	if amountDue == "" {
		return nil, apperrors.Validation("amount_due is required", nil)
	}
	if !isValidDecimal(amountDue) {
		return nil, apperrors.Validation("amount_due must be a decimal string", nil)
	}

	amountPaid := "0"
	if req.AmountPaid != nil {
		v := strings.TrimSpace(*req.AmountPaid)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("amount_paid must be a decimal string", nil)
			}
			amountPaid = v
		}
	}

	status := "pending"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "pending" && v != "paid" && v != "overdue" {
				return nil, apperrors.Validation("status is invalid", nil)
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
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}
	return &inst, nil
}

// ListInstallments returns all installments for a debt.
func (s *Service) ListInstallments(ctx context.Context, userID, debtID string) ([]domain.DebtInstallment, error) {
	items, err := s.repo.ListInstallments(ctx, userID, debtID)
	if err != nil {
		if errors.Is(err, apperrors.ErrDebtNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "debt not found", err)
		}
		return nil, err
	}
	return items, nil
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

func minRat(a *big.Rat, b *big.Rat) *big.Rat {
	if a.Cmp(b) <= 0 {
		return new(big.Rat).Set(a)
	}
	return new(big.Rat).Set(b)
}

func ratToDecimalString(r *big.Rat) string {
	return r.FloatString(2)
}

