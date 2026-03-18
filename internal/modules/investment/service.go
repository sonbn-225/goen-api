package investment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/i18n"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/storage"
)

type PatchInvestmentAccountRequest struct {
	FeeSettings any `json:"fee_settings,omitempty"`
	TaxSettings any `json:"tax_settings,omitempty"`
}

type CreateTradeRequest struct {
	ClientID         *string `json:"client_id,omitempty"`
	SecurityID       string  `json:"security_id"`
	FeeTransactionID *string `json:"fee_transaction_id,omitempty"`
	TaxTransactionID *string `json:"tax_transaction_id,omitempty"`
	Provenance       *string `json:"provenance,omitempty"`
	Side             string  `json:"side"`
	Quantity         string  `json:"quantity"`
	Price            string  `json:"price"`
	Fees             *string `json:"fees,omitempty"`
	Taxes            *string `json:"taxes,omitempty"`
	OccurredAt       *string `json:"occurred_at,omitempty"`
	OccurredDate     *string `json:"occurred_date,omitempty"`
	OccurredTime     *string `json:"occurred_time,omitempty"`
	Note             *string `json:"note,omitempty"`

	PrincipalCategoryID  *string `json:"principal_category_id,omitempty"`
	PrincipalDescription *string `json:"principal_description,omitempty"`
	FeeCategoryID        *string `json:"fee_category_id,omitempty"`
	FeeDescription       *string `json:"fee_description,omitempty"`
	TaxCategoryID        *string `json:"tax_category_id,omitempty"`
	TaxDescription       *string `json:"tax_description,omitempty"`
}

// Service handles investment business logic.
type Service struct {
	repo       domain.InvestmentRepository
	accountSvc AccountServiceDep
	txSvc      TransactionServiceDep
	cfg        *config.Config
	redis      *storage.Redis
}

type EligibleAction struct {
	Event            domain.SecurityEvent `json:"event"`
	HoldingQuantity  string               `json:"holding_quantity"`
	EntitledQuantity string               `json:"entitled_quantity"`
	Status           string               `json:"status"` // 'eligible', 'claimed', 'dismissed'
	ElectionID       *string              `json:"election_id,omitempty"`
}

type ClaimCorporateActionRequest struct {
	ElectedQuantity *string `json:"elected_quantity,omitempty"`
	Note            *string `json:"note,omitempty"`
}

// NewService creates a new investment service.
func NewService(
	repo domain.InvestmentRepository,
	accountSvc AccountServiceDep,
	txSvc TransactionServiceDep,
	cfg *config.Config,
	redis *storage.Redis,
) *Service {
	return &Service{
		repo:       repo,
		accountSvc: accountSvc,
		txSvc:      txSvc,
		cfg:        cfg,
		redis:      redis,
	}
}

// GetInvestmentAccount retrieves an investment account by ID.
func (s *Service) GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*domain.InvestmentAccount, error) {
	return s.repo.GetInvestmentAccount(ctx, userID, investmentAccountID)
}

// ListInvestmentAccounts lists all investment accounts for a user.
func (s *Service) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	return s.repo.ListInvestmentAccounts(ctx, userID)
}

// GetSecurity retrieves a security by ID.
func (s *Service) GetSecurity(ctx context.Context, securityID string) (*domain.Security, error) {
	return s.repo.GetSecurity(ctx, securityID)
}

// ListSecurities lists all securities.
func (s *Service) ListSecurities(ctx context.Context) ([]domain.Security, error) {
	return s.repo.ListSecurities(ctx)
}

// ListTrades lists trades for an investment account.
func (s *Service) ListTrades(ctx context.Context, userID, brokerAccountID string) ([]domain.Trade, error) {
	return s.repo.ListTrades(ctx, userID, brokerAccountID)
}

// ListHoldings lists holdings for an investment account.
func (s *Service) ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]domain.Holding, error) {
	return s.repo.ListHoldings(ctx, userID, brokerAccountID)
}

// ListSecurityPrices lists price history for a security.
func (s *Service) ListSecurityPrices(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityPriceDaily, error) {
	return s.repo.ListSecurityPrices(ctx, securityID, from, to)
}

// ListSecurityEvents lists events for a security.
func (s *Service) ListSecurityEvents(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityEvent, error) {
	return s.repo.ListSecurityEvents(ctx, securityID, from, to)
}

func (s *Service) GetTrade(ctx context.Context, userID, tradeID string) (*domain.Trade, error) {
	return s.repo.GetTrade(ctx, userID, tradeID)
}

