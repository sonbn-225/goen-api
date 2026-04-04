package debt

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/contact"
)

type service struct {
	repo           Repository
	txService      TransactionService
	contactService ContactService
}

var _ Service = (*service)(nil)

func NewService(repo Repository, txService TransactionService, contactService ContactService) Service {
	return &service{repo: repo, txService: txService, contactService: contactService}
}

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*Debt, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "create")
	logger.Info("debt_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if s.repo == nil {
		return nil, apperrors.New(apperrors.KindInternal, "debt repository not configured")
	}

	direction := strings.TrimSpace(input.Direction)
	if direction != "borrowed" && direction != "lent" {
		return nil, apperrors.New(apperrors.KindValidation, "direction is invalid")
	}

	principal := strings.TrimSpace(input.Principal)
	if principal == "" {
		return nil, apperrors.New(apperrors.KindValidation, "principal is required")
	}
	if !isValidDecimal(principal) {
		return nil, apperrors.New(apperrors.KindValidation, "principal must be a decimal string")
	}

	accountID := strings.TrimSpace(input.AccountID)
	if accountID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "account_id is required")
	}

	startDate := strings.TrimSpace(input.StartDate)
	if startDate == "" {
		return nil, apperrors.New(apperrors.KindValidation, "start_date is required")
	}
	startT, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "start_date must be YYYY-MM-DD")
	}

	dueDate := strings.TrimSpace(input.DueDate)
	if dueDate == "" {
		return nil, apperrors.New(apperrors.KindValidation, "due_date is required")
	}
	dueT, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "due_date must be YYYY-MM-DD")
	}
	if dueT.Before(startT) {
		return nil, apperrors.New(apperrors.KindValidation, "due_date must be >= start_date")
	}

	var interestRate *string
	if input.InterestRate != nil {
		v := strings.TrimSpace(*input.InterestRate)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, apperrors.New(apperrors.KindValidation, "interest_rate must be a decimal string")
			}
			interestRate = &v
		}
	}

	var interestRule *string
	if input.InterestRule != nil {
		v := strings.TrimSpace(*input.InterestRule)
		if v != "" {
			if v != "interest_first" && v != "principal_first" {
				return nil, apperrors.New(apperrors.KindValidation, "interest_rule is invalid")
			}
			interestRule = &v
		}
	}

	status := "active"
	if input.Status != nil {
		v := strings.TrimSpace(*input.Status)
		if v != "" {
			if v != "active" && v != "overdue" && v != "closed" {
				return nil, apperrors.New(apperrors.KindValidation, "status is invalid")
			}
			status = v
		}
	}

	now := time.Now().UTC()
	name := normalizeOptionalString(input.Name)
	contactID := normalizeOptionalString(input.ContactID)

	if contactID == nil && name != nil {
		if s.contactService == nil {
			return nil, apperrors.New(apperrors.KindInternal, "contact service not configured")
		}

		contacts, err := s.contactService.List(ctx, userID)
		if err != nil {
			logger.Error("debt_create_failed", "error", err)
			return nil, err
		}
		for _, c := range contacts {
			if strings.EqualFold(c.Name, *name) {
				id := c.ID
				contactID = &id
				break
			}
		}

		if contactID == nil {
			createdContact, err := s.contactService.Create(ctx, userID, contact.CreateInput{Name: *name})
			if err != nil {
				logger.Error("debt_create_failed", "error", err)
				return nil, err
			}
			id := createdContact.ID
			contactID = &id
		}
	}

	item := Debt{
		ID:                   uuid.NewString(),
		ClientID:             normalizeOptionalString(input.ClientID),
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

	if err := s.repo.Create(ctx, userID, item); err != nil {
		logger.Error("debt_create_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to create debt", err)
	}

	created, err := s.repo.GetByID(ctx, userID, item.ID)
	if err != nil {
		logger.Error("debt_create_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to load debt", err)
	}
	if created == nil {
		return nil, apperrors.New(apperrors.KindInternal, "created debt not found")
	}

	logger.Info("debt_create_succeeded", "debt_id", created.ID)
	return created, nil
}

