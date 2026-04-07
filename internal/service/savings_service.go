package service
 
import (
	"context"
	"time"
 
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)
 
type SavingsService struct {
	repo interfaces.SavingsRepository
}
 
func NewSavingsService(repo interfaces.SavingsRepository) *SavingsService {
	return &SavingsService{repo: repo}
}
 
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
		Status:           "active",
	}
 
	if err := s.repo.CreateSavings(ctx, userID, instr); err != nil {
		return nil, err
	}
	resp := dto.NewSavingsResponse(instr)
	return &resp, nil
}
 
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
 
func (s *SavingsService) ListSavings(ctx context.Context, userID uuid.UUID) ([]dto.SavingsResponse, error) {
	items, err := s.repo.ListSavings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewSavingsResponses(items), nil
}
 
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
		if *req.Status == "closed" || *req.Status == "matured" {
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
 
func (s *SavingsService) DeleteSavings(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.DeleteSavings(ctx, userID, id)
}