func (s *Service) DeleteTrade(ctx context.Context, userID, investmentAccountID, tradeID string) error {
	bid := strings.TrimSpace(investmentAccountID)
	if bid == "" {
		return apperrors.Validation("investmentAccountId is required", nil)
	}

	tr, err := s.repo.GetTrade(ctx, userID, tradeID)
	if err != nil {
		return err
	}
	if tr.BrokerAccountID != bid {
		return apperrors.ErrInvestmentForbidden
	}

	// 1. If it's a Buy trade, check if share lots are still 'active' and quantity matches.
	if tr.Side == "buy" {
		lots, err := s.repo.ListShareLots(ctx, userID, bid, tr.SecurityID)
		if err != nil {
			return err
		}
		for _, l := range lots {
			if l.BuyTradeID != nil && *l.BuyTradeID == tr.ID {
				if l.Status != "active" || l.Quantity != tr.Quantity {
					return apperrors.Validation("cannot delete buy trade because some of its shares have already been sold or modified", nil)
				}
			}
		}
		if err := s.repo.DeleteShareLotsByTradeID(ctx, userID, tr.ID); err != nil {
			return err
		}
	} else {
		// 2. If it's a Sell trade, restore quantity to original lots.
		logs, err := s.repo.ListRealizedLogsByTradeID(ctx, userID, tr.ID)
		if err != nil {
			return err
		}
		for _, l := range logs {
			lots, err := s.repo.ListShareLots(ctx, userID, bid, tr.SecurityID)
			if err != nil {
				return err
			}
			var targetLot *domain.ShareLot
			for _, lot := range lots {
				if lot.ID == l.SourceShareLot {
					targetLot = &lot
					break
				}
			}
			if targetLot == nil {
				return apperrors.Validation("source share lot not found for sell trade restoration", nil)
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

	// 3. Delete associated transactions (Fee & Tax).
	if tr.FeeTransactionID != nil {
		_ = s.txSvc.Delete(ctx, userID, *tr.FeeTransactionID)
	}
	if tr.TaxTransactionID != nil {
		_ = s.txSvc.Delete(ctx, userID, *tr.TaxTransactionID)
	}

	// 4. Principal transaction by Deterministic ExternalRef.
	// Since we don't have transaction ID for principal, we rely on the Deterministic ExternalRef
	// which is trade:{clientID}:{tradeID}:principal or trade:{tradeID}:principal.
	// However, without a Search/DeleteByExternalRef method in txSvc, we might leave it orphaned
	// until we implement a better cleanup.
	// TODO: Add DeleteByExternalRef to transaction module or store principal_transaction_id in trades table.

	// 5. Delete the trade record.
	if err := s.repo.DeleteTrade(ctx, userID, tr.ID); err != nil {
		return err
	}

	// 6. Refresh holding.
	return s.upsertHoldingFromLots(ctx, userID, bid, tr.SecurityID)
}

func (s *Service) UpdateTrade(ctx context.Context, userID, investmentAccountID, tradeID string, req CreateTradeRequest) (*domain.Trade, error) {
	// Revert-and-Apply: delete old trade effects and create new ones.
	if err := s.DeleteTrade(ctx, userID, investmentAccountID, tradeID); err != nil {
		return nil, err
	}

	// Create new trade. Note: we might want to preserve the old trade ID if possible,
	// but CreateTrade returns a new ID. For now, this is the safest path.
	return s.CreateTrade(ctx, userID, investmentAccountID, req)
}

func (s *Service) UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, req PatchInvestmentAccountRequest) (*domain.InvestmentAccount, error) {
	id := strings.TrimSpace(investmentAccountID)
	if id == "" {
		return nil, apperrors.Validation("investmentAccountId is required", nil)
	}

	if err := validateChargeSettings(req.FeeSettings, "fee_settings"); err != nil {
		return nil, err
	}
	if err := validateChargeSettings(req.TaxSettings, "tax_settings"); err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateInvestmentAccountSettings(ctx, userID, id, req.FeeSettings, req.TaxSettings)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) CreateTrade(ctx context.Context, userID, brokerAccountID string, req CreateTradeRequest) (*domain.Trade, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, apperrors.Validation("investmentAccountId is required", nil)
	}

	ia, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	lang := httpx.LangFromContext(ctx)
	securityID := strings.TrimSpace(req.SecurityID)
	if securityID == "" {
		return nil, apperrors.Validation("security_id is required", nil)
	}
	if _, err := s.repo.GetSecurity(ctx, securityID); err != nil {
		return nil, err
	}

	side := strings.TrimSpace(req.Side)
	if side != "buy" && side != "sell" {
		return nil, apperrors.Validation("side is invalid", nil)
	}

	quantity := strings.TrimSpace(req.Quantity)
	if quantity == "" {
		return nil, apperrors.Validation("quantity is required", nil)
	}
	if !isValidDecimal(quantity) {
		return nil, apperrors.Validation("quantity must be a decimal string", nil)
	}

	price := strings.TrimSpace(req.Price)
	if price == "" {
		return nil, apperrors.Validation("price is required", nil)
	}
	if !isValidDecimal(price) {
		return nil, apperrors.Validation("price must be a decimal string", nil)
	}

	provenance := normalizeLotProvenance(req.Provenance)
	if provenance == "" {
		provenance = "regular_buy"
	}
	if provenance != "regular_buy" && provenance != "stock_dividend" && provenance != "rights_offering" {
		return nil, apperrors.Validation("provenance is invalid", map[string]any{"field": "provenance"})
	}

	// For sell trades, we may need FIFO lot allocation to compute stock-dividend tax.
	var sellPlan []lotConsumptionPlan
	dividendQtyStr := "0"
	if side == "sell" {
		lots, err := s.repo.ListShareLots(ctx, userID, bid, securityID)
		if err != nil {
			return nil, err
		}

		// Backward-compat: if there are no lots yet but a holding exists, seed a synthetic lot
		// so FIFO can work for existing portfolios.
		if len(lots) == 0 {
			if h, err := s.repo.GetHolding(ctx, userID, bid, securityID); err == nil && h != nil {
				if strings.TrimSpace(h.Quantity) != "" {
					lotQty, ok := new(big.Rat).SetString(strings.TrimSpace(h.Quantity))
					if ok && lotQty.Cmp(big.NewRat(0, 1)) > 0 {
						costPer := "0"
						if h.AvgCost != nil && strings.TrimSpace(*h.AvgCost) != "" {
							costPer = strings.TrimSpace(*h.AvgCost)
						}
						seedID := uuid.NewString()
						seed := domain.ShareLot{
							ID:              seedID,
							BrokerAccountID: bid,
							SecurityID:      securityID,
							Quantity:        lotQty.FloatString(8),
							AcquisitionDate: time.Now().UTC().Format("2006-01-02"),
							CostBasisPer:    costPer,
							Provenance:      "regular_buy",
							Status:          "active",
							BuyTradeID:      nil,
							CreatedAt:       time.Now().UTC(),
							UpdatedAt:       time.Now().UTC(),
						}
						if err := s.repo.CreateShareLot(ctx, userID, seed); err != nil {
							return nil, err
						}
						lots = append(lots, seed)
					}
				}
			}
		}

		plan, divQty, err := planFIFOSell(lots, quantity)
		if err != nil {
			return nil, err
		}
		sellPlan = plan
		dividendQtyStr = divQty
	}

	fees := normalizeDecimalOrZero(req.Fees)
	if fees == "" {
		// Mirror goen: rights offering buys typically do not incur transaction fees.
		if side == "buy" && provenance == "rights_offering" {
			fees = "0"
		} else {
			computed, err := computeChargesFromSettings(ia.FeeSettings, tradeChargeContext{
				Side:             side,
				SecurityID:       securityID,
				Quantity:         quantity,
				Price:            price,
				DividendQuantity: dividendQtyStr,
				Allocations:      nil,
			})
			if err != nil {
				return nil, err
			}
			fees = computed
		}
	}
	if fees == "" {
		fees = "0"
	}
	if !isValidDecimal(fees) {
		return nil, apperrors.Validation("fees must be a decimal string", nil)
	}

	taxes := normalizeDecimalOrZero(req.Taxes)
	if taxes == "" {
		computed, err := computeChargesFromSettings(ia.TaxSettings, tradeChargeContext{
			Side:             side,
			SecurityID:       securityID,
			Quantity:         quantity,
			Price:            price,
			DividendQuantity: dividendQtyStr,
			Allocations:      nil,
		})
		if err != nil {
			return nil, err
		}
		taxes = computed
	}
	if taxes == "" {
		taxes = "0"
	}
	if !isValidDecimal(taxes) {
		return nil, apperrors.Validation("taxes must be a decimal string", nil)
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tradeID := uuid.NewString()

	feeTxID := normalizeOptionalString(req.FeeTransactionID)
	taxTxID := normalizeOptionalString(req.TaxTransactionID)

	principalCategoryID := req.PrincipalCategoryID
	if principalCategoryID == nil || strings.TrimSpace(*principalCategoryID) == "" {
		defaultPrincipalCategory := "cat_def_financial_invest_buy"
		if side == "sell" {
			defaultPrincipalCategory = "cat_def_financial_invest_sell"
		}
		principalCategoryID = &defaultPrincipalCategory
	}

	feeCategoryID := req.FeeCategoryID
	if feeCategoryID == nil || strings.TrimSpace(*feeCategoryID) == "" {
		defaultFeeCategory := "cat_def_financial_invest_fees"
		feeCategoryID = &defaultFeeCategory
	}

	taxCategoryID := req.TaxCategoryID
	if taxCategoryID == nil || strings.TrimSpace(*taxCategoryID) == "" {
		defaultTaxCategory := "cat_def_other_taxes"
		taxCategoryID = &defaultTaxCategory
	}

	// Auto-create principal cashflow transaction (buy/sell notional) so account balances reflect trades,
	// not only fee/tax transactions.
	// - BUY: creates an expense (cash outflow)
	// - SELL: creates an income (cash inflow)
	// For stock dividends, the user should record price=0 if there is no cashflow; additionally, we skip
	// principal cashflow for new trades explicitly marked as stock_dividend provenance.
	if !(side == "buy" && provenance == "stock_dividend") {
		qRat, qOK := new(big.Rat).SetString(strings.TrimSpace(quantity))
		pRat, pOK := new(big.Rat).SetString(strings.TrimSpace(price))
		if qOK && pOK {
			notional := new(big.Rat).Mul(qRat, pRat)
			if notional.Cmp(big.NewRat(0, 1)) > 0 {
				principalAmount := formatRatDecimalScale(notional, 2)
				externalRef := deriveTradeExternalRef(req.ClientID, tradeID, "principal")
				kind := "expense"
				if side == "sell" {
					kind = "income"
				}
				desc := ""
				if req.PrincipalDescription != nil {
					desc = *req.PrincipalDescription
				}
				if desc == "" {
					// Fallback if frontend didn't provide it
					desc = i18n.T(lang, "trade_"+side)
					if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil && sec != nil {
						desc += ": " + sec.Symbol
					}
				}

				_, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
					ClientID:     req.ClientID,
					Type:         kind,
					OccurredDate: &occurredDate,
					Amount:       principalAmount,
					Description:  &desc,
					AccountID:    &ia.AccountID,
					ExternalRef:  externalRef,
					CategoryID:   principalCategoryID,
				})
				if err != nil {
					// Allow idempotent retries/backfills.
					if !isUniqueViolation(err) {
						return nil, err
					}
				}
			}
		}
	}

	// Auto-create fee/tax transactions if requested amounts > 0 and no explicit transaction ids provided.
	if feeTxID == nil {
		if amt, ok := new(big.Rat).SetString(fees); ok && amt.Cmp(new(big.Rat)) > 0 {
			desc := ""
			if req.FeeDescription != nil {
				desc = *req.FeeDescription
			}
			if desc == "" {
				desc = i18n.T(lang, "trade_fee")
				if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil {
					desc += ": " + sec.Symbol
				}
			}

			externalRef := deriveTradeExternalRef(req.ClientID, tradeID, "fee")
			tx, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
				Type:         "expense",
				OccurredDate: &occurredDate,
				Amount:       fees,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  externalRef,
				CategoryID:   feeCategoryID,
			})
			if err != nil {
				return nil, err
			}
			feeTxID = &tx.ID
		}
	}

	if taxTxID == nil {
		if amt, ok := new(big.Rat).SetString(taxes); ok && amt.Cmp(new(big.Rat)) > 0 {
			desc := ""
			if req.TaxDescription != nil {
				desc = *req.TaxDescription
			}
			if desc == "" {
				desc = i18n.T(lang, "trade_tax")
				if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil {
					desc += ": " + sec.Symbol
				}
			}

			externalRef := deriveTradeExternalRef(req.ClientID, tradeID, "tax")
			tx, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
				Type:         "expense",
				OccurredDate: &occurredDate,
				Amount:       taxes,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  externalRef,
				CategoryID:   taxCategoryID,
			})
			if err != nil {
				return nil, err
			}
			taxTxID = &tx.ID
		}
	}

	trade := domain.Trade{
		ID:               tradeID,
		ClientID:         normalizeOptionalString(req.ClientID),
		BrokerAccountID:  bid,
		SecurityID:       securityID,
		FeeTransactionID: feeTxID,
		TaxTransactionID: taxTxID,
		Side:             side,
		Quantity:         quantity,
		Price:            price,
		Fees:             fees,
		Taxes:            taxes,
		OccurredAt:       occurredAt,
		Note:             normalizeOptionalString(req.Note),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateTrade(ctx, userID, trade); err != nil {
		return nil, err
	}

	// Apply share lots + recompute holding.
	if side == "buy" {
		acqDate := addBusinessDays(occurredAt.UTC().Truncate(24*time.Hour), 2).Format("2006-01-02")
		lot := domain.ShareLot{
			ID:              uuid.NewString(),
			BrokerAccountID: bid,
			SecurityID:      securityID,
			Quantity:        quantity,
			AcquisitionDate: acqDate,
			CostBasisPer:    price,
			Provenance:      provenance,
			Status:          "active",
			BuyTradeID:      &tradeID,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}
		if err := s.repo.CreateShareLot(ctx, userID, lot); err != nil {
			return nil, err
		}
	} else {
		for _, c := range sellPlan {
			soldQ, ok := new(big.Rat).SetString(strings.TrimSpace(c.SoldQuantity))
			if !ok {
				return nil, apperrors.Validation("sold quantity is invalid", map[string]any{"field": "quantity"})
			}
			sellP, ok := new(big.Rat).SetString(strings.TrimSpace(price))
			if !ok {
				return nil, apperrors.Validation("price must be a decimal string", map[string]any{"field": "price"})
			}
			proceeds := new(big.Rat).Mul(soldQ, sellP)
			costTotal, ok := new(big.Rat).SetString(strings.TrimSpace(c.CostBasisTotal))
			if !ok {
				costTotal = big.NewRat(0, 1)
			}
			pnl := new(big.Rat).Sub(proceeds, costTotal)
			proceedsStr := formatRatDecimalScale(proceeds, 2)
			pnlStr := formatRatDecimalScale(pnl, 2)

			if err := s.repo.UpdateShareLotQuantity(ctx, userID, c.LotID, c.NewQuantity); err != nil {
				return nil, err
			}
			if err := s.repo.CreateRealizedTradeLog(ctx, userID, domain.RealizedTradeLog{
				ID:              uuid.NewString(),
				BrokerAccountID: bid,
				SecurityID:      securityID,
				SellTradeID:     tradeID,
				SourceShareLot:  c.LotID,
				Quantity:        c.SoldQuantity,
				AcquisitionDate: c.AcquisitionDate,
				CostBasisTotal:  c.CostBasisTotal,
				SellPrice:       price,
				Proceeds:        proceedsStr,
				RealizedPnL:     pnlStr,
				Provenance:      c.Provenance,
				CreatedAt:       time.Now().UTC(),
			}); err != nil {
				return nil, err
			}
		}
	}

	if err := s.upsertHoldingFromLots(ctx, userID, bid, securityID); err != nil {
		return nil, err
	}

	return &trade, nil
}