func (s *service) Get(ctx context.Context, userID, debtID string) (*Debt, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "get")
	logger.Info("debt_get_started", "user_id", userID, "debt_id", debtID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(debtID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "debtId is required")
	}

	item, err := s.repo.GetByID(ctx, userID, debtID)
	if err != nil {
		logger.Error("debt_get_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to get debt", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "debt not found")
	}

	logger.Info("debt_get_succeeded", "debt_id", item.ID)
	return item, nil
}

func (s *service) List(ctx context.Context, userID string) ([]Debt, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "list")
	logger.Info("debt_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("debt_list_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to list debts", err)
	}

	logger.Info("debt_list_succeeded", "count", len(items))
	return items, nil
}

func (s *service) CreatePayment(ctx context.Context, userID, debtID string, input CreatePaymentInput) (*DebtPaymentLink, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "create_payment")
	logger.Info("debt_create_payment_started", "user_id", userID, "debt_id", debtID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(debtID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "debtId is required")
	}
	if s.txService == nil {
		return nil, apperrors.New(apperrors.KindInternal, "transaction service not configured")
	}

	transactionID := strings.TrimSpace(input.TransactionID)
	if transactionID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "transaction_id is required")
	}

	debtItem, err := s.repo.GetByID(ctx, userID, debtID)
	if err != nil {
		logger.Error("debt_create_payment_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to load debt", err)
	}
	if debtItem == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "debt not found")
	}

	tx, err := s.txService.Get(ctx, userID, transactionID)
	if err != nil {
		logger.Error("debt_create_payment_failed", "error", err)
		return nil, err
	}

	amountRat, ok := new(big.Rat).SetString(tx.Amount.String())
	if !ok {
		return nil, apperrors.New(apperrors.KindValidation, "transaction amount is invalid")
	}
	principalRat, ok := new(big.Rat).SetString(debtItem.Principal)
	if !ok {
		return nil, apperrors.New(apperrors.KindValidation, "principal is invalid")
	}
	outstandingRat, ok := new(big.Rat).SetString(debtItem.OutstandingPrincipal)
	if !ok {
		return nil, apperrors.New(apperrors.KindValidation, "outstanding_principal is invalid")
	}
	accruedRat, ok := new(big.Rat).SetString(debtItem.AccruedInterest)
	if !ok {
		return nil, apperrors.New(apperrors.KindValidation, "accrued_interest is invalid")
	}

	isTopUp := false
	if debtItem.Direction == "borrowed" && tx.Type == "income" {
		isTopUp = true
	} else if debtItem.Direction == "lent" && tx.Type == "expense" {
		isTopUp = true
	}

	var principalPaidRat *big.Rat
	var interestPaidRat *big.Rat
	var newPrincipal *big.Rat
	var newOutstanding *big.Rat
	var newAccrued *big.Rat
	status := debtItem.Status

	if isTopUp {
		newPrincipal = new(big.Rat).Add(principalRat, amountRat)
		newOutstanding = new(big.Rat).Add(outstandingRat, amountRat)
		newAccrued = accruedRat
		principalPaidRat = big.NewRat(0, 1)
		interestPaidRat = big.NewRat(0, 1)
	} else {
		if debtItem.Direction == "borrowed" && tx.Type != "expense" {
			return nil, apperrors.New(apperrors.KindValidation, "transaction.type must be expense for borrowed debt payment")
		}
		if debtItem.Direction == "lent" && tx.Type != "income" {
			return nil, apperrors.New(apperrors.KindValidation, "transaction.type must be income for lent debt collection")
		}

		if input.PrincipalPaid != nil || input.InterestPaid != nil {
			if input.PrincipalPaid != nil {
				v := strings.TrimSpace(*input.PrincipalPaid)
				if v != "" {
					if !isValidDecimal(v) {
						return nil, apperrors.New(apperrors.KindValidation, "principal_paid must be a decimal string")
					}
					parsed, ok := new(big.Rat).SetString(v)
					if !ok {
						return nil, apperrors.New(apperrors.KindValidation, "principal_paid must be a decimal string")
					}
					principalPaidRat = parsed
				}
			}
			if input.InterestPaid != nil {
				v := strings.TrimSpace(*input.InterestPaid)
				if v != "" {
					if !isValidDecimal(v) {
						return nil, apperrors.New(apperrors.KindValidation, "interest_paid must be a decimal string")
					}
					parsed, ok := new(big.Rat).SetString(v)
					if !ok {
						return nil, apperrors.New(apperrors.KindValidation, "interest_paid must be a decimal string")
					}
					interestPaidRat = parsed
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
				return nil, apperrors.New(apperrors.KindValidation, "principal_paid + interest_paid must equal transaction.amount")
			}
		} else {
			interestFirst := true
			if debtItem.InterestRule != nil && *debtItem.InterestRule == "principal_first" {
				interestFirst = false
			}

			remaining := new(big.Rat).Set(amountRat)
			var principalPaidLocal *big.Rat
			var interestPaidLocal *big.Rat
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
				return nil, apperrors.New(apperrors.KindValidation, "payment exceeds total due")
			}
		}

		newPrincipal = principalRat
		newOutstanding = new(big.Rat).Sub(outstandingRat, principalPaidRat)
		newAccrued = new(big.Rat).Sub(accruedRat, interestPaidRat)

		if newOutstanding.Sign() < 0 || newAccrued.Sign() < 0 {
			return nil, apperrors.New(apperrors.KindValidation, "outstanding cannot become negative")
		}
		if newOutstanding.Sign() == 0 && newAccrued.Sign() == 0 {
			status = "closed"
		}
	}

	principalPaidStr := ratToDecimalString(principalPaidRat)
	interestPaidStr := ratToDecimalString(interestPaidRat)

	now := time.Now().UTC()
	var closedAt *time.Time
	if status == "closed" {
		if debtItem.Status == "closed" {
			closedAt = debtItem.ClosedAt
		} else {
			closedAt = &now
		}
	}

	link := DebtPaymentLink{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		TransactionID: transactionID,
		PrincipalPaid: &principalPaidStr,
		InterestPaid:  &interestPaidStr,
		CreatedAt:     now,
	}

	update := DebtUpdate{
		Principal:            ratToDecimalString(newPrincipal),
		OutstandingPrincipal: ratToDecimalString(newOutstanding),
		AccruedInterest:      ratToDecimalString(newAccrued),
		Status:               status,
		ClosedAt:             closedAt,
		UpdatedAt:            now,
	}

	if err := s.repo.CreatePaymentLink(ctx, userID, link, update); err != nil {
		logger.Error("debt_create_payment_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to create debt payment link", err)
	}

	logger.Info("debt_create_payment_succeeded", "debt_id", debtID, "payment_link_id", link.ID)
	return &link, nil
}

