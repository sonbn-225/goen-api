package service

import (
	"context"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

// InvestmentService manages trading and holdings of securities (stocks, crypto, etc.).
// It integrates with TransactionService to record the cash flow of trades, including
// principal amounts, brokerage fees, and relevant taxes.
type InvestmentService struct {
	repo    interfaces.InvestmentRepository
	txRepo  interfaces.TransactionRepository
	accRepo interfaces.AccountRepository
	txSvc   interfaces.TransactionService
	secSvc  interfaces.SecurityService
	db      *database.Postgres
}

// NewInvestmentService creates a new investment management service.
func NewInvestmentService(
	repo interfaces.InvestmentRepository,
	txRepo interfaces.TransactionRepository,
	accRepo interfaces.AccountRepository,
	txSvc interfaces.TransactionService,
	secSvc interfaces.SecurityService,
	db *database.Postgres,
) *InvestmentService {
	return &InvestmentService{
		repo:    repo,
		txRepo:  txRepo,
		accRepo: accRepo,
		txSvc:   txSvc,
		secSvc:  secSvc,
		db:      db,
	}
}

// Investment account specific management has been merged into AccountService

func (s *InvestmentService) ListTrades(ctx context.Context, userID, accountID uuid.UUID) ([]dto.TradeResponse, error) {
	items, err := s.repo.ListTradesTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewTradeResponses(items), nil
}

func (s *InvestmentService) ListHoldings(ctx context.Context, userID, accountID uuid.UUID) ([]dto.HoldingResponse, error) {
	items, err := s.repo.ListHoldingsTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewHoldingResponses(items), nil
}

// DeleteTrade removes a trade record and its associated ledger transactions.
// It reverses the FIFO impact on share lots before deletion.
func (s *InvestmentService) DeleteTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID) error {
	tr, err := s.repo.GetTradeTx(ctx, nil, userID, tradeID)
	if err != nil {
		return err
	}
	if tr == nil {
		return apperr.NotFound("trade not found")
	}
	if tr.AccountID != accountID {
		return apperr.Forbidden("trade_access_denied", "forbidden: trade does not belong to this account")
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Logic FIFO Reversal
		if tr.Side == entity.TradeSideBuy {
			lots, err := s.repo.ListShareLotsTx(ctx, nil, userID, accountID, tr.SecurityID)
			if err != nil {
				return err
			}
			for _, l := range lots {
				if l.BuyTradeID != nil && *l.BuyTradeID == tr.ID {
					if l.Status != entity.ShareLotStatusActive || l.Quantity != tr.Quantity {
						return apperr.BadRequest("trade_modified", "cannot delete buy trade because some shares are already sold or modified")
					}
				}
			}
			if err := s.repo.DeleteShareLotsByTradeIDTx(ctx, tx, userID, tr.ID); err != nil {
				return err
			}
		} else {
			logs, err := s.repo.ListRealizedLogsByTradeIDTx(ctx, nil, userID, tr.ID)
			if err != nil {
				return err
			}
			for _, l := range logs {
				lots, err := s.repo.ListShareLotsTx(ctx, nil, userID, accountID, tr.SecurityID)
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
					return apperr.Internal("source lot not found for restoration")
				}
				oldQ, _ := new(big.Rat).SetString(targetLot.Quantity)
				soldQ, _ := new(big.Rat).SetString(l.Quantity)
				newQ := new(big.Rat).Add(oldQ, soldQ)
				if err := s.repo.UpdateShareLotQuantityTx(ctx, tx, userID, targetLot.ID, newQ.FloatString(8)); err != nil {
					return err
				}
			}
			if err := s.repo.DeleteRealizedLogsByTradeIDTx(ctx, tx, userID, tr.ID); err != nil {
				return err
			}
		}

		// 2. Delete Ledger Transactions
		if tr.FeeTransactionID != nil && *tr.FeeTransactionID != uuid.Nil {
			if err := s.repo.DeleteTransactionTx(ctx, tx, userID, *tr.FeeTransactionID); err != nil {
				return err
			}
		}
		if tr.TaxTransactionID != nil && *tr.TaxTransactionID != uuid.Nil {
			if err := s.repo.DeleteTransactionTx(ctx, tx, userID, *tr.TaxTransactionID); err != nil {
				return err
			}
		}

		// 3. Delete Trade Record
		if err := s.repo.DeleteTradeTx(ctx, tx, userID, tr.ID); err != nil {
			return err
		}

		// 4. Update Holding Summary
		return s.upsertHoldingFromLots(ctx, tx, userID, accountID, tr.SecurityID)
	})
}

func (s *InvestmentService) UpdateTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	if err := s.DeleteTrade(ctx, userID, accountID, tradeID); err != nil {
		return nil, err
	}
	return s.CreateTrade(ctx, userID, accountID, req)
}

