package investment

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListInvestmentAccounts(ctx context.Context, userID string) ([]InvestmentAccount, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	items, err := s.repo.ListInvestmentAccounts(ctx, userID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list investment accounts", err)
	}
	return items, nil
}

func (s *service) GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*InvestmentAccount, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	investmentAccountID = strings.TrimSpace(investmentAccountID)
	if investmentAccountID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "investmentAccountId is required")
	}

	item, err := s.repo.GetInvestmentAccount(ctx, userID, investmentAccountID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get investment account", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "investment account not found")
	}
	return item, nil
}

func (s *service) UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, input UpdateInvestmentAccountSettingsInput) (*InvestmentAccount, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	investmentAccountID = strings.TrimSpace(investmentAccountID)
	if investmentAccountID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "investmentAccountId is required")
	}
	if input.FeeSettings == nil && input.TaxSettings == nil {
		return nil, apperrors.New(apperrors.KindValidation, "no fields to update")
	}

	item, err := s.repo.UpdateInvestmentAccountSettings(ctx, userID, investmentAccountID, input.FeeSettings, input.TaxSettings)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to update investment account settings", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "investment account not found")
	}
	return item, nil
}

func (s *service) ListSecurities(ctx context.Context, userID string) ([]Security, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	items, err := s.repo.ListSecurities(ctx)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list securities", err)
	}
	return items, nil
}

func (s *service) GetSecurity(ctx context.Context, userID, securityID string) (*Security, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	securityID = strings.TrimSpace(securityID)
	if securityID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "securityId is required")
	}

	item, err := s.repo.GetSecurity(ctx, securityID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get security", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "security not found")
	}
	return item, nil
}

func (s *service) ListSecurityPrices(ctx context.Context, userID, securityID string, from *string, to *string) ([]SecurityPriceDaily, error) {
	if _, err := s.GetSecurity(ctx, userID, securityID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListSecurityPrices(ctx, securityID, from, to)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list security prices", err)
	}
	return items, nil
}

func (s *service) ListSecurityEvents(ctx context.Context, userID, securityID string, from *string, to *string) ([]SecurityEvent, error) {
	if _, err := s.GetSecurity(ctx, userID, securityID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListSecurityEvents(ctx, securityID, from, to)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list security events", err)
	}
	return items, nil
}

func (s *service) CreateTrade(ctx context.Context, userID, investmentAccountID string, input CreateTradeInput) (*Trade, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "investment", "operation", "create_trade")
	logger.Info("investment_create_trade_started", "user_id", userID, "investment_account_id", investmentAccountID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	investmentAccountID = strings.TrimSpace(investmentAccountID)
	if investmentAccountID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "investmentAccountId is required")
	}

	if _, err := s.GetInvestmentAccount(ctx, userID, investmentAccountID); err != nil {
		return nil, err
	}

	securityID := strings.TrimSpace(input.SecurityID)
	if securityID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "security_id is required")
	}
	if _, err := s.GetSecurity(ctx, userID, securityID); err != nil {
		return nil, err
	}

	side := strings.ToLower(strings.TrimSpace(input.Side))
	if side != "buy" && side != "sell" {
		return nil, apperrors.New(apperrors.KindValidation, "side must be one of: buy, sell")
	}

	qtyRat, err := parsePositiveDecimal(input.Quantity)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "quantity must be a decimal string greater than zero")
	}
	priceRat, err := parsePositiveDecimal(input.Price)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "price must be a decimal string greater than zero")
	}

	feesRat, err := parseNonNegativeDecimal(defaultDecimal(input.Fees, "0"))
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "fees must be a decimal string greater than or equal to zero")
	}
	taxesRat, err := parseNonNegativeDecimal(defaultDecimal(input.Taxes, "0"))
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "taxes must be a decimal string greater than or equal to zero")
	}

	occurredAt := time.Now().UTC()
	if input.OccurredAt != nil {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*input.OccurredAt))
		if err != nil {
			return nil, apperrors.New(apperrors.KindValidation, "occurred_at must be RFC3339")
		}
		occurredAt = parsed.UTC()
	}

	now := time.Now().UTC()
	trade := &Trade{
		ID:               uuid.NewString(),
		ClientID:         normalizeOptionalString(input.ClientID),
		BrokerAccountID:  investmentAccountID,
		SecurityID:       securityID,
		FeeTransactionID: normalizeOptionalString(input.FeeTransactionID),
		TaxTransactionID: normalizeOptionalString(input.TaxTransactionID),
		Side:             side,
		Quantity:         qtyRat.FloatString(8),
		Price:            priceRat.FloatString(8),
		Fees:             feesRat.FloatString(2),
		Taxes:            taxesRat.FloatString(2),
		OccurredAt:       occurredAt,
		Note:             normalizeOptionalString(input.Note),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateTrade(ctx, userID, *trade); err != nil {
		logger.Error("investment_create_trade_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to create trade", err)
	}

	if err := s.applyTradeToHolding(ctx, userID, *trade); err != nil {
		logger.Error("investment_create_trade_failed", "error", err)
		return nil, err
	}

	logger.Info("investment_create_trade_succeeded", "trade_id", trade.ID)
	return trade, nil
}