type BackfillTradePrincipalResult struct {
	TradesTotal          int `json:"trades_total"`
	TransactionsCreated  int `json:"transactions_created"`
	TransactionsExisting int `json:"transactions_existing"`
	SkippedZeroNotional  int `json:"skipped_zero_notional"`
	SkippedStockDividend int `json:"skipped_stock_dividend"`
}

// BackfillTradePrincipalTransactions creates missing principal cashflow transactions for historical trades.
// It is safe to call multiple times; duplicates are ignored via unique (account_id, external_ref).
func (s *Service) BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID string) (*BackfillTradePrincipalResult, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, apperrors.Validation("investmentAccountId is required", nil)
	}

	ia, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		return nil, err
	}
	if ia == nil {
		return nil, apperrors.NotFound("investment account not found", nil)
	}

	trades, err := s.repo.ListTrades(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	// Cache lots by security to detect stock-dividend buys (no cashflow).
	lotsBySecurity := map[string][]domain.ShareLot{}
	securitySymbol := map[string]string{}

	result := &BackfillTradePrincipalResult{TradesTotal: len(trades)}

	for _, tr := range trades {
		securityID := strings.TrimSpace(tr.SecurityID)
		if securityID == "" {
			continue
		}

		// Heuristic: if there is a share lot created by this buy trade and its provenance is stock_dividend,
		// skip principal cashflow.
		if strings.TrimSpace(tr.Side) == "buy" {
			lots, ok := lotsBySecurity[securityID]
			if !ok {
				ll, err := s.repo.ListShareLots(ctx, userID, bid, securityID)
				if err != nil {
					return nil, err
				}
				lots = ll
				lotsBySecurity[securityID] = ll
			}
			isStockDiv := false
			for _, l := range lots {
				if l.BuyTradeID != nil && strings.TrimSpace(*l.BuyTradeID) == strings.TrimSpace(tr.ID) {
					if strings.TrimSpace(l.Provenance) == "stock_dividend" {
						isStockDiv = true
						break
					}
				}
			}
			if isStockDiv {
				result.SkippedStockDividend++
				continue
			}
		}

		qRat, qOK := new(big.Rat).SetString(strings.TrimSpace(tr.Quantity))
		pRat, pOK := new(big.Rat).SetString(strings.TrimSpace(tr.Price))
		if !qOK || !pOK {
			continue
		}
		notional := new(big.Rat).Mul(qRat, pRat)
		if notional.Cmp(big.NewRat(0, 1)) <= 0 {
			result.SkippedZeroNotional++
			continue
		}

		principalAmount := formatRatDecimalScale(notional, 2)
		externalRef := deriveTradeExternalRef(tr.ClientID, tr.ID, "principal")
		kind := "expense"
		if strings.TrimSpace(tr.Side) == "sell" {
			kind = "income"
		}

		// Resolve symbol for nicer description (cached).
		sym := securitySymbol[securityID]
		if sym == "" {
			if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil && sec != nil {
				sym = sec.Symbol
				securitySymbol[securityID] = sym
			}
		}
		lang := httpx.LangFromContext(ctx)
		desc := i18n.T(lang, "trade_"+strings.TrimSpace(tr.Side))
		if sym != "" {
			desc = desc + ": " + sym
		}

		occurredAt := tr.OccurredAt.UTC().Format(time.RFC3339Nano)
		_, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
			ClientID:    tr.ClientID,
			Type:        kind,
			OccurredAt:  &occurredAt,
			Amount:      principalAmount,
			Description: &desc,
			AccountID:   &ia.AccountID,
			ExternalRef: externalRef,
			CategoryID: func() *string {
				cid := "cat_def_financial_invest_buy"
				if strings.TrimSpace(tr.Side) == "sell" {
					cid = "cat_def_financial_invest_sell"
				}
				return &cid
			}(),
		})
		if err != nil {
			if isUniqueViolation(err) {
				result.TransactionsExisting++
				continue
			}
			return nil, err
		}
		result.TransactionsCreated++
	}

	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type lotConsumptionPlan struct {
	LotID           string
	AcquisitionDate string
	Provenance      string
	SoldQuantity    string
	NewQuantity     string
	CostBasisTotal  string
	Proceeds        string
	RealizedPnL     string
}