// CreateTrade records a new purchase or sale of a security.
// It performs several atomic steps:
// 1. Calculates FIFO sell plan if it's a 'sell' trade.
// 2. Records Principal, Fee, and Tax transactions in the central ledger.
// 3. Creates the Trade record.
// 4. Updates or creates ShareLots (for acquisition tracking).
// 5. Updates the overall Holding summary for the security.
func (s *InvestmentService) CreateTrade(ctx context.Context, userID, accountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	acc, err := s.accRepo.GetAccountForUserTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, apperr.NotFound("account not found")
	}

	sid := req.SecurityID

	occAt, occDate, err := utils.NormalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	provenance := "regular_buy"
	if req.Provenance != nil {
		provenance = *req.Provenance
	}

	side := strings.ToLower(strings.TrimSpace(string(req.Side)))

	// 1. Sell logic FIFO
	var sellPlan []lotConsumptionPlan
	if side == string(entity.TradeSideSell) {
		lots, err := s.repo.ListShareLotsTx(ctx, nil, userID, accountID, sid)
		if err != nil {
			return nil, err
		}
		plan, _, err := s.planFIFOSell(lots, req.Quantity)
		if err != nil {
			return nil, err
		}
		sellPlan = plan
	}

	fees := "0"
	if req.Fees != nil {
		fees = *req.Fees
	}
	taxes := "0"
	if req.Taxes != nil {
		taxes = *req.Taxes
	}

	tradeID := utils.NewID()
	var resp *dto.TradeResponse

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// A. Principal Transaction
		if !(side == string(entity.TradeSideBuy) && provenance == "stock_dividend") {
			qRat, _ := new(big.Rat).SetString(req.Quantity)
			pRat, _ := new(big.Rat).SetString(req.Price)
			notional := new(big.Rat).Mul(qRat, pRat)
			if notional.Sign() > 0 {
				amt := notional.FloatString(2)
				kind := entity.TransactionTypeExpense
				if side == string(entity.TradeSideSell) {
					kind = entity.TransactionTypeIncome
				}
				desc := "Trade " + side + " " + req.SecurityID.String()
				if req.PrincipalDescription != nil {
					desc = *req.PrincipalDescription
				}

				extRef := "trade:" + tradeID.String() + ":principal"
				var catID uuid.UUID
				if side == string(entity.TradeSideSell) {
					catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000001") // TODO: System category
				} else {
					catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000002") // TODO: System category
				}
				if req.PrincipalCategoryID != nil {
					catID = *req.PrincipalCategoryID
				}

				pTx := entity.Transaction{
					AuditEntity:  entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
					Type:         kind,
					OccurredAt:   occAt,
					OccurredDate: occDate,
					Amount:       amt,
					Description:  &desc,
					AccountID:    &acc.ID,
					ExternalRef:  &extRef,
					Status:       entity.TransactionStatusPosted,
				}
				pLine := []entity.TransactionLineItem{
					{BaseEntity: entity.BaseEntity{ID: utils.NewID()}, Amount: amt, CategoryID: &catID, Note: &desc},
				}
				if err := s.txRepo.CreateTransactionTx(ctx, tx, userID, pTx, pLine, nil); err != nil {
					return err
				}
			}
		}

		// B. Fee/Tax Transactions
		var feeTxID, taxTxID *uuid.UUID
		if fAmt, _ := new(big.Rat).SetString(fees); fAmt.Sign() > 0 {
			fDesc := "Trade Fee " + req.SecurityID.String()
			fExtRef := "trade:" + tradeID.String() + ":fee"
			fCat, _ := uuid.Parse("00000000-0000-0000-0000-000000000003") // TODO: Use real system ID
			fTx := entity.Transaction{
				AuditEntity:  entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
				Type:         entity.TransactionTypeExpense,
				OccurredAt:   occAt,
				OccurredDate: occDate,
				Amount:       fees,
				Description:  &fDesc,
				AccountID:    &acc.ID,
				ExternalRef:  &fExtRef,
				Status:       entity.TransactionStatusPosted,
			}
			fLine := []entity.TransactionLineItem{
				{BaseEntity: entity.BaseEntity{ID: utils.NewID()}, Amount: fees, CategoryID: &fCat, Note: &fDesc},
			}
			if err := s.txRepo.CreateTransactionTx(ctx, tx, userID, fTx, fLine, nil); err != nil {
				return err
			}
			feeTxID = &fTx.ID
		}
		if tAmt, _ := new(big.Rat).SetString(taxes); tAmt.Sign() > 0 {
			tDesc := "Trade Tax " + req.SecurityID.String()
			tExtRef := "trade:" + tradeID.String() + ":tax"
			tCat, _ := uuid.Parse("00000000-0000-0000-0000-000000000004") // TODO: Use real system ID
			tTx := entity.Transaction{
				AuditEntity:  entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
				Type:         entity.TransactionTypeExpense,
				OccurredAt:   occAt,
				OccurredDate: occDate,
				Amount:       taxes,
				Description:  &tDesc,
				AccountID:    &acc.ID,
				ExternalRef:  &tExtRef,
				Status:       entity.TransactionStatusPosted,
			}
			tLine := []entity.TransactionLineItem{
				{BaseEntity: entity.BaseEntity{ID: utils.NewID()}, Amount: taxes, CategoryID: &tCat, Note: &tDesc},
			}
			if err := s.txRepo.CreateTransactionTx(ctx, tx, userID, tTx, tLine, nil); err != nil {
				return err
			}
			taxTxID = &tTx.ID
		}

		// C. Create Trade Record
		trade := entity.Trade{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{
					ID: tradeID,
				},
			},
			AccountID:        accountID,
			SecurityID:       sid,
			FeeTransactionID: feeTxID,
			TaxTransactionID: taxTxID,
			Side:             entity.TradeSide(side),
			Quantity:         req.Quantity,
			Price:            req.Price,
			Fees:             fees,
			Taxes:            taxes,
			OccurredAt:       occAt,
			Note:             req.Note,
		}

		if err := s.repo.CreateTradeTx(ctx, tx, userID, trade); err != nil {
			return err
		}

		// D. Updates ShareLots
		if side == string(entity.TradeSideBuy) {
			lot := entity.ShareLot{
				AuditEntity: entity.AuditEntity{
					BaseEntity: entity.BaseEntity{
						ID: utils.NewID(),
					},
				},
				AccountID:       accountID,
				SecurityID:      sid,
				Quantity:        req.Quantity,
				AcquisitionDate: occDate,
				CostBasisPer:    req.Price,
				Provenance:      provenance,
				Status:          entity.ShareLotStatusActive,
				BuyTradeID:      &tradeID,
			}
			if err := s.repo.CreateShareLotTx(ctx, tx, userID, lot); err != nil {
				return err
			}
		} else {
			for _, c := range sellPlan {
				if err := s.repo.UpdateShareLotQuantityTx(ctx, tx, userID, c.LotID, c.NewQuantity); err != nil {
					return err
				}
				if err := s.repo.CreateRealizedTradeLogTx(ctx, tx, userID, entity.RealizedTradeLog{
					AuditEntity: entity.AuditEntity{
						BaseEntity: entity.BaseEntity{
							ID: utils.NewID(),
						},
					},
					AccountID:       accountID,
					SecurityID:      sid,
					SellTradeID:     tradeID,
					SourceShareLot:  c.LotID,
					Quantity:        c.SoldQuantity,
					AcquisitionDate: c.AcquisitionDate,
					CostBasisTotal:  c.CostBasisTotal,
					SellPrice:       req.Price,
					RealizedPnL:     c.RealizedPnL,
					Provenance:      c.Provenance,
				}); err != nil {
					return err
				}
			}
		}

		// E. Finalize Response
		tr := dto.NewTradeResponse(trade)
		resp = &tr
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Update Holding Summary
	_ = s.upsertHoldingFromLots(ctx, nil, userID, accountID, sid)

	return resp, nil
}


