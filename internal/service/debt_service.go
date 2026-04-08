package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/repository/postgres"
)

// DebtService quản lý các khoản nợ (IOU), các khoản cho vay và nợ chi phí dùng chung.
// Nó tích hợp với TransactionService để đảm bảo các khoản cho vay, đi vay
// và trả nợ được phản ánh chính xác trong số dư tài khoản và lịch sử giao dịch.
//
// Các trách nhiệm chính:
// - Ghi nhận nợ mới và theo dõi việc giải ngân.
// - Tính toán nợ gốc chưa thanh toán và lãi tích lũy.
// - Liên kết các khoản thanh toán (giao dịch) với các công cụ nợ cụ thể.
// - Hoàn trả trạng thái nợ khi các giao dịch trả nợ bị xóa.
//
// Nợ gốc và lãi được xử lý dưới dạng số thập phân độ chính xác cao (thông qua big.Rat nội bộ
// khi tính toán, nhưng lưu trữ dạng chuỗi) để tránh lỗi số dấu phẩy động.
type DebtService struct {
	repo       interfaces.DebtRepository
	contactSvc interfaces.ContactService
	db         *database.Postgres
}

// NewDebtService khởi tạo một dịch vụ quản lý nợ mới.
func NewDebtService(repo interfaces.DebtRepository, contactSvc interfaces.ContactService, db *database.Postgres) *DebtService {
	return &DebtService{repo: repo, contactSvc: contactSvc, db: db}
}

// Create ghi nhận một khoản nợ mới. Nó bao bọc CreateTx trong một giao dịch cơ sở dữ liệu.
func (s *DebtService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error) {
	var resp *dto.DebtResponse
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		resp, err = s.CreateTx(ctx, tx, userID, req)
		return err
	})
	return resp, err
}

// CreateTx thực hiện việc tạo bản ghi nợ một cách nguyên tử.
// Nếu CreateTransaction là true, nó cũng sẽ tạo một giao dịch Thu thập/Chi phí
// trong sổ cái trung tâm để phản ánh việc giải ngân khoản vay ban đầu.
func (s *DebtService) CreateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error) {
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

	var originatingTxID *uuid.UUID
	if req.OriginatingTransactionID != nil && *req.OriginatingTransactionID != "" {
		if id, err := uuid.Parse(*req.OriginatingTransactionID); err == nil {
			originatingTxID = &id
		}
	}

	// 3. Create Associated Transaction if requested (BEFORE creating debt to get its ID)
	if req.CreateTransaction {
		txType := entity.TransactionTypeIncome
		if req.Direction == entity.DebtDirectionLent {
			txType = entity.TransactionTypeExpense
		}

		desc := "Initial debt disbursement"
		if req.Name != nil {
			desc = "Debt: " + *req.Name
		} else if req.ContactName != nil {
			desc = "Debt with " + *req.ContactName
		}

		createTx := entity.Transaction{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
			},
			Type:         txType,
			OccurredAt:   utils.Now(),
			OccurredDate: utils.NowDateString(),
			Amount:       principal,
			Description:  &desc,
			AccountID:    &accountID,
			Status:       entity.TransactionStatusPosted,
		}

		// Create a single line item for the transaction
		lineItems := []entity.TransactionLineItem{
			{
				BaseEntity: entity.BaseEntity{ID: utils.NewID()},
				Amount:     principal,
				Note:       &desc,
			},
		}

		if err := postgres.CreateTransactionTx(ctx, tx, userID, createTx, lineItems, nil); err != nil {
			return nil, fmt.Errorf("failed to create associated transaction: %w", err)
		}

		// Link the newly created transaction to the debt
		originatingTxID = &createTx.ID
	}

	d := entity.Debt{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID:                   userID,
		AccountID:                &accountID,
		OriginatingTransactionID: originatingTxID,
		Direction:                req.Direction,
		Name:                     utils.NormalizeOptionalString(req.Name),
		ContactID:                contactID,
		Principal:                principal,
		StartDate:                req.StartDate,
		DueDate:                  req.DueDate,
		InterestRate:             utils.NormalizeOptionalString(req.InterestRate),
		InterestRule:             utils.NormalizeOptionalString(req.InterestRule),
		OutstandingPrincipal:     principal,
		AccruedInterest:          "0",
		Status:                   entity.DebtStatusActive,
	}

	if err := s.repo.CreateDebtTx(ctx, tx, d); err != nil {
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

// Update chỉnh sửa thông tin khoản nợ.
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

	if cur.Status == entity.DebtStatusPaid && cur.ClosedAt == nil {
		now := utils.Now()
		cur.ClosedAt = &now
	}

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		return s.repo.UpdateDebtTx(ctx, tx, userID, *cur)
	})
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

// Delete xóa một khoản nợ.
func (s *DebtService) Delete(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error {
	return s.repo.DeleteDebtTx(ctx, nil, userID, debtID)
}

// AddPayment liên kết một giao dịch hiện có như một khoản thanh toán cho khoản nợ.
func (s *DebtService) AddPayment(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtPaymentRequest) (*dto.DebtResponse, error) {
	var resp *dto.DebtResponse
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		resp, err = s.AddPaymentTx(ctx, tx, userID, debtID, req)
		return err
	})
	return resp, err
}

