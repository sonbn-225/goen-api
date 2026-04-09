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
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

// TradeService owns trade business logic as an independent module.
type TradeService struct {
	repo     interfaces.TradeRepository
	txRepo   interfaces.TransactionRepository
	accRepo  interfaces.AccountRepository
	auditSvc interfaces.AuditService
	db       *database.Postgres
}

func NewTradeService(
	repo interfaces.TradeRepository,
	txRepo interfaces.TransactionRepository,
	accRepo interfaces.AccountRepository,
	auditSvc interfaces.AuditService,
	db *database.Postgres,
) *TradeService {
	return &TradeService{
		repo:     repo,
		txRepo:   txRepo,
		accRepo:  accRepo,
		auditSvc: auditSvc,
		db:       db,
	}
}

func (s *TradeService) ListTrades(ctx context.Context, userID, accountID uuid.UUID) ([]dto.TradeResponse, error) {
	items, err := s.repo.ListTradesTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewTradeResponses(items), nil
}

func (s *TradeService) CreateTrade(ctx context.Context, userID, accountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
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
					catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000001")
				} else {
					catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000002")
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

		var feeTxID, taxTxID *uuid.UUID
		if fAmt, _ := new(big.Rat).SetString(fees); fAmt.Sign() > 0 {
			fDesc := "Trade Fee " + req.SecurityID.String()
			fExtRef := "trade:" + tradeID.String() + ":fee"
			fCat, _ := uuid.Parse("00000000-0000-0000-0000-000000000003")
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
			tCat, _ := uuid.Parse("00000000-0000-0000-0000-000000000004")
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

		trade := entity.Trade{
			AuditEntity:      entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: tradeID}},
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

		if side == string(entity.TradeSideBuy) {
			lot := entity.ShareLot{
				AuditEntity:     entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
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
					AuditEntity:     entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
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

		tr := dto.NewTradeResponse(trade)
		resp = &tr

		_ = s.auditSvc.Record(ctx, tx, userID, &accountID, entity.ResourceTrade, entity.ActionCreated, tradeID, nil, trade)
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.upsertHoldingFromLots(ctx, nil, userID, accountID, sid)
	return resp, nil
}

func (s *TradeService) UpdateTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	if err := s.DeleteTrade(ctx, userID, accountID, tradeID); err != nil {
		return nil, err
	}
	return s.CreateTrade(ctx, userID, accountID, req)
}

func (s *TradeService) DeleteTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID) error {
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

		if err := s.repo.DeleteTradeTx(ctx, tx, userID, tr.ID); err != nil {
			return err
		}

		return s.upsertHoldingFromLots(ctx, tx, userID, accountID, tr.SecurityID)
	})
}

type lotConsumptionPlan struct {
	LotID           uuid.UUID
	AcquisitionDate string
	Provenance      string
	SoldQuantity    string
	NewQuantity     string
	CostBasisTotal  string
	RealizedPnL     string
}

func (s *TradeService) planFIFOSell(lots []entity.ShareLot, sellQty string) ([]lotConsumptionPlan, string, error) {
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

func (s *TradeService) upsertHoldingFromLots(ctx context.Context, tx pgx.Tx, userID, accountID, securityID uuid.UUID) error {
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
		return nil
	}

	avg := new(big.Rat).Quo(totalCost, totalQ)
	h := entity.Holding{
		AuditEntity:    entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
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
