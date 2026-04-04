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
)

type InvestmentService struct {
	repo  interfaces.InvestmentRepository
	txSvc interfaces.TransactionService
}

func NewInvestmentService(
	repo interfaces.InvestmentRepository,
	txSvc interfaces.TransactionService,
) *InvestmentService {
	return &InvestmentService{
		repo:  repo,
		txSvc: txSvc,
	}
}

func (s *InvestmentService) GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*dto.InvestmentAccountResponse, error) {
	it, err := s.repo.GetInvestmentAccount(ctx, userID, investmentAccountID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewInvestmentAccountResponse(*it)
	return &resp, nil
}

func (s *InvestmentService) ListInvestmentAccounts(ctx context.Context, userID string) ([]dto.InvestmentAccountResponse, error) {
	items, err := s.repo.ListInvestmentAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewInvestmentAccountResponses(items), nil
}

func (s *InvestmentService) GetSecurity(ctx context.Context, securityID string) (*dto.SecurityResponse, error) {
	it, err := s.repo.GetSecurity(ctx, securityID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewSecurityResponse(*it)
	return &resp, nil
}

func (s *InvestmentService) ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error) {
	items, err := s.repo.ListSecurities(ctx)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityResponses(items), nil
}

func (s *InvestmentService) ListTrades(ctx context.Context, userID, brokerAccountID string) ([]dto.TradeResponse, error) {
	items, err := s.repo.ListTrades(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}
	return dto.NewTradeResponses(items), nil
}

func (s *InvestmentService) ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]dto.HoldingResponse, error) {
	items, err := s.repo.ListHoldings(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}
	return dto.NewHoldingResponses(items), nil
}

func (s *InvestmentService) ListSecurityPrices(ctx context.Context, securityID string, from, to *string) ([]dto.SecurityPriceDailyResponse, error) {
	items, err := s.repo.ListSecurityPrices(ctx, securityID, from, to)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityPriceDailyResponses(items), nil
}

func (s *InvestmentService) ListSecurityEvents(ctx context.Context, securityID string, from, to *string) ([]dto.SecurityEventResponse, error) {
	items, err := s.repo.ListSecurityEvents(ctx, securityID, from, to)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityEventResponses(items), nil
}

func (s *InvestmentService) DeleteTrade(ctx context.Context, userID, investmentAccountID, tradeID string) error {
	bid := strings.TrimSpace(investmentAccountID)
	if bid == "" {
		return errors.New("investmentAccountId is required")
	}

	tr, err := s.repo.GetTrade(ctx, userID, tradeID)
	if err != nil {
		return err
	}
	if tr.BrokerAccountID != bid {
		return errors.New("forbidden: trade does not belong to this account")
	}

	// 1. Logic FIFO Reversal
	if tr.Side == "buy" {
		lots, err := s.repo.ListShareLots(ctx, userID, bid, tr.SecurityID)
		if err != nil {
			return err
		}
		for _, l := range lots {
			if l.BuyTradeID != nil && *l.BuyTradeID == tr.ID {
				if l.Status != "active" || l.Quantity != tr.Quantity {
					return errors.New("cannot delete buy trade because some shares are already sold or modified")
				}
			}
		}
		if err := s.repo.DeleteShareLotsByTradeID(ctx, userID, tr.ID); err != nil {
			return err
		}
	} else {
		logs, err := s.repo.ListRealizedLogsByTradeID(ctx, userID, tr.ID)
		if err != nil {
			return err
		}
		for _, l := range logs {
			lots, err := s.repo.ListShareLots(ctx, userID, bid, tr.SecurityID)
			if err != nil {
				return err
			}
			var targetLot *entity.ShareLot
			for _, lot := range lots {
				if lot.ID == l.SourceShareLot {
					targetLot = &lot
					break
				}
			}
			if targetLot == nil {
				return errors.New("source lot not found for restoration")
			}
			oldQ, _ := new(big.Rat).SetString(targetLot.Quantity)
			soldQ, _ := new(big.Rat).SetString(l.Quantity)
			newQ := new(big.Rat).Add(oldQ, soldQ)
			if err := s.repo.UpdateShareLotQuantity(ctx, userID, targetLot.ID, newQ.FloatString(8)); err != nil {
				return err
			}
		}
		if err := s.repo.DeleteRealizedLogsByTradeID(ctx, userID, tr.ID); err != nil {
			return err
		}
	}

	// 2. Delete Transactions
	if tr.FeeTransactionID != nil {
		_ = s.txSvc.Delete(ctx, userID, *tr.FeeTransactionID)
	}
	if tr.TaxTransactionID != nil {
		_ = s.txSvc.Delete(ctx, userID, *tr.TaxTransactionID)
	}

	// 3. Delete Trade
	if err := s.repo.DeleteTrade(ctx, userID, tr.ID); err != nil {
		return err
	}

	// 4. Update Holding
	return s.upsertHoldingFromLots(ctx, userID, bid, tr.SecurityID)
}

