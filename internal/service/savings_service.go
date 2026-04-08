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
	"github.com/sonbn-225/goen-api/internal/repository/postgres"
)

// SavingsService manages savings goals and term deposits.
// It ensures that funding a savings goal triggers a corresponding transfer
// in the central ledger (TransactionService).
type SavingsService struct {
	repo  interfaces.SavingsRepository
	txSvc interfaces.TransactionService
	db    *database.Postgres
}

// NewSavingsService creates a new savings management service.
func NewSavingsService(repo interfaces.SavingsRepository, txSvc interfaces.TransactionService, db *database.Postgres) *SavingsService {
	return &SavingsService{repo: repo, txSvc: txSvc, db: db}
}
 
// CreateSavings records a new savings goal or term deposit. 
// If both ParentAccountID and SavingsAccountID are provided, it automatically
// creates a 'transfer' transaction in the central ledger to reflect the movement of funds.
func (s *SavingsService) CreateSavings(ctx context.Context, userID uuid.UUID, req dto.CreateSavingsRequest) (*dto.SavingsResponse, error) {
	instr := entity.Savings{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		SavingsAccountID: req.SavingsAccountID,
		ParentAccountID:  req.ParentAccountID,
		Principal:        req.Principal,
		InterestRate:     req.InterestRate,
		TermMonths:       req.TermMonths,
		StartDate:        req.StartDate,
		MaturityDate:     req.MaturityDate,
		AutoRenew:        req.AutoRenew,
		AccruedInterest:  "0",
		Status:           entity.SavingsStatusActive,
	}

	var resp *dto.SavingsResponse
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Create Savings Record
		if err := s.repo.CreateSavingsTx(ctx, tx, userID, instr); err != nil {
			return err
		}

		// 2. Automated Ledger Transfer (if applicable)
		if instr.ParentAccountID != uuid.Nil && instr.SavingsAccountID != uuid.Nil {
			dateStr := ""
			if instr.StartDate != nil {
				dateStr = *instr.StartDate
			} else {
				dateStr = time.Now().Format("2006-01-02")
			}
			occAt, _ := time.Parse("2006-01-02", dateStr)
			desc := "Funding Savings: " + instr.ID.String()
			
			parentID := instr.ParentAccountID
			savingsID := instr.SavingsAccountID

			ledgerTx := entity.Transaction{
				AuditEntity:   entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
				Type:          entity.TransactionTypeTransfer,
				OccurredAt:    occAt.UTC(),
				OccurredDate:  dateStr,
				Amount:        instr.Principal,
				FromAccountID: &parentID,
				ToAccountID:   &savingsID,
				Description:   &desc,
				Status:        entity.TransactionStatusPosted,
			}
			
			if err := postgres.CreateTransactionTx(ctx, tx, userID, ledgerTx, nil, nil); err != nil {
				return err
			}
		}

		r := dto.NewSavingsResponse(instr)
		resp = &r
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
 
// GetSavings retrieves a specific savings record.
func (s *SavingsService) GetSavings(ctx context.Context, userID, id uuid.UUID) (*dto.SavingsResponse, error) {
	it, err := s.repo.GetSavings(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewSavingsResponse(*it)
	return &resp, nil
}
 
// ListSavings returns all active and closed savings records for a user.
func (s *SavingsService) ListSavings(ctx context.Context, userID uuid.UUID) ([]dto.SavingsResponse, error) {
	items, err := s.repo.ListSavings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewSavingsResponses(items), nil
}
 
// PatchSavings updates a savings goal. If the status is changed to 'Closed' or 'Matured',
// it captures the closure timestamp for record-keeping.
func (s *SavingsService) PatchSavings(ctx context.Context, userID, id uuid.UUID, req dto.PatchSavingsRequest) (*dto.SavingsResponse, error) {
	cur, err := s.repo.GetSavings(ctx, userID, id)
	if err != nil {
		return nil, err
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
		if *req.Status == entity.SavingsStatusClosed || *req.Status == entity.SavingsStatusMatured {
			now := time.Now().UTC()
			cur.ClosedAt = &now
		}
	}
 
	if err := s.repo.UpdateSavings(ctx, userID, *cur); err != nil {
		return nil, err
	}
 
	resp := dto.NewSavingsResponse(*cur)
	return &resp, nil
}
 
// DeleteSavings removes the savings record from the database.
func (s *SavingsService) DeleteSavings(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.DeleteSavings(ctx, userID, id)
}
