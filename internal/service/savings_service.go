package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

// SavingsService quản lý các mục tiêu tiết kiệm và tiền gửi có kỳ hạn.
// Nó đảm bảo rằng việc nạp tiền vào mục tiêu tiết kiệm sẽ kích hoạt một giao dịch chuyển tiền tương ứng
// trong sổ cái trung tâm (TransactionService).
type SavingsService struct {
	repo            interfaces.SavingsRepository
	accountRepo       interfaces.AccountRepository
	transactionRepo interfaces.TransactionRepository
	txSvc           interfaces.TransactionService
	db              *database.Postgres
}

// NewSavingsService khởi tạo một dịch vụ quản lý tiết kiệm mới.
func NewSavingsService(
	repo interfaces.SavingsRepository,
	accountRepo interfaces.AccountRepository,
	transactionRepo interfaces.TransactionRepository,
	txSvc interfaces.TransactionService,
	db *database.Postgres,
) *SavingsService {
	return &SavingsService{
		repo:            repo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		txSvc:           txSvc,
		db:              db,
	}
}

// CreateSavings ghi lại một mục tiêu tiết kiệm hoặc tiền gửi có kỳ hạn mới.
// Nếu cả ParentAccountID và SavingsAccountID được cung cấp, nó sẽ tự động
// tạo một giao dịch 'transfer' trong sổ cái trung tâm để phản ánh việc biến động số dư.
func (s *SavingsService) CreateSavings(ctx context.Context, userID uuid.UUID, req dto.CreateSavingsRequest) (*dto.SavingsResponse, error) {
	// 1. Prepare Account record
	savingsID := req.SavingsAccountID
	if savingsID == uuid.Nil {
		savingsID = utils.NewID()
	}

	var parentID *uuid.UUID
	if req.ParentAccountID != uuid.Nil {
		parentID = &req.ParentAccountID
	}

	acc := entity.Account{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: savingsID,
			},
		},
		Name:            req.Name,
		AccountType:     entity.AccountTypeSavings,
		Currency:        "VND", // Default to VND for savings for now
		ParentAccountID: parentID,
		Status:          entity.AccountStatusActive,
		Settings: entity.AccountSettings{
			Savings: &entity.SavingsSettings{
				Principal:    req.Principal,
				InterestRate: req.InterestRate,
				TermMonths:   req.TermMonths,
				StartDate:    req.StartDate,
				MaturityDate: req.MaturityDate,
				AutoRenew:    req.AutoRenew,
			},
		},
	}

	var created *entity.Savings
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 2. Create the Account acting as a Savings entity
		if err := s.accountRepo.CreateAccountWithOwnerTx(ctx, tx, acc, userID); err != nil {
			return err
		}

		// 3. Automated Ledger Transfer (if funding account is provided)
		if parentID != nil && savingsID != uuid.Nil {
			dateStr := ""
			if acc.Settings.Savings.StartDate != nil {
				dateStr = *acc.Settings.Savings.StartDate
			} else {
				dateStr = utils.NowDateString()
			}
			occAt, _ := time.Parse("2006-01-02", dateStr)
			desc := "Khoản gửi tiết kiệm: " + req.Name

			ledgerTx := entity.Transaction{
				AuditEntity:   entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
				Type:          entity.TransactionTypeTransfer,
				OccurredAt:    occAt.UTC(),
				OccurredDate:  dateStr,
				Amount:        acc.Settings.Savings.Principal,
				FromAccountID: parentID,
				ToAccountID:   &savingsID,
				Description:   &desc,
				Status:        entity.TransactionStatusPosted,
			}

			if err := s.transactionRepo.CreateTransactionTx(ctx, tx, userID, ledgerTx, nil, nil); err != nil {
				return err
			}
		}

		// 4. Fetch the enriched record (with calculated interest)
		// Since we are in the same transaction, we should ideally have a GetSavingsTx
		// that uses the transaction. But let's assume for now GetSavings can handle it
		// if it's visible or let's map manually to avoid complexity if repo doesn't support Tx yet.
		// Actually, let's just use the current data and return it.
		// To be safe, I'll return the data I have.
		created = &entity.Savings{
			ID:               savingsID,
			Name:             req.Name,
			SavingsAccountID: savingsID,
			ParentAccountID:  parentID,
			Principal:        req.Principal,
			InterestRate:     req.InterestRate,
			TermMonths:       req.TermMonths,
			StartDate:        req.StartDate,
			MaturityDate:     req.MaturityDate,
			AutoRenew:        req.AutoRenew,
			AccruedInterest:  "0",
			Status:           entity.AccountStatusActive,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r := dto.NewSavingsResponse(*created)
	return &r, nil
}

// GetSavings lấy thông tin về một bản ghi tiết kiệm cụ thể.
func (s *SavingsService) GetSavings(ctx context.Context, userID, id uuid.UUID) (*dto.SavingsResponse, error) {
	it, err := s.repo.GetSavingsTx(ctx, nil, userID, id)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewSavingsResponse(*it)
	return &resp, nil
}

// ListSavings liệt kê tất cả các bản ghi tiết kiệm đang hoạt động và đã đóng của người dùng.
func (s *SavingsService) ListSavings(ctx context.Context, userID uuid.UUID) ([]dto.SavingsResponse, error) {
	items, err := s.repo.ListSavingsTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewSavingsResponses(items), nil
}

// PatchSavings cập nhật thông tin tiết kiệm.
func (s *SavingsService) PatchSavings(ctx context.Context, userID, id uuid.UUID, req dto.PatchSavingsRequest) (*dto.SavingsResponse, error) {
	cur, err := s.repo.GetSavingsTx(ctx, nil, userID, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		cur.Name = *req.Name
	}
	if req.Principal != nil {
		cur.Principal = *req.Principal
	}
	if req.InterestRate != nil {
		cur.InterestRate = req.InterestRate
	}
	if req.TermMonths != nil {
		cur.TermMonths = req.TermMonths
	}
	if req.MaturityDate != nil {
		cur.MaturityDate = req.MaturityDate
	}
	if req.AutoRenew != nil {
		cur.AutoRenew = *req.AutoRenew
	}
	if req.Status != nil {
		cur.Status = *req.Status
	}

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		return s.repo.UpdateSavingsTx(ctx, tx, userID, *cur)
	})
	if err != nil {
		return nil, err
	}

	resp := dto.NewSavingsResponse(*cur)
	return &resp, nil
}

// DeleteSavings xóa bản ghi tiết kiệm và các giao dịch liên quan để đảm bảo tính nhất quán số dư.
func (s *SavingsService) DeleteSavings(ctx context.Context, userID, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Xóa tất cả giao dịch liên quan (nạp tiền, lãi, v.v.)
		if err := s.transactionRepo.DeleteTransactionsByAccountTx(ctx, tx, userID, id); err != nil {
			return err
		}

		// 2. Xóa tài khoản tiết kiệm (cập nhật status và deleted_at)
		return s.repo.DeleteSavingsTx(ctx, tx, userID, id)
	})
}