func planFIFOSell(lots []domain.ShareLot, sellQty string) ([]lotConsumptionPlan, string, error) {
	toSell, ok := new(big.Rat).SetString(strings.TrimSpace(sellQty))
	if !ok {
		return nil, "0", apperrors.Validation("quantity must be a decimal string", map[string]any{"field": "quantity"})
	}
	if toSell.Cmp(big.NewRat(0, 1)) <= 0 {
		return nil, "0", apperrors.Validation("quantity must be > 0", map[string]any{"field": "quantity"})
	}

	available := big.NewRat(0, 1)
	for _, l := range lots {
		if strings.TrimSpace(l.Status) != "active" {
			continue
		}
		q, ok := new(big.Rat).SetString(strings.TrimSpace(l.Quantity))
		if ok && q.Cmp(big.NewRat(0, 1)) > 0 {
			available.Add(available, q)
		}
	}
	if available.Cmp(toSell) < 0 {
		return nil, "0", apperrors.Validation("not enough shares to sell", map[string]any{"field": "quantity"})
	}

	dividendSold := big.NewRat(0, 1)
	plan := []lotConsumptionPlan{}
	remaining := new(big.Rat).Set(toSell)
	for _, l := range lots {
		if remaining.Cmp(big.NewRat(0, 1)) <= 0 {
			break
		}
		if strings.TrimSpace(l.Status) != "active" {
			continue
		}
		lotQty, ok := new(big.Rat).SetString(strings.TrimSpace(l.Quantity))
		if !ok || lotQty.Cmp(big.NewRat(0, 1)) <= 0 {
			continue
		}
		sold := new(big.Rat)
		if lotQty.Cmp(remaining) <= 0 {
			sold.Set(lotQty)
		} else {
			sold.Set(remaining)
		}
		newQty := new(big.Rat).Sub(lotQty, sold)

		costPer := big.NewRat(0, 1)
		if strings.TrimSpace(l.CostBasisPer) != "" {
			if v, ok := new(big.Rat).SetString(strings.TrimSpace(l.CostBasisPer)); ok {
				costPer = v
			}
		}
		costTotal := new(big.Rat).Mul(sold, costPer)

		plan = append(plan, lotConsumptionPlan{
			LotID:           l.ID,
			AcquisitionDate: strings.TrimSpace(l.AcquisitionDate),
			Provenance:      strings.TrimSpace(l.Provenance),
			SoldQuantity:    sold.FloatString(8),
			NewQuantity:     newQty.FloatString(8),
			CostBasisTotal:  formatRatDecimalScale(costTotal, 2),
			// Proceeds/PNL are filled later (need sell price).
			Proceeds:    "0",
			RealizedPnL: "0",
		})

		if strings.TrimSpace(l.Provenance) == "stock_dividend" {
			dividendSold.Add(dividendSold, sold)
		}
		remaining.Sub(remaining, sold)
	}

	return plan, dividendSold.FloatString(8), nil
}

func normalizeLotProvenance(p *string) string {
	if p == nil {
		return ""
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return ""
	}
	v = strings.ToLower(v)
	v = strings.ReplaceAll(v, "-", "_")
	switch v {
	case "regular_buy", "regularbuy", "buy":
		return "regular_buy"
	case "stock_dividend", "stockdividend", "dividend":
		return "stock_dividend"
	case "rights_offering", "rightsoffering", "rights":
		return "rights_offering"
	case "regular_buy_security", "regularybuy":
		return "regular_buy"
	default:
		// Also accept goen-style enum strings.
		uv := strings.ToUpper(strings.TrimSpace(*p))
		switch uv {
		case "REGULAR_BUY":
			return "regular_buy"
		case "STOCK_DIVIDEND":
			return "stock_dividend"
		case "RIGHTS_OFFERING":
			return "rights_offering"
		default:
			return v
		}
	}
}

func addBusinessDays(d time.Time, days int) time.Time {
	if days <= 0 {
		return d
	}
	cur := d
	added := 0
	for added < days {
		cur = cur.AddDate(0, 0, 1)
		wd := cur.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			added++
		}
	}
	return cur
}

func (s *Service) upsertHoldingFromLots(ctx context.Context, userID, brokerAccountID, securityID string) error {
	lots, err := s.repo.ListShareLots(ctx, userID, brokerAccountID, securityID)
	if err != nil {
		return err
	}

	totalQty := big.NewRat(0, 1)
	totalCost := big.NewRat(0, 1)
	for _, l := range lots {
		if strings.TrimSpace(l.Status) != "active" {
			continue
		}
		q, ok := new(big.Rat).SetString(strings.TrimSpace(l.Quantity))
		if !ok || q.Cmp(big.NewRat(0, 1)) <= 0 {
			continue
		}
		cp := big.NewRat(0, 1)
		if strings.TrimSpace(l.CostBasisPer) != "" {
			if v, ok := new(big.Rat).SetString(strings.TrimSpace(l.CostBasisPer)); ok {
				cp = v
			}
		}
		totalQty.Add(totalQty, q)
		totalCost.Add(totalCost, new(big.Rat).Mul(q, cp))
	}

	createdAt := time.Now().UTC()
	if existing, err := s.repo.GetHolding(ctx, userID, brokerAccountID, securityID); err == nil && existing != nil {
		createdAt = existing.CreatedAt
	}

	qtyStr := totalQty.FloatString(8)
	costStrVal := formatRatDecimalScale(totalCost, 2)
	costStr := &costStrVal
	var avgStr *string
	if totalQty.Cmp(big.NewRat(0, 1)) > 0 {
		a := new(big.Rat).Quo(totalCost, totalQty)
		avgVal := formatRatDecimalScale(a, 8)
		avgStr = &avgVal
	} else {
		z := "0"
		avgStr = &z
	}

	now := time.Now().UTC()
	h := domain.Holding{
		ID:              uuid.NewString(),
		BrokerAccountID: brokerAccountID,
		SecurityID:      securityID,
		Quantity:        qtyStr,
		CostBasisTotal:  costStr,
		AvgCost:         avgStr,
		SourceOfTruth:   "lots",
		CreatedAt:       createdAt,
		UpdatedAt:       now,
	}
	_, err = s.repo.UpsertHolding(ctx, userID, h)
	return err
}