func (s *service) ListTrades(ctx context.Context, userID, investmentAccountID string) ([]Trade, error) {
	if _, err := s.GetInvestmentAccount(ctx, userID, investmentAccountID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListTrades(ctx, userID, investmentAccountID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list trades", err)
	}
	return items, nil
}

func (s *service) ListHoldings(ctx context.Context, userID, investmentAccountID string) ([]Holding, error) {
	if _, err := s.GetInvestmentAccount(ctx, userID, investmentAccountID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListHoldings(ctx, userID, investmentAccountID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list holdings", err)
	}
	return items, nil
}

func (s *service) applyTradeToHolding(ctx context.Context, userID string, trade Trade) error {
	current, err := s.repo.GetHolding(ctx, userID, trade.BrokerAccountID, trade.SecurityID)
	if err != nil {
		return passThroughOrWrapInternal("failed to read holding", err)
	}

	currentQty := big.NewRat(0, 1)
	currentCost := big.NewRat(0, 1)
	if current != nil {
		if parsed, ok := new(big.Rat).SetString(current.Quantity); ok {
			currentQty = parsed
		}
		if current.CostBasisTotal != nil {
			if parsed, ok := new(big.Rat).SetString(*current.CostBasisTotal); ok {
				currentCost = parsed
			}
		}
	}

	tradeQty, _ := new(big.Rat).SetString(trade.Quantity)
	tradePrice, _ := new(big.Rat).SetString(trade.Price)
	fees, _ := new(big.Rat).SetString(trade.Fees)
	taxes, _ := new(big.Rat).SetString(trade.Taxes)

	notional := new(big.Rat).Mul(tradeQty, tradePrice)
	newQty := new(big.Rat).Set(currentQty)
	newCost := new(big.Rat).Set(currentCost)

	if trade.Side == "buy" {
		newQty = new(big.Rat).Add(currentQty, tradeQty)
		newCost = new(big.Rat).Add(currentCost, new(big.Rat).Add(notional, new(big.Rat).Add(fees, taxes)))
	} else {
		if currentQty.Cmp(tradeQty) < 0 {
			return apperrors.New(apperrors.KindValidation, "sell quantity exceeds current holding")
		}
		if currentQty.Sign() == 0 {
			return apperrors.New(apperrors.KindValidation, "holding quantity is zero")
		}

		avgCost := new(big.Rat).Quo(currentCost, currentQty)
		costToRemove := new(big.Rat).Mul(avgCost, tradeQty)
		newQty = new(big.Rat).Sub(currentQty, tradeQty)
		newCost = new(big.Rat).Sub(currentCost, costToRemove)
		if newCost.Sign() < 0 {
			newCost = big.NewRat(0, 1)
		}
	}

	holding := Holding{
		ID:              uuid.NewString(),
		BrokerAccountID: trade.BrokerAccountID,
		SecurityID:      trade.SecurityID,
		Quantity:        newQty.FloatString(8),
		SourceOfTruth:   "trades",
		UpdatedAt:       time.Now().UTC(),
		CreatedAt:       time.Now().UTC(),
	}

	if newQty.Sign() > 0 {
		cost := newCost.FloatString(2)
		avg := new(big.Rat).Quo(newCost, newQty).FloatString(8)
		holding.CostBasisTotal = &cost
		holding.AvgCost = &avg
	}

	_, err = s.repo.UpsertHolding(ctx, userID, holding)
	if err != nil {
		return passThroughOrWrapInternal("failed to update holding", err)
	}
	return nil
}

func parsePositiveDecimal(raw string) (*big.Rat, error) {
	v := strings.TrimSpace(raw)
	rat, ok := new(big.Rat).SetString(v)
	if !ok || rat.Sign() <= 0 {
		return nil, errors.New("invalid")
	}
	return rat, nil
}

func parseNonNegativeDecimal(raw string) (*big.Rat, error) {
	v := strings.TrimSpace(raw)
	rat, ok := new(big.Rat).SetString(v)
	if !ok || rat.Sign() < 0 {
		return nil, errors.New("invalid")
	}
	return rat, nil
}

func defaultDecimal(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	clean := strings.TrimSpace(*v)
	if clean == "" {
		return fallback
	}
	return clean
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func passThroughOrWrapInternal(message string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.Wrap(apperrors.KindInternal, message, err)
}