// AddPaymentTx liên kết một giao dịch hiện có như một khoản thanh toán cho khoản nợ.
// Nó tính toán tác động lên nợ gốc chưa thanh toán và lãi tích lũy.
//
// Logic phân bổ thanh toán:
// 1. Trả hết lãi tích lũy trước.
// 2. Số tiền còn lại được áp dụng vào nợ gốc chưa thanh toán.
// 3. Nếu mọi thứ đã được trả hết, trạng thái được cập nhật thành 'paid'.
func (s *DebtService) AddPaymentTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debtID uuid.UUID, req dto.DebtPaymentRequest) (*dto.DebtResponse, error) {
	debt, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	if debt == nil {
		return nil, errors.New("debt not found")
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
	pPaidRat, _ := new(big.Rat).SetString(principalPaid)
	// Note: We don't change the original principal, just outstanding
	
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
		newStatus = entity.DebtStatusPaid
		now := utils.Now()
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
		CreatedAt:     utils.Now(),
	}

	if err := s.repo.CreatePaymentLinkTx(ctx, tx, userID, link, debt.Principal, newOutstandingRat.FloatString(2), newAccruedRat.FloatString(2), newStatus, closedAt); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetDebt(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	resp := dto.NewDebtResponse(*updated)
	return &resp, nil
}

// Repay thực hiện việc trả nợ bằng cách tạo một giao dịch mới.
func (s *DebtService) Repay(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtRepayRequest) (*dto.DebtResponse, error) {
	// Start transaction
	var resp *dto.DebtResponse
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		debt, err := s.repo.GetDebt(ctx, userID, debtID)
		if err != nil {
			return err
		}
		if debt == nil {
			return errors.New("debt not found")
		}

		accountID, err := uuid.Parse(req.AccountID)
		if err != nil {
			return errors.New("invalid account ID")
		}

		// 1. Create Transaction for repayment
		txType := entity.TransactionTypeExpense
		if debt.Direction == entity.DebtDirectionLent {
			txType = entity.TransactionTypeIncome
		}

		desc := "Debt repayment"
		if debt.Name != nil {
			desc = "Repayment: " + *debt.Name
		}
		if req.Note != nil {
			desc = *req.Note
		}

		createTx := entity.Transaction{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
			},
			Type:         txType,
			OccurredAt:   utils.Now(),
			OccurredDate: utils.NowDateString(),
			Amount:       req.Amount,
			Description:  &desc,
			AccountID:    &accountID,
			Status:       entity.TransactionStatusPosted,
		}

		lineItems := []entity.TransactionLineItem{
			{
				BaseEntity: entity.BaseEntity{ID: utils.NewID()},
				Amount:     req.Amount,
				Note:       &desc,
			},
		}

		if err := postgres.CreateTransactionTx(ctx, tx, userID, createTx, lineItems, nil); err != nil {
			return fmt.Errorf("failed to create repayment transaction: %w", err)
		}

		// 2. Add Payment link
		resp, err = s.AddPaymentTx(ctx, tx, userID, debtID, dto.DebtPaymentRequest{
			TransactionID: createTx.ID.String(),
			AmountPaid:    &req.Amount,
		})
		return err
	})

	return resp, err
}

func (s *DebtService) ListPayments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]dto.DebtPaymentLinkResponse, error) {
	links, err := s.repo.ListPaymentLinks(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	resps := make([]dto.DebtPaymentLinkResponse, 0, len(links))
	for _, l := range links {
		resps = append(resps, dto.NewDebtPaymentLinkResponse(l))
	}
	return resps, nil
}

// CleanupTransactionLinksTx gỡ bỏ tất cả các khoản nợ và liên kết thanh toán liên quan đến một giao dịch.
// Điều này được sử dụng khi một giao dịch bị xóa để duy trì tính toàn vẹn của sổ cái.
func (s *DebtService) CleanupTransactionLinksTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error {
	// 1. Xóa các liên kết thanh toán (điều này hoàn trả các tác động trả nợ nếu chúng ta tính toán lại,
	// nhưng hiện tại chúng ta chỉ xóa các liên kết và bên gọi xử lý trạng thái nợ nếu cần)
	// Thực tế, nếu chúng ta xóa một giao dịch trả nợ, chúng ta nên hoàn trả nợ gốc còn thiếu của khoản nợ.
	
	links, err := s.repo.ListPaymentLinksByTransaction(ctx, userID, transactionID)
	if err != nil {
		return err
	}

	for _, link := range links {
		debt, err := s.repo.GetDebt(ctx, userID, link.DebtID)
		if err != nil {
			return err
		}
		if debt == nil {
			continue
		}

		// Hoàn trả trạng thái nợ
		pPaid, _ := new(big.Rat).SetString(*link.PrincipalPaid)
		iPaid, _ := new(big.Rat).SetString(*link.InterestPaid)
		
		curOutstanding, _ := new(big.Rat).SetString(debt.OutstandingPrincipal)
		curAccrued, _ := new(big.Rat).SetString(debt.AccruedInterest)
		
		newOutstanding := new(big.Rat).Add(curOutstanding, pPaid)
		newAccrued := new(big.Rat).Add(curAccrued, iPaid)
		
		debt.OutstandingPrincipal = newOutstanding.FloatString(2)
		debt.AccruedInterest = newAccrued.FloatString(2)
		debt.Status = entity.DebtStatusActive
		debt.ClosedAt = nil
		debt.UpdatedAt = utils.Now()

		if err := s.repo.UpdateDebtTx(ctx, tx, userID, *debt); err != nil {
			return err
		}
	}

	// 2. Xóa chính các liên kết thanh toán
	if err := s.repo.DeletePaymentLinksByTransactionTx(ctx, tx, userID, transactionID); err != nil {
		return err
	}

	// 3. Xóa các khoản nợ bắt nguồn từ giao dịch này
	return s.repo.DeleteDebtsByOriginatingTransactionTx(ctx, tx, userID, transactionID)
}