func (s *Service) upsertHoldingFromTrade(ctx context.Context, userID, brokerAccountID, securityID, side, quantity, price string) error {
	q, ok := new(big.Rat).SetString(quantity)
	if !ok {
		return apperrors.Validation("quantity must be a decimal string", nil)
	}
	p, ok := new(big.Rat).SetString(price)
	if !ok {
		return apperrors.Validation("price must be a decimal string", nil)
	}

	oldQty := big.NewRat(0, 1)
	oldCost := big.NewRat(0, 1)
	createdAt := time.Now().UTC()

	if existing, err := s.repo.GetHolding(ctx, userID, brokerAccountID, securityID); err == nil {
		if v, ok := new(big.Rat).SetString(existing.Quantity); ok {
			oldQty = v
		}
		if existing.CostBasisTotal != nil {
			if v, ok := new(big.Rat).SetString(*existing.CostBasisTotal); ok {
				oldCost = v
			}
		} else if existing.AvgCost != nil {
			if avg, ok := new(big.Rat).SetString(*existing.AvgCost); ok {
				oldCost = new(big.Rat).Mul(oldQty, avg)
			}
		}
		createdAt = existing.CreatedAt
	}

	newQty := new(big.Rat).Set(oldQty)
	newCost := new(big.Rat).Set(oldCost)

	if side == "buy" {
		newQty.Add(newQty, q)
		newCost.Add(newCost, new(big.Rat).Mul(q, p))
	} else {
		if oldQty.Cmp(q) < 0 {
			return apperrors.Validation("sell quantity exceeds holding quantity", map[string]any{"field": "quantity"})
		}
		if oldQty.Cmp(big.NewRat(0, 1)) == 0 {
			return apperrors.Validation("cannot sell with zero holdings", map[string]any{"field": "quantity"})
		}
		avg := new(big.Rat).Quo(oldCost, oldQty)
		newQty.Sub(newQty, q)
		newCost.Sub(newCost, new(big.Rat).Mul(q, avg))
	}

	if newQty.Cmp(big.NewRat(0, 1)) < 0 {
		return apperrors.Validation("holding quantity cannot be negative", map[string]any{"field": "quantity"})
	}

	var qtyStr = newQty.FloatString(8)
	var costStr *string
	var avgStr *string
	if newQty.Cmp(big.NewRat(0, 1)) == 0 {
		z := "0"
		costStr = &z
	} else {
		c := newCost.FloatString(2)
		costStr = &c
		a := new(big.Rat).Quo(newCost, newQty).FloatString(8)
		avgStr = &a
	}

	now := time.Now().UTC()
	h := domain.Holding{
		ID:              uuid.NewString(),
		BrokerAccountID: brokerAccountID,
		SecurityID:      securityID,
		Quantity:        qtyStr,
		CostBasisTotal:  costStr,
		AvgCost:         avgStr,
		SourceOfTruth:   "trades",
		CreatedAt:       createdAt,
		UpdatedAt:       now,
	}

	_, err := s.repo.UpsertHolding(ctx, userID, h)
	return err
}

func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeDecimalOrZero(s *string) string {
	if s == nil {
		return ""
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return ""
	}
	return v
}

type tradeChargeContext struct {
	Side       string
	SecurityID string
	Quantity   string
	Price      string
	// DividendQuantity is the number of shares sold that came from STOCK_DIVIDEND lots (FIFO-derived).
	DividendQuantity string
	Allocations      []chargeAllocation
}

type chargeAllocation struct {
	BucketKey string
	Quantity  *big.Rat
	BasePrice *big.Rat
}

type feeConfigV2 struct {
	Name              string  `json:"name"`
	Enabled           *bool   `json:"enabled,omitempty"`
	TriggerEvent      string  `json:"trigger_event"`
	CalculationMethod string  `json:"calculation_method"`
	Value             string  `json:"value"`
	MinFee            *string `json:"min_fee,omitempty"`
	MaxFee            *string `json:"max_fee,omitempty"`
	// BasePricePerShare is used for STOCK_DIVIDEND trigger_event.
	// If omitted, defaults to 10000.
	BasePricePerShare *string `json:"base_price_per_share,omitempty"`
}

type feeSettingsV2 struct {
	Version int           `json:"version,omitempty"`
	Fees    []feeConfigV2 `json:"fees"`
}

func parseFeeConfigsV2(settings any) ([]feeConfigV2, bool, error) {
	if settings == nil {
		return nil, false, nil
	}

	b, err := json.Marshal(settings)
	if err != nil {
		return nil, false, err
	}

	var probe map[string]any
	if err := json.Unmarshal(b, &probe); err != nil {
		return nil, false, err
	}
	if probe == nil {
		return nil, false, nil
	}
	if _, ok := probe["fees"]; !ok {
		return nil, false, nil
	}

	var fs feeSettingsV2
	if err := json.Unmarshal(b, &fs); err != nil {
		return nil, true, err
	}
	if len(fs.Fees) == 0 {
		return nil, true, nil
	}
	return fs.Fees, true, nil
}

func isEnabledV2(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}

func normalizeTriggerEventV2(s string) string {
	v := strings.TrimSpace(s)
	v = strings.ToUpper(v)
	return v
}

func normalizeCalcMethodV2(s string) string {
	v := strings.TrimSpace(s)
	v = strings.ToUpper(v)
	return v
}

func roundRat(r *big.Rat, scale int) *big.Rat {
	if r == nil {
		return big.NewRat(0, 1)
	}
	if scale < 0 {
		scale = 0
	}

	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	num := new(big.Int).Mul(r.Num(), factor)
	den := new(big.Int).Set(r.Denom())
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	// round half-up for non-negative values
	if rem.Sign() >= 0 {
		twoRem := new(big.Int).Mul(rem, big.NewInt(2))
		if twoRem.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	}
	return new(big.Rat).SetFrac(q, factor)
}

func formatRatDecimalScale(r *big.Rat, scale int) string {
	rr := roundRat(r, scale)
	return rr.FloatString(scale)
}