func (s *service) ListPayments(ctx context.Context, userID, debtID string) ([]DebtPaymentLink, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "list_payments")
	logger.Info("debt_list_payments_started", "user_id", userID, "debt_id", debtID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(debtID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "debtId is required")
	}

	items, err := s.repo.ListPaymentLinks(ctx, userID, debtID)
	if err != nil {
		logger.Error("debt_list_payments_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to list debt payments", err)
	}

	logger.Info("debt_list_payments_succeeded", "count", len(items))
	return items, nil
}

func (s *service) ListPaymentsByTransaction(ctx context.Context, userID, transactionID string) ([]DebtPaymentLink, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "list_payments_by_transaction")
	logger.Info("debt_list_payments_by_transaction_started", "user_id", userID, "transaction_id", transactionID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	transactionID = strings.TrimSpace(transactionID)
	if transactionID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "transactionId is required")
	}
	if s.txService == nil {
		return nil, apperrors.New(apperrors.KindInternal, "transaction service not configured")
	}

	if _, err := s.txService.Get(ctx, userID, transactionID); err != nil {
		logger.Error("debt_list_payments_by_transaction_failed", "error", err)
		return nil, err
	}

	items, err := s.repo.ListPaymentLinksByTransaction(ctx, userID, transactionID)
	if err != nil {
		logger.Error("debt_list_payments_by_transaction_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to list debt links", err)
	}

	logger.Info("debt_list_payments_by_transaction_succeeded", "count", len(items))
	return items, nil
}

