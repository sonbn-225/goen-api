package investment

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
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
}

// Service handles investment business logic.
type Service struct {
	repo       domain.InvestmentRepository
	accountSvc AccountServiceDep
	txSvc      TransactionServiceDep
	cfg        *config.Config
	redis      *storage.Redis
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

	// Auto-create fee/tax transactions if requested amounts > 0 and no explicit transaction ids provided.
	if feeTxID == nil {
		if amt, ok := new(big.Rat).SetString(fees); ok && amt.Cmp(new(big.Rat)) > 0 {
			desc := "Trade fee"
			if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil {
				desc = "Trade fee: " + sec.Symbol
			}
			externalRef := deriveTradeExternalRef(req.ClientID, tradeID, "fee")
			tx, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
				Type:         "expense",
				OccurredDate: &occurredDate,
				Amount:       fees,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  externalRef,
			})
			if err != nil {
				return nil, err
			}
			feeTxID = &tx.ID
		}
	}

	if taxTxID == nil {
		if amt, ok := new(big.Rat).SetString(taxes); ok && amt.Cmp(new(big.Rat)) > 0 {
			desc := "Trade tax"
			if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil {
				desc = "Trade tax: " + sec.Symbol
			}
			externalRef := deriveTradeExternalRef(req.ClientID, tradeID, "tax")
			tx, err := s.txSvc.Create(ctx, userID, transaction.CreateRequest{
				Type:         "expense",
				OccurredDate: &occurredDate,
				Amount:       taxes,
				Description:  &desc,
				AccountID:    &ia.AccountID,
				ExternalRef:  externalRef,
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

func isValidDecimal(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	return ok
}