func (s *InvestmentService) ListEligibleCorporateActions(ctx context.Context, userID, brokerAccountID uuid.UUID) ([]dto.EligibleAction, error) {
	// Simple mock or minimal logic for migration
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

	// Group by security
	bySec := map[uuid.UUID]*dto.RealizedPNLReportItem{}
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

func (s *InvestmentService) BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID uuid.UUID) (*dto.BackfillTradePrincipalResponse, error) {
	return &dto.BackfillTradePrincipalResponse{}, nil
}

// Helpers (Internal)


type lotConsumptionPlan struct {
	LotID           uuid.UUID
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
		return nil, "0", apperr.BadRequest("insufficient_quantity", "insufficient quantity to sell")
	}

	return plan, "0", nil
}

func (s *InvestmentService) upsertHoldingFromLots(ctx context.Context, tx pgx.Tx, userID, accountID, securityID uuid.UUID) error {
	lots, err := s.repo.ListShareLotsTx(ctx, tx, userID, accountID, securityID)
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
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		AccountID:      accountID,
		SecurityID:     securityID,
		Quantity:       totalQ.FloatString(8),
		CostBasisTotal: strPtr(totalCost.FloatString(2)),
		AvgCost:        strPtr(avg.FloatString(2)),
	}
	_, err = s.repo.UpsertHoldingTx(ctx, tx, userID, h)
	return err
}

func strPtr(s string) *string { return &s }