func computeChargesFromFeeConfigsV2(cfgs []feeConfigV2, ctx tradeChargeContext) (string, error) {
	q, ok := new(big.Rat).SetString(strings.TrimSpace(ctx.Quantity))
	if !ok {
		return "", apperrors.Validation("quantity must be a decimal string", map[string]any{"field": "quantity"})
	}
	p, ok := new(big.Rat).SetString(strings.TrimSpace(ctx.Price))
	if !ok {
		return "", apperrors.Validation("price must be a decimal string", map[string]any{"field": "price"})
	}
	notional := new(big.Rat).Mul(q, p)

	divQty := big.NewRat(0, 1)
	if strings.TrimSpace(ctx.DividendQuantity) != "" {
		if v, ok := new(big.Rat).SetString(strings.TrimSpace(ctx.DividendQuantity)); ok {
			divQty = v
		}
	}

	total := big.NewRat(0, 1)
	for _, c := range cfgs {
		if !isEnabledV2(c.Enabled) {
			continue
		}

		event := normalizeTriggerEventV2(c.TriggerEvent)
		apply := false
		switch event {
		case "BUY_SECURITY":
			apply = ctx.Side == "buy"
		case "SELL_SECURITY":
			apply = ctx.Side == "sell"
		case "STOCK_DIVIDEND":
			apply = ctx.Side == "sell" && divQty.Cmp(big.NewRat(0, 1)) > 0
		default:
			// Ignore events not relevant for trade creation.
			apply = false
		}
		if !apply {
			continue
		}

		valueStr := strings.TrimSpace(c.Value)
		if valueStr == "" || !isValidDecimal(valueStr) {
			return "", apperrors.Validation("fee value must be a decimal string", nil)
		}
		val, _ := new(big.Rat).SetString(valueStr)

		base := new(big.Rat).Set(notional)
		qtyBase := new(big.Rat).Set(q)
		if event == "STOCK_DIVIDEND" {
			bp := "10000"
			if c.BasePricePerShare != nil && strings.TrimSpace(*c.BasePricePerShare) != "" {
				if !isValidDecimal(*c.BasePricePerShare) {
					return "", apperrors.Validation("base_price_per_share must be a decimal string", nil)
				}
				bp = strings.TrimSpace(*c.BasePricePerShare)
			}
			bpRat, _ := new(big.Rat).SetString(bp)
			base = new(big.Rat).Mul(divQty, bpRat)
			qtyBase = new(big.Rat).Set(divQty)
		}

		method := normalizeCalcMethodV2(c.CalculationMethod)
		amount := big.NewRat(0, 1)
		switch method {
		case "PERCENTAGE":
			pct := new(big.Rat).Quo(val, big.NewRat(100, 1))
			amount = new(big.Rat).Mul(base, pct)
		case "FIXED_AMOUNT":
			amount = new(big.Rat).Set(val)
		case "PER_SHARE":
			amount = new(big.Rat).Mul(qtyBase, val)
		default:
			return "", apperrors.Validation("calculation_method is invalid", nil)
		}

		if c.MinFee != nil && strings.TrimSpace(*c.MinFee) != "" {
			if !isValidDecimal(*c.MinFee) {
				return "", apperrors.Validation("min_fee must be a decimal string", nil)
			}
			minRat, _ := new(big.Rat).SetString(strings.TrimSpace(*c.MinFee))
			if amount.Cmp(minRat) < 0 {
				amount = minRat
			}
		}
		if c.MaxFee != nil && strings.TrimSpace(*c.MaxFee) != "" {
			if !isValidDecimal(*c.MaxFee) {
				return "", apperrors.Validation("max_fee must be a decimal string", nil)
			}
			maxRat, _ := new(big.Rat).SetString(strings.TrimSpace(*c.MaxFee))
			if amount.Cmp(maxRat) > 0 {
				amount = maxRat
			}
		}

		if amount.Cmp(big.NewRat(0, 1)) < 0 {
			return "", apperrors.Validation("computed charges cannot be negative", nil)
		}
		total.Add(total, amount)
	}

	if total.Cmp(big.NewRat(0, 1)) < 0 {
		return "", apperrors.Validation("computed charges cannot be negative", nil)
	}
	return formatRatDecimalScale(total, 2), nil
}

func validateChargeSettings(settings any, field string) error {
	if settings == nil {
		return nil
	}

	if cfgs, ok, err := parseFeeConfigsV2(settings); err != nil {
		return apperrors.Validation(field+" is invalid", map[string]any{"field": field})
	} else if ok {
		for _, c := range cfgs {
			if strings.TrimSpace(c.TriggerEvent) == "" {
				return apperrors.Validation(field+" fee trigger_event is required", map[string]any{"field": field})
			}
			ev := normalizeTriggerEventV2(c.TriggerEvent)
			switch ev {
			case "BUY_SECURITY", "SELL_SECURITY", "CASH_WITHDRAWAL", "CASH_DIVIDEND", "STOCK_DIVIDEND", "MONTHLY_CUSTODY":
				// ok
			default:
				return apperrors.Validation(field+" fee trigger_event is invalid", map[string]any{"field": field})
			}
			cm := normalizeCalcMethodV2(c.CalculationMethod)
			if cm != "PERCENTAGE" && cm != "FIXED_AMOUNT" && cm != "PER_SHARE" {
				return apperrors.Validation(field+" fee calculation_method is invalid", map[string]any{"field": field})
			}
			if strings.TrimSpace(c.Value) == "" || !isValidDecimal(c.Value) {
				return apperrors.Validation(field+" fee value is invalid", map[string]any{"field": field})
			}
			if c.MinFee != nil && strings.TrimSpace(*c.MinFee) != "" && !isValidDecimal(*c.MinFee) {
				return apperrors.Validation(field+" fee min_fee is invalid", map[string]any{"field": field})
			}
			if c.MaxFee != nil && strings.TrimSpace(*c.MaxFee) != "" && !isValidDecimal(*c.MaxFee) {
				return apperrors.Validation(field+" fee max_fee is invalid", map[string]any{"field": field})
			}
			if c.BasePricePerShare != nil && strings.TrimSpace(*c.BasePricePerShare) != "" && !isValidDecimal(*c.BasePricePerShare) {
				return apperrors.Validation(field+" fee base_price_per_share is invalid", map[string]any{"field": field})
			}
		}
		return nil
	}
	return apperrors.Validation(field+" must be V2 settings: {version:2, fees:[...]} ", map[string]any{"field": field})
}

func computeChargesFromSettings(settings any, ctx tradeChargeContext) (string, error) {
	if cfgs, ok, err := parseFeeConfigsV2(settings); err != nil {
		return "", apperrors.Validation("settings is invalid", nil)
	} else if ok {
		return computeChargesFromFeeConfigsV2(cfgs, ctx)
	}
	return "", nil
}

func deriveTradeExternalRef(clientID *string, tradeID string, kind string) *string {
	if tradeID == "" || kind == "" {
		return nil
	}
	if clientID != nil && strings.TrimSpace(*clientID) != "" {
		v := fmt.Sprintf("trade:%s:%s:%s", strings.TrimSpace(*clientID), tradeID, kind)
		return &v
	}
	v := fmt.Sprintf("trade:%s:%s", tradeID, kind)
	return &v
}

// normalizeOccurredAt mirrors transaction module parsing behavior.
func normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
	if occurredAt != nil && strings.TrimSpace(*occurredAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*occurredAt))
		if err != nil {
			return time.Time{}, "", apperrors.Validation("occurred_at is invalid", map[string]any{"field": "occurred_at"})
		}
		return parsed.UTC(), parsed.UTC().Format("2006-01-02"), nil
	}

	date := ""
	if occurredDate != nil {
		date = strings.TrimSpace(*occurredDate)
	}
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return time.Time{}, "", apperrors.Validation("occurred_date is invalid", map[string]any{"field": "occurred_date"})
	}

	tStr := "00:00"
	if occurredTime != nil && strings.TrimSpace(*occurredTime) != "" {
		tStr = strings.TrimSpace(*occurredTime)
		// allow HH:MM or HH:MM:SS
		if _, err := time.Parse("15:04", tStr); err != nil {
			if _, err2 := time.Parse("15:04:05", tStr); err2 != nil {
				return time.Time{}, "", apperrors.Validation("occurred_time is invalid", map[string]any{"field": "occurred_time"})
			}
		}
	}

	// Build RFC3339 timestamp in UTC.
	stamp := fmt.Sprintf("%sT%s:00Z", date, normalizeTimeToHHMM(tStr))
	parsed, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return time.Time{}, "", apperrors.Validation("occurred_at is invalid", map[string]any{"field": "occurred_at"})
	}
	return parsed.UTC(), date, nil
}

func normalizeTimeToHHMM(t string) string {
	if strings.TrimSpace(t) == "" {
		return "00:00"
	}
	if parsed, err := time.Parse("15:04", t); err == nil {
		return parsed.Format("15:04")
	}
	if parsed, err := time.Parse("15:04:05", t); err == nil {
		return parsed.Format("15:04")
	}
	return "00:00"
}

