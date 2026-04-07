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

func (s *DebtService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error) {
	var contactID *uuid.UUID
	if req.ContactID != nil && *req.ContactID != "" {
		id, err := uuid.Parse(*req.ContactID)
		if err == nil {
			contactID = &id
		}
	}

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

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	d := entity.Debt{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID:               userID,
		AccountID:            &accountID,
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
	}

	if err := s.repo.CreateDebt(ctx, d); err != nil {
		return nil, err
	}

	created, err := s.repo.GetDebt(ctx, userID, d.ID)
	if err != nil {
		return nil, err
	}
	if created == nil {
		return nil, nil
	}
	resp := dto.NewDebtResponse(*created)
	return &resp, nil
}

func (s *DebtService) Get(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) (*dto.DebtResponse, error) {
	it, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewDebtResponse(*it)
	return &resp, nil
}

func (s *DebtService) List(ctx context.Context, userID uuid.UUID) ([]dto.DebtResponse, error) {
	items, err := s.repo.ListDebts(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewDebtResponses(items), nil
}

func (s *DebtService) Update(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.UpdateDebtRequest) (*dto.DebtResponse, error) {
	cur, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		cur.Name = req.Name
	}
	if req.DueDate != nil {
		cur.DueDate = *req.DueDate
	}
	if req.Status != nil {
		cur.Status = *req.Status
	}
	if req.InterestRate != nil {
		cur.InterestRate = req.InterestRate
	}

	if cur.Status == "paid" && cur.ClosedAt == nil {
		now := time.Now().UTC()
		cur.ClosedAt = &now
	}

	if err := s.repo.UpdateDebt(ctx, userID, *cur); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	resp := dto.NewDebtResponse(*updated)
	return &resp, nil
}

func (s *DebtService) Delete(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error {
	return s.repo.DeleteDebt(ctx, userID, debtID)
}

func (s *DebtService) AddPayment(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtPaymentRequest) (*dto.DebtResponse, error) {
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

		if paid != nil && accrued != nil && paid.Cmp(accrued) >= 0 {
			interestPaid = accrued.FloatString(2)
			rem := new(big.Rat).Sub(paid, accrued)
			principalPaid = rem.FloatString(2)
		} else if paid != nil {
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
	if newPrincipalRat != nil && pPaidRat != nil {
		newPrincipalRat.Sub(newPrincipalRat, pPaidRat)
	}

	newOutstandingRat, _ := new(big.Rat).SetString(debt.OutstandingPrincipal)
	if newOutstandingRat != nil && pPaidRat != nil {
		newOutstandingRat.Sub(newOutstandingRat, pPaidRat)
	}

	newAccruedRat, _ := new(big.Rat).SetString(debt.AccruedInterest)
	iPaidRat, _ := new(big.Rat).SetString(interestPaid)
	if newAccruedRat != nil && iPaidRat != nil {
		newAccruedRat.Sub(newAccruedRat, iPaidRat)
	}

	newStatus := debt.Status
	var closedAt *time.Time
	if newOutstandingRat != nil && newAccruedRat != nil && newOutstandingRat.Sign() <= 0 && newAccruedRat.Sign() <= 0 {
		newStatus = "paid"
		now := time.Now().UTC()
		closedAt = &now
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	link := entity.DebtPaymentLink{
		BaseEntity: entity.BaseEntity{
			ID: utils.NewID(),
		},
		DebtID:        debtID,
		TransactionID: transactionID,
		PrincipalPaid: &principalPaid,
		InterestPaid:  &interestPaid,
		CreatedAt:     time.Now().UTC(),
	}

	err = s.repo.CreatePaymentLink(ctx, userID, link, newPrincipalRat.FloatString(2), newOutstandingRat.FloatString(2), newAccruedRat.FloatString(2), newStatus, closedAt)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	resp := dto.NewDebtResponse(*updated)
	return &resp, nil
}

func (s *DebtService) ListPayments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]dto.DebtPaymentLinkResponse, error) {
	items, err := s.repo.ListPaymentLinks(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	return dto.NewDebtPaymentLinkResponses(items), nil
}
