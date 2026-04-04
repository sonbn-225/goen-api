package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type DebtService struct {
	repo       interfaces.DebtRepository
	contactSvc interfaces.ContactService
}

func NewDebtService(repo interfaces.DebtRepository, contactSvc interfaces.ContactService) *DebtService {
	return &DebtService{repo: repo, contactSvc: contactSvc}
}

func (s *DebtService) Create(ctx context.Context, userID string, req dto.CreateDebtRequest) (*entity.Debt, error) {
	contactID := utils.NormalizeOptionalString(req.ContactID)
	if contactID == nil && req.ContactName != nil && strings.TrimSpace(*req.ContactName) != "" {
		id, err := s.contactSvc.GetOrCreateByName(ctx, userID, *req.ContactName)
		if err == nil {
			contactID = &id
		}
	}

	principal := strings.TrimSpace(req.Principal)
	if !utils.IsValidDecimal(principal) {
		return nil, errors.New("invalid principal amount")
	}

	now := time.Now().UTC()
	d := entity.Debt{
		ID:                   uuid.NewString(),
		ClientID:             utils.NormalizeOptionalString(req.ClientID),
		UserID:               userID,
		AccountID:            &req.AccountID,
		Direction:            req.Direction,
		Name:                 utils.NormalizeOptionalString(req.Name),
		ContactID:            contactID,
		Principal:            principal,
		StartDate:            req.StartDate,
		DueDate:              req.DueDate,
		InterestRate:         utils.NormalizeOptionalString(req.InterestRate),
		InterestRule:         utils.NormalizeOptionalString(req.InterestRule),
		OutstandingPrincipal: principal,
		AccruedInterest:      "0",
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.repo.CreateDebt(ctx, d); err != nil {
		return nil, err
	}
	return s.repo.GetDebt(ctx, userID, d.ID)
}

func (s *DebtService) Get(ctx context.Context, userID string, debtID string) (*entity.Debt, error) {
	return s.repo.GetDebt(ctx, userID, debtID)
}

func (s *DebtService) List(ctx context.Context, userID string) ([]entity.Debt, error) {
	return s.repo.ListDebts(ctx, userID)
}

func (s *DebtService) Update(ctx context.Context, userID string, debtID string, req dto.UpdateDebtRequest) (*entity.Debt, error) {
	cur, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil { cur.Name = req.Name }
	if req.DueDate != nil { cur.DueDate = *req.DueDate }
	if req.Status != nil { cur.Status = *req.Status }
	if req.InterestRate != nil { cur.InterestRate = req.InterestRate }
	cur.UpdatedAt = time.Now().UTC()

	if cur.Status == "paid" && cur.ClosedAt == nil {
		now := time.Now().UTC()
		cur.ClosedAt = &now
	}

	if err := s.repo.UpdateDebt(ctx, userID, *cur); err != nil {
		return nil, err
	}
	return s.repo.GetDebt(ctx, userID, debtID)
}

func (s *DebtService) Delete(ctx context.Context, userID string, debtID string) error {
	return s.repo.DeleteDebt(ctx, userID, debtID)
}

func (s *DebtService) AddPayment(ctx context.Context, userID string, debtID string, req dto.DebtPaymentRequest) (*entity.Debt, error) {
	debt, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}

	principalPaid := "0"
	interestPaid := "0"

	if req.AmountPaid != nil && utils.IsValidDecimal(*req.AmountPaid) {
		// Auto-distribute: interest first, then principal
		paid, _ := new(big.Rat).SetString(*req.AmountPaid)
		accrued, _ := new(big.Rat).SetString(debt.AccruedInterest)
		
		if paid.Cmp(accrued) >= 0 {
			interestPaid = accrued.FloatString(2)
			rem := new(big.Rat).Sub(paid, accrued)
			principalPaid = rem.FloatString(2)
		} else {
			interestPaid = paid.FloatString(2)
			principalPaid = "0"
		}
	} else {
		if req.PrincipalPaid != nil && utils.IsValidDecimal(*req.PrincipalPaid) {
			principalPaid = *req.PrincipalPaid
		}
		if req.InterestPaid != nil && utils.IsValidDecimal(*req.InterestPaid) {
			interestPaid = *req.InterestPaid
		}
	}

	// Update Debt state
	newPrincipalRat, _ := new(big.Rat).SetString(debt.Principal)
	pPaidRat, _ := new(big.Rat).SetString(principalPaid)
	newPrincipalRat.Sub(newPrincipalRat, pPaidRat)

	newOutstandingRat, _ := new(big.Rat).SetString(debt.OutstandingPrincipal)
	newOutstandingRat.Sub(newOutstandingRat, pPaidRat)

	newAccruedRat, _ := new(big.Rat).SetString(debt.AccruedInterest)
	iPaidRat, _ := new(big.Rat).SetString(interestPaid)
	newAccruedRat.Sub(newAccruedRat, iPaidRat)

	newStatus := debt.Status
	var closedAt *time.Time
	if newOutstandingRat.Sign() <= 0 && newAccruedRat.Sign() <= 0 {
		newStatus = "paid"
		now := time.Now().UTC()
		closedAt = &now
	}

	link := entity.DebtPaymentLink{
		ID:            uuid.NewString(),
		DebtID:        debtID,
		TransactionID: req.TransactionID,
		PrincipalPaid: &principalPaid,
		InterestPaid:  &interestPaid,
		CreatedAt:     time.Now().UTC(),
	}

	err = s.repo.CreatePaymentLink(ctx, userID, link, newPrincipalRat.FloatString(2), newOutstandingRat.FloatString(2), newAccruedRat.FloatString(2), newStatus, closedAt)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebt(ctx, userID, debtID)
}

func (s *DebtService) ListPayments(ctx context.Context, userID string, debtID string) ([]entity.DebtPaymentLink, error) {
	return s.repo.ListPaymentLinks(ctx, userID, debtID)
}
