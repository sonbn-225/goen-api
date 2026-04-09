package service

import (
	"context"
	"math/big"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

// InvestmentService manages holdings, reports, and investment-level actions.
type InvestmentService struct {
	repo interfaces.InvestmentRepository
}

// NewInvestmentService creates a new investment management service.
func NewInvestmentService(repo interfaces.InvestmentRepository) *InvestmentService {
	return &InvestmentService{repo: repo}
}

func (s *InvestmentService) ListHoldings(ctx context.Context, userID, accountID uuid.UUID) ([]dto.HoldingResponse, error) {
	items, err := s.repo.ListHoldingsTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewHoldingResponses(items), nil
}

func (s *InvestmentService) ListEligibleCorporateActions(ctx context.Context, userID, brokerAccountID uuid.UUID) ([]dto.EligibleAction, error) {
	// TODO: implement investment corporate actions flow.
	return []dto.EligibleAction{}, nil
}

func (s *InvestmentService) ClaimCorporateAction(ctx context.Context, userID, brokerAccountID, eventID uuid.UUID, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error) {
	return nil, apperr.Internal("not implemented")
}

func (s *InvestmentService) GetRealizedPNLReport(ctx context.Context, userID, accountID uuid.UUID) (*dto.RealizedPNLReport, error) {
	logs, err := s.repo.ListRealizedLogsTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}

	report := &dto.RealizedPNLReport{Items: []dto.RealizedPNLReportItem{}}
	totalNet := new(big.Rat)

	bySec := map[uuid.UUID]*dto.RealizedPNLReportItem{}
	for _, l := range logs {
		item, ok := bySec[l.SecurityID]
		if !ok {
			item = &dto.RealizedPNLReportItem{SecurityID: l.SecurityID}
			bySec[l.SecurityID] = item
			report.Items = append(report.Items, *item)
		}
		pnl, _ := new(big.Rat).SetString(l.RealizedPnL)
		totalNet.Add(totalNet, pnl)
	}
	report.TotalNet = totalNet.FloatString(2)
	return report, nil
}

func (s *InvestmentService) BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID uuid.UUID) (*dto.BackfillTradePrincipalResponse, error) {
	// TODO: implement real backfill flow.
	return &dto.BackfillTradePrincipalResponse{}, nil
}