func (s *service) CreateInstallment(ctx context.Context, userID, debtID string, input CreateInstallmentInput) (*DebtInstallment, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "create_installment")
	logger.Info("debt_create_installment_started", "user_id", userID, "debt_id", debtID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(debtID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "debtId is required")
	}

	if input.InstallmentNo <= 0 {
		return nil, apperrors.New(apperrors.KindValidation, "installment_no must be > 0")
	}

	dueDate := strings.TrimSpace(input.DueDate)
	if dueDate == "" {
		return nil, apperrors.New(apperrors.KindValidation, "due_date is required")
	}
	if _, err := time.Parse("2006-01-02", dueDate); err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "due_date must be YYYY-MM-DD")
	}

	amountDue := strings.TrimSpace(input.AmountDue)
	if amountDue == "" {
		return nil, apperrors.New(apperrors.KindValidation, "amount_due is required")
	}
	if !isValidDecimal(amountDue) {
		return nil, apperrors.New(apperrors.KindValidation, "amount_due must be a decimal string")
	}

	amountPaid := "0"
	if input.AmountPaid != nil {
		v := strings.TrimSpace(*input.AmountPaid)
		if v != "" {
			if !isValidDecimal(v) {
				return nil, apperrors.New(apperrors.KindValidation, "amount_paid must be a decimal string")
			}
			amountPaid = v
		}
	}

	status := "pending"
	if input.Status != nil {
		v := strings.TrimSpace(*input.Status)
		if v != "" {
			if v != "pending" && v != "paid" && v != "overdue" {
				return nil, apperrors.New(apperrors.KindValidation, "status is invalid")
			}
			status = v
		}
	}

	item := DebtInstallment{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		InstallmentNo: input.InstallmentNo,
		DueDate:       dueDate,
		AmountDue:     amountDue,
		AmountPaid:    amountPaid,
		Status:        status,
	}

	if err := s.repo.CreateInstallment(ctx, userID, item); err != nil {
		logger.Error("debt_create_installment_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to create installment", err)
	}

	logger.Info("debt_create_installment_succeeded", "debt_id", debtID, "installment_id", item.ID)
	return &item, nil
}

func (s *service) ListInstallments(ctx context.Context, userID, debtID string) ([]DebtInstallment, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "debt", "operation", "list_installments")
	logger.Info("debt_list_installments_started", "user_id", userID, "debt_id", debtID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(debtID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "debtId is required")
	}

	items, err := s.repo.ListInstallments(ctx, userID, debtID)
	if err != nil {
		logger.Error("debt_list_installments_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to list installments", err)
	}

	logger.Info("debt_list_installments_succeeded", "count", len(items))
	return items, nil
}

func passThroughOrWrapInternal(message string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.Wrap(apperrors.KindInternal, message, err)
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func isValidDecimal(v string) bool {
	_, ok := new(big.Rat).SetString(strings.TrimSpace(v))
	return ok
}

func minRat(a, b *big.Rat) *big.Rat {
	if a.Cmp(b) <= 0 {
		return new(big.Rat).Set(a)
	}
	return new(big.Rat).Set(b)
}

func ratToDecimalString(v *big.Rat) string {
	if v == nil {
		return "0.00"
	}
	return v.FloatString(2)
}
