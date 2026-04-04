package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type SavingsService struct {
	repo interfaces.SavingsRepository
}

func NewSavingsService(repo interfaces.SavingsRepository) *SavingsService {
	return &SavingsService{repo: repo}
}

func (s *SavingsService) CreateSavingsInstrument(ctx context.Context, userID string, req dto.CreateSavingsInstrumentRequest) (*entity.SavingsInstrument, error) {
	instr := entity.SavingsInstrument{
		ID:               uuid.NewString(),
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
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := s.repo.CreateSavingsInstrument(ctx, userID, instr); err != nil {
		return nil, err
	}
	return &instr, nil
}

func (s *SavingsService) GetSavingsInstrument(ctx context.Context, userID, id string) (*entity.SavingsInstrument, error) {
	return s.repo.GetSavingsInstrument(ctx, userID, id)
}

func (s *SavingsService) ListSavingsInstruments(ctx context.Context, userID string) ([]entity.SavingsInstrument, error) {
	return s.repo.ListSavingsInstruments(ctx, userID)
}

func (s *SavingsService) DeleteSavingsInstrument(ctx context.Context, userID, id string) error {
	return s.repo.DeleteSavingsInstrument(ctx, userID, id)
}