func (s *Service) ListEligibleCorporateActions(ctx context.Context, userID, investmentAccountID string) ([]EligibleAction, error) {
	bid := strings.TrimSpace(investmentAccountID)
	if bid == "" {
		return nil, apperrors.Validation("investmentAccountId is required", nil)
	}

	holdings, err := s.repo.ListHoldings(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	trades, err := s.repo.ListTrades(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	tradesBySecurity := make(map[string][]domain.Trade)
	securityIDs := make(map[string]struct{})
	for _, tr := range trades {
		securityID := strings.TrimSpace(tr.SecurityID)
		if securityID == "" {
			continue
		}
		tradesBySecurity[securityID] = append(tradesBySecurity[securityID], tr)
		securityIDs[securityID] = struct{}{}
	}
	for _, h := range holdings {
		securityID := strings.TrimSpace(h.SecurityID)
		if securityID == "" {
			continue
		}
		securityIDs[securityID] = struct{}{}
	}

	elections, err := s.repo.ListSecurityEventElections(ctx, userID, bid, nil)
	if err != nil {
		return nil, err
	}
	electionMap := make(map[string]domain.SecurityEventElection)
	for _, el := range elections {
		electionMap[el.SecurityEventID] = el
	}

	var out []EligibleAction
	for securityID := range securityIDs {
		events, err := s.repo.ListSecurityEvents(ctx, securityID, nil, nil)
		if err != nil {
			continue
		}

		for _, ev := range events {
			entitlementDate, hasEntitlementDate := entitlementAsOfDate(ev)
			if !hasEntitlementDate {
				continue
			}

			holdingQty := holdingQuantityAsOf(tradesBySecurity[securityID], entitlementDate)
			holdingQtyText := formatRatDecimalScale(holdingQty, 8)

			el, claimed := electionMap[ev.ID]
			if claimed && el.Status == "dismissed" {
				continue
			}

			status := "eligible"
			if claimed {
				status = "claimed"
			}

			// Entitlement is computed from point-in-time holdings at ex_date (or record_date fallback).
			entitled := "0"
			qty := holdingQty
			if ev.EventType == "dividend_cash" && ev.CashAmountPerShare != nil {
				cash, ok := new(big.Rat).SetString(*ev.CashAmountPerShare)
				if ok {
					entitled = formatRatDecimalScale(new(big.Rat).Mul(qty, cash), 2)
				}
			} else if (ev.EventType == "bonus_issue" || ev.EventType == "split" || ev.EventType == "stock_dividend") && ev.RatioNumerator != nil && ev.RatioDenominator != nil {
				num, numOK := new(big.Rat).SetString(*ev.RatioNumerator)
				den, denOK := new(big.Rat).SetString(*ev.RatioDenominator)
				if numOK && denOK && den.Cmp(new(big.Rat)) > 0 {
					entitled = formatRatDecimalScale(new(big.Rat).Quo(new(big.Rat).Mul(qty, num), den), 8)
				}
			}

			var elID *string
			if claimed {
				elID = &el.ID
			}

			out = append(out, EligibleAction{
				Event:            ev,
				HoldingQuantity:  holdingQtyText,
				EntitledQuantity: entitled,
				Status:           status,
				ElectionID:       elID,
			})
		}
	}

	return out, nil
}

func entitlementAsOfDate(ev domain.SecurityEvent) (time.Time, bool) {
	if dt, ok := parseEventDate(ev.ExDate); ok {
		return dt, true
	}
	if dt, ok := parseEventDate(ev.RecordDate); ok {
		return dt, true
	}
	return time.Time{}, false
}

func parseEventDate(raw *string) (time.Time, bool) {
	if raw == nil {
		return time.Time{}, false
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return time.Time{}, false
	}
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	return date.UTC().Add(24*time.Hour - time.Nanosecond), true
}

func holdingQuantityAsOf(trades []domain.Trade, asOf time.Time) *big.Rat {
	total := big.NewRat(0, 1)
	for _, tr := range trades {
		if tr.OccurredAt.UTC().After(asOf) {
			continue
		}

		qty, ok := new(big.Rat).SetString(strings.TrimSpace(tr.Quantity))
		if !ok {
			continue
		}

		side := strings.ToLower(strings.TrimSpace(tr.Side))
		switch side {
		case "buy":
			total = new(big.Rat).Add(total, qty)
		case "sell":
			total = new(big.Rat).Sub(total, qty)
		}
	}

	if total.Cmp(big.NewRat(0, 1)) < 0 {
		return big.NewRat(0, 1)
	}
	return total
}

func (s *Service) ClaimCorporateAction(ctx context.Context, userID, investmentAccountID, securityEventID string, req ClaimCorporateActionRequest) (*domain.SecurityEventElection, error) {
	bid := strings.TrimSpace(investmentAccountID)
	evID := strings.TrimSpace(securityEventID)
	if bid == "" || evID == "" {
		return nil, apperrors.Validation("investmentAccountId and securityEventId are required", nil)
	}

	ev, err := s.repo.GetSecurityEvent(ctx, evID)
	if err != nil {
		return nil, err
	}

	h, err := s.repo.GetHolding(ctx, userID, bid, ev.SecurityID)
	if err != nil {
		return nil, err
	}

	// Double-claim check.
	elections, err := s.repo.ListSecurityEventElections(ctx, userID, bid, nil)
	if err == nil {
		for _, el := range elections {
			if el.SecurityEventID == evID && el.Status != "dismissed" {
				return nil, apperrors.Validation("this event has already been claimed for this account", nil)
			}
		}
	}

	qty, _ := new(big.Rat).SetString(h.Quantity)
	entitled := "0"
	if ev.EventType == "dividend_cash" && ev.CashAmountPerShare != nil {
		cash, _ := new(big.Rat).SetString(*ev.CashAmountPerShare)
		entitled = formatRatDecimalScale(new(big.Rat).Mul(qty, cash), 2)
	} else if (ev.EventType == "bonus_issue" || ev.EventType == "split" || ev.EventType == "stock_dividend") && ev.RatioNumerator != nil && ev.RatioDenominator != nil {
		num, _ := new(big.Rat).SetString(*ev.RatioNumerator)
		den, _ := new(big.Rat).SetString(*ev.RatioDenominator)
		if den.Cmp(new(big.Rat)) > 0 {
			entitled = formatRatDecimalScale(new(big.Rat).Quo(new(big.Rat).Mul(qty, num), den), 8)
		}
	}

	elected := entitled
	if req.ElectedQuantity != nil && strings.TrimSpace(*req.ElectedQuantity) != "" {
		elected = *req.ElectedQuantity
	}

	now := time.Now().UTC()
	election := domain.SecurityEventElection{
		ID:                           uuid.NewString(),
		UserID:                       userID,
		BrokerAccountID:              bid,
		SecurityEventID:              evID,
		SecurityID:                   ev.SecurityID,
		EntitlementDate:              derefString(ev.ExDate),
		HoldingQuantityAtEntitlement: h.Quantity,
		EntitledQuantity:             entitled,
		ElectedQuantity:              elected,
		Status:                       "confirmed",
		ConfirmedAt:                  &now,
		Note:                         req.Note,
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}

	// 1. Create the election record.
	res, err := s.repo.UpsertSecurityEventElection(ctx, userID, election)
	if err != nil {
		return nil, err
	}

	// 2. Perform the actual action.
	if ev.EventType == "dividend_cash" {
		ia, _ := s.repo.GetInvestmentAccount(ctx, userID, bid)
		desc := fmt.Sprintf("Cash Dividend: %s", ev.SecurityID)
		if ev.Note != nil {
			desc += " - " + *ev.Note
		}
		occDate := time.Now().UTC().Format("2006-01-02")
		if ev.PayDate != nil {
			occDate = *ev.PayDate
		}
		_, err = s.txSvc.Create(ctx, userID, transaction.CreateRequest{
			Type:         "income",
			OccurredDate: &occDate,
			Amount:       elected,
			Description:  &desc,
			AccountID:    &ia.AccountID,
			ExternalRef:  ptr(deriveEventExternalRef(evID, bid, "cash")),
		})
	} else if ev.EventType == "bonus_issue" || ev.EventType == "stock_dividend" || ev.EventType == "split" {
		occDate := time.Now().UTC().Format("2006-01-02")
		if ev.EffectiveDate != nil {
			occDate = *ev.EffectiveDate
		}

		// Calculation for stock-based events:
		// 1. Trade quantity = Floor(elected entitlement)
		// 2. Residual cash = Fraction * 10,000 (par value in VND)
		entRat, _ := new(big.Rat).SetString(elected)

		floorQty := new(big.Int).Div(entRat.Num(), entRat.Denom())
		tradeQty := new(big.Rat).SetInt(floorQty)

		residualQty := new(big.Rat).Sub(entRat, tradeQty)

		// Create trade for whole shares if any
		if tradeQty.Cmp(big.NewRat(0, 1)) > 0 {
			prov := "stock_dividend"
			_, err = s.CreateTrade(ctx, userID, bid, CreateTradeRequest{
				SecurityID:   ev.SecurityID,
				Side:         "buy",
				Quantity:     tradeQty.FloatString(8),
				Price:        "0",
				Provenance:   &prov,
				OccurredDate: &occDate,
				Note:         req.Note,
			})
			if err != nil {
				return nil, err
			}
		}

		// Create income transaction for residual cash if any
		if residualQty.Cmp(big.NewRat(0, 1)) > 0 {
			parValue := big.NewRat(10000, 1)
			cashAmt := new(big.Rat).Mul(residualQty, parValue)
			cashAmtStr := formatRatDecimalScale(cashAmt, 2)

			ia, _ := s.repo.GetInvestmentAccount(ctx, userID, bid)
			desc := fmt.Sprintf("Residual Cash (%s): %s", ev.EventType, ev.SecurityID)

			_, err = s.txSvc.Create(ctx, userID, transaction.CreateRequest{
				Type:         "income",
				OccurredDate: &occDate,
				Amount:       cashAmtStr,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  ptr(deriveEventExternalRef(evID, bid, "residual")),
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return res, err
}

func deriveEventExternalRef(eventID, brokerAccountID, suffix string) string {
	return fmt.Sprintf("event:%s:%s:%s", eventID, brokerAccountID, suffix)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptr[T any](v T) *T {
	return &v
}

func (s *Service) GetRealizedPNLReport(ctx context.Context, userID, brokerAccountID string) (*domain.RealizedPNLReport, error) {
	// 1. Fetch data
	logs, err := s.repo.ListRealizedLogs(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}
	divs, err := s.repo.ListDividends(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}
	trades, err := s.repo.ListTrades(ctx, userID, brokerAccountID)
	if err != nil {
		return nil, err
	}
	securities, err := s.repo.ListSecurities(ctx)
	if err != nil {
		return nil, err
	}
	secMap := make(map[string]domain.Security)
	for _, sec := range securities {
		secMap[sec.ID] = sec
	}

	// 2. Aggregate
	items := make(map[string]*domain.RealizedPNLReportItem)

	// Process Trade Logs (Proceeds, Cost Basis, Trade PNL)
	for _, l := range logs {
		it := items[l.SecurityID]
		if it == nil {
			it = &domain.RealizedPNLReportItem{
				SecurityID:        l.SecurityID,
				Symbol:            secMap[l.SecurityID].Symbol,
				GrossRealizedGain: "0",
				TradeGain:         "0",
				DividendGain:      "0",
				Proceeds:          "0",
				CostBasis:         "0",
				Fees:              "0",
				Taxes:             "0",
				NetPNL:            "0",
			}
			items[l.SecurityID] = it
		}
		it.TradeGain = addMoneyStrings(it.TradeGain, l.RealizedPnL)
		it.Proceeds = addMoneyStrings(it.Proceeds, l.Proceeds)
		it.CostBasis = addMoneyStrings(it.CostBasis, l.CostBasisTotal)
	}

	// Process Dividends
	// We need to map dividend transactions to securities.
	// Since we don't have a direct link in Transaction struct, we'll fetch events to map IDs.
	// (Optimization: we could parse ExternalRef if it contains SecurityID, but currently it's eventID).
	eventToSec := make(map[string]string)
	for _, d := range divs {
		if d.ExternalRef == nil {
			continue
		}
		parts := strings.Split(*d.ExternalRef, ":")
		if len(parts) >= 2 && parts[0] == "event" {
			evID := parts[1]
			if _, ok := eventToSec[evID]; !ok {
				ev, err := s.repo.GetSecurityEvent(ctx, evID)
				if err == nil && ev != nil {
					eventToSec[evID] = ev.SecurityID
				}
			}
			secID := eventToSec[evID]
			if secID != "" {
				it := items[secID]
				if it == nil {
					it = &domain.RealizedPNLReportItem{
						SecurityID:        secID,
						Symbol:            secMap[secID].Symbol,
						GrossRealizedGain: "0",
						TradeGain:         "0",
						DividendGain:      "0",
						Proceeds:          "0",
						CostBasis:         "0",
						Fees:              "0",
						Taxes:             "0",
						NetPNL:            "0",
					}
					items[secID] = it
				}
				it.DividendGain = addMoneyStrings(it.DividendGain, d.Amount)
			}
		}
	}

	// Process Fees and Taxes from Trades
	// Note: We sum all fees/taxes for the security.
	// In some contexts, you might only want SELL fees for realized PNL,
	// but usually all trade costs for that security are deducted from its lifecycle profit.
	for _, t := range trades {
		it := items[t.SecurityID]
		if it == nil {
			continue // No realized gains yet, and we only report on realized items for now.
		}
		it.Fees = addMoneyStrings(it.Fees, t.Fees)
		it.Taxes = addMoneyStrings(it.Taxes, t.Taxes)
	}

	// 3. Finalize items (Gross and Net)
	var report domain.RealizedPNLReport
	totalGross := big.NewRat(0, 1)
	totalNet := big.NewRat(0, 1)

	for _, it := range items {
		gross := addMoneyStrings(it.TradeGain, it.DividendGain)
		it.GrossRealizedGain = gross

		costs := addMoneyStrings(it.Fees, it.Taxes)
		it.NetPNL = subMoneyStrings(gross, costs)

		if gRat, ok := new(big.Rat).SetString(it.GrossRealizedGain); ok {
			totalGross.Add(totalGross, gRat)
		}
		if nRat, ok := new(big.Rat).SetString(it.NetPNL); ok {
			totalNet.Add(totalNet, nRat)
		}

		report.Items = append(report.Items, *it)
	}

	report.TotalGross = formatRatDecimalScale(totalGross, 2)
	report.TotalNet = formatRatDecimalScale(totalNet, 2)

	return &report, nil
}

func addMoneyStrings(a, b string) string {
	ra, ok1 := new(big.Rat).SetString(strings.TrimSpace(a))
	rb, ok2 := new(big.Rat).SetString(strings.TrimSpace(b))
	if !ok1 {
		ra = big.NewRat(0, 1)
	}
	if !ok2 {
		rb = big.NewRat(0, 1)
	}
	return formatRatDecimalScale(new(big.Rat).Add(ra, rb), 2)
}

func subMoneyStrings(a, b string) string {
	ra, ok1 := new(big.Rat).SetString(strings.TrimSpace(a))
	rb, ok2 := new(big.Rat).SetString(strings.TrimSpace(b))
	if !ok1 {
		ra = big.NewRat(0, 1)
	}
	if !ok2 {
		rb = big.NewRat(0, 1)
	}
	return formatRatDecimalScale(new(big.Rat).Sub(ra, rb), 2)
}

func isValidDecimal(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	return ok
}