func (s *InvestmentService) UpdateTrade(ctx context.Context, userID, brokerAccountID, tradeID string, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	if err := s.DeleteTrade(ctx, userID, brokerAccountID, tradeID); err != nil {
		return nil, err
	}
	return s.CreateTrade(ctx, userID, brokerAccountID, req)
}

func (s *InvestmentService) CreateTrade(ctx context.Context, userID, brokerAccountID string, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	bid := strings.TrimSpace(brokerAccountID)
	ia, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	occAt, occDate, err := s.normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	provenance := "regular_buy"
	if req.Provenance != nil {
		provenance = *req.Provenance
	}

	side := strings.ToLower(strings.TrimSpace(req.Side))

	// Sell logic FIFO
	var sellPlan []lotConsumptionPlan
	if side == "sell" {
		lots, err := s.repo.ListShareLots(ctx, userID, bid, req.SecurityID)
		if err != nil {
			return nil, err
		}
		plan, _, err := s.planFIFOSell(lots, req.Quantity)
		if err != nil {
			return nil, err
		}
		sellPlan = plan
	}

	// Fees & Taxes (legacy logic simplification for migration)
	fees := "0"
	if req.Fees != nil {
		fees = *req.Fees
	}
	taxes := "0"
	if req.Taxes != nil {
		taxes = *req.Taxes
	}

	tradeID := uuid.NewString()

	// Principal Transaction
	if !(side == "buy" && provenance == "stock_dividend") {
		qRat, _ := new(big.Rat).SetString(req.Quantity)
		pRat, _ := new(big.Rat).SetString(req.Price)
		notional := new(big.Rat).Mul(qRat, pRat)
		if notional.Sign() > 0 {
			amt := notional.FloatString(2)
			kind := "expense"
			if side == "sell" {
				kind = "income"
			}
			desc := "Trade " + side + " " + req.SecurityID
			if req.PrincipalDescription != nil {
				desc = *req.PrincipalDescription
			}

			extRef := "trade:" + tradeID + ":principal"
			catID := "cat_sys_invest_buy"
			if side == "sell" {
				catID = "cat_sys_invest_sell"
			}
			if req.PrincipalCategoryID != nil {
				catID = *req.PrincipalCategoryID
			}

			_, _ = s.txSvc.Create(ctx, userID, dto.CreateTransactionRequest{
				Type:         kind,
				OccurredDate: &occDate,
				Amount:       amt,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  &extRef,
				CategoryID:   &catID,
			})
		}
	}

	// Fee/Tax Transactions
	var feeTxID, taxTxID *string
	if fAmt, _ := new(big.Rat).SetString(fees); fAmt.Sign() > 0 {
		fDesc := "Trade Fee " + req.SecurityID
		fExtRef := "trade:" + tradeID + ":fee"
		fCat := "cat_sys_invest_fees"
		tx, _ := s.txSvc.Create(ctx, userID, dto.CreateTransactionRequest{
			Type: "expense", OccurredDate: &occDate, Amount: fees, Description: &fDesc, AccountID: &ia.AccountID, ExternalRef: &fExtRef, CategoryID: &fCat,
		})
		if tx != nil {
			feeTxID = &tx.ID
		}
	}
	if tAmt, _ := new(big.Rat).SetString(taxes); tAmt.Sign() > 0 {
		tDesc := "Trade Tax " + req.SecurityID
		tExtRef := "trade:" + tradeID + ":tax"
		tCat := "cat_def_other_taxes"
		tx, _ := s.txSvc.Create(ctx, userID, dto.CreateTransactionRequest{
			Type: "expense", OccurredDate: &occDate, Amount: taxes, Description: &tDesc, AccountID: &ia.AccountID, ExternalRef: &tExtRef, CategoryID: &tCat,
		})
		if tx != nil {
			taxTxID = &tx.ID
		}
	}

	trade := entity.Trade{
		ID: tradeID, ClientID: req.ClientID, BrokerAccountID: bid, SecurityID: req.SecurityID,
		FeeTransactionID: feeTxID, TaxTransactionID: taxTxID, Side: side, Quantity: req.Quantity, Price: req.Price,
		Fees: fees, Taxes: taxes, OccurredAt: occAt, Note: req.Note, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateTrade(ctx, userID, trade); err != nil {
		return nil, err
	}

	if side == "buy" {
		lot := entity.ShareLot{
			ID: uuid.NewString(), BrokerAccountID: bid, SecurityID: req.SecurityID, Quantity: req.Quantity,
			AcquisitionDate: occDate, CostBasisPer: req.Price, Provenance: provenance, Status: "active",
			BuyTradeID: &tradeID, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		_ = s.repo.CreateShareLot(ctx, userID, lot)
	} else {
		for _, c := range sellPlan {
			_ = s.repo.UpdateShareLotQuantity(ctx, userID, c.LotID, c.NewQuantity)
			_ = s.repo.CreateRealizedTradeLog(ctx, userID, entity.RealizedTradeLog{
				ID: uuid.NewString(), BrokerAccountID: bid, SecurityID: req.SecurityID, SellTradeID: tradeID,
				SourceShareLot: c.LotID, Quantity: c.SoldQuantity, AcquisitionDate: c.AcquisitionDate,
				CostBasisTotal: c.CostBasisTotal, SellPrice: req.Price, RealizedPnL: c.RealizedPnL, Provenance: c.Provenance, CreatedAt: time.Now().UTC(),
			})
		}
	}

	_ = s.upsertHoldingFromLots(ctx, userID, bid, req.SecurityID)

	resp := dto.NewTradeResponse(trade)
	return &resp, nil
}

func (s *InvestmentService) UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, req dto.PatchInvestmentAccountRequest) (*dto.InvestmentAccountResponse, error) {
	it, err := s.repo.UpdateInvestmentAccountSettings(ctx, userID, investmentAccountID, req.FeeSettings, req.TaxSettings)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewInvestmentAccountResponse(*it)
	return &resp, nil
}

func (s *InvestmentService) ListEligibleCorporateActions(ctx context.Context, userID, brokerAccountID string) ([]dto.EligibleAction, error) {
	// Simple mock or minimal logic for migration
	return []dto.EligibleAction{}, nil
}

func (s *InvestmentService) ClaimCorporateAction(ctx context.Context, userID, brokerAccountID, eventID string, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *InvestmentService) GetRealizedPNLReport(ctx context.Context, userID, brokerAccountID string) (*dto.RealizedPNLReport, error) {
	logs, err := s.repo.ListRealizedLogs(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}

	report := &dto.RealizedPNLReport{Items: []dto.RealizedPNLReportItem{}}
	totalNet := new(big.Rat)

	// Group by security
	bySec := map[string]*dto.RealizedPNLReportItem{}
	for _, l := range logs {
		item, ok := bySec[l.SecurityID]
		if !ok {
			item = &dto.RealizedPNLReportItem{SecurityID: l.SecurityID}
			bySec[l.SecurityID] = item
			report.Items = append(report.Items, *item)
		}
		// ... summation logic ...
		pnl, _ := new(big.Rat).SetString(l.RealizedPnL)
		totalNet.Add(totalNet, pnl)
	}
	report.TotalNet = totalNet.FloatString(2)
	return report, nil
}

func (s *InvestmentService) BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID string) (*dto.BackfillTradePrincipalResponse, error) {
	return &dto.BackfillTradePrincipalResponse{}, nil
}

// Helpers (Internal)

func (s *InvestmentService) normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
	if occurredAt != nil {
		t, err := time.Parse(time.RFC3339, *occurredAt)
		if err == nil {
			return t.UTC(), t.UTC().Format("2006-01-02"), nil
		}
	}
	if occurredDate == nil {
		now := time.Now().UTC()
		return now, now.Format("2006-01-02"), nil
	}
	d, err := time.Parse("2006-01-02", *occurredDate)
	if err != nil {
		return time.Time{}, "", err
	}
	t := d
	if occurredTime != nil {
		tm, _ := time.Parse("15:04", *occurredTime)
		t = time.Date(d.Year(), d.Month(), d.Day(), tm.Hour(), tm.Minute(), 0, 0, time.UTC)
	}
	return t, t.Format("2006-01-02"), nil
}

type lotConsumptionPlan struct {
	LotID           string
	AcquisitionDate string
	Provenance      string
	SoldQuantity    string
	NewQuantity     string
	CostBasisTotal  string
	RealizedPnL     string
}

func (s *InvestmentService) planFIFOSell(lots []entity.ShareLot, sellQty string) ([]lotConsumptionPlan, string, error) {
	toSell, _ := new(big.Rat).SetString(sellQty)
	remaining := new(big.Rat).Set(toSell)
	var plan []lotConsumptionPlan

	for _, l := range lots {
		if remaining.Sign() <= 0 {
			break
		}
		lotQ, _ := new(big.Rat).SetString(l.Quantity)
		if lotQ.Sign() <= 0 {
			continue
		}

		consume := new(big.Rat)
		if lotQ.Cmp(remaining) >= 0 {
			consume.Set(remaining)
			remaining.SetInt64(0)
		} else {
			consume.Set(lotQ)
			remaining.Sub(remaining, lotQ)
		}

		newQ := new(big.Rat).Sub(lotQ, consume)
		costBasisPer, _ := new(big.Rat).SetString(l.CostBasisPer)
		costTotal := new(big.Rat).Mul(consume, costBasisPer)

		plan = append(plan, lotConsumptionPlan{
			LotID: l.ID, AcquisitionDate: l.AcquisitionDate, Provenance: l.Provenance,
			SoldQuantity: consume.FloatString(8), NewQuantity: newQ.FloatString(8),
			CostBasisTotal: costTotal.FloatString(2),
		})
	}

	if remaining.Sign() > 0 {
		return nil, "0", errors.New("insufficient quantity to sell")
	}

	return plan, "0", nil
}

func (s *InvestmentService) upsertHoldingFromLots(ctx context.Context, userID, brokerAccountID, securityID string) error {
	lots, err := s.repo.ListShareLots(ctx, userID, brokerAccountID, securityID)
	if err != nil {
		return err
	}

	totalQ := new(big.Rat)
	totalCost := new(big.Rat)
	for _, l := range lots {
		q, _ := new(big.Rat).SetString(l.Quantity)
		totalQ.Add(totalQ, q)
		cp, _ := new(big.Rat).SetString(l.CostBasisPer)
		totalCost.Add(totalCost, new(big.Rat).Mul(q, cp))
	}

	if totalQ.Sign() == 0 {
		// Possibly delete holding or set to zero
		return nil
	}

	avg := new(big.Rat).Quo(totalCost, totalQ)
	h := entity.Holding{
		ID: uuid.NewString(), BrokerAccountID: brokerAccountID, SecurityID: securityID,
		Quantity: totalQ.FloatString(8), CostBasisTotal: strPtr(totalCost.FloatString(2)),
		AvgCost: strPtr(avg.FloatString(2)), UpdatedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
	}
	_, err = s.repo.UpsertHolding(ctx, userID, h)
	return err
}

func strPtr(s string) *string { return &s }
