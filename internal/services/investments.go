package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type InvestmentService interface {
	// Investment accounts
	CreateInvestmentAccount(ctx context.Context, userID string, req CreateInvestmentAccountRequest) (*domain.InvestmentAccount, error)
	GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*domain.InvestmentAccount, error)
	ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error)

	// Securities
	GetSecurity(ctx context.Context, userID string, securityID string) (*domain.Security, error)
	ListSecurities(ctx context.Context, userID string) ([]domain.Security, error)

	// Market data (read-only)
	ListSecurityPrices(ctx context.Context, userID string, securityID string, from *string, to *string) ([]domain.SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, userID string, securityID string, from *string, to *string) ([]domain.SecurityEvent, error)

	// Trades
	CreateTrade(ctx context.Context, userID string, brokerAccountID string, req CreateTradeRequest) (*domain.Trade, error)
	ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]domain.Trade, error)

	// Holdings
	ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]domain.Holding, error)

	// Elections
	UpsertSecurityEventElection(ctx context.Context, userID string, brokerAccountID string, req UpsertSecurityEventElectionRequest) (*domain.SecurityEventElection, error)
	ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]domain.SecurityEventElection, error)
}

type CreateInvestmentAccountRequest struct {
	AccountID             *string `json:"account_id,omitempty"`
	ParentAccountID       *string `json:"parent_account_id,omitempty"`
	BrokerAccountName     *string `json:"broker_account_name,omitempty"`
	BrokerAccountNumber   *string `json:"broker_account_number,omitempty"`
	BrokerAccountColor    *string `json:"broker_account_color,omitempty"`
	BrokerAccountCurrency *string `json:"broker_account_currency,omitempty"`
	FeeSettings           any     `json:"fee_settings,omitempty"`
	TaxSettings           any     `json:"tax_settings,omitempty"`
}

type CreateTradeRequest struct {
	ClientID         *string `json:"client_id,omitempty"`
	SecurityID       string  `json:"security_id"`
	FeeTransactionID *string `json:"fee_transaction_id,omitempty"`
	TaxTransactionID *string `json:"tax_transaction_id,omitempty"`
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

type UpsertSecurityEventElectionRequest struct {
	SecurityEventID string  `json:"security_event_id"`
	ElectedQuantity string  `json:"elected_quantity"`
	Status          *string `json:"status,omitempty"`
	Note            *string `json:"note,omitempty"`
}

type investmentService struct {
	accounts AccountService
	tx       TransactionService
	repo     domain.InvestmentRepository
}

func NewInvestmentService(accounts AccountService, tx TransactionService, repo domain.InvestmentRepository) InvestmentService {
	return &investmentService{accounts: accounts, tx: tx, repo: repo}
}

func (s *investmentService) CreateInvestmentAccount(ctx context.Context, userID string, req CreateInvestmentAccountRequest) (*domain.InvestmentAccount, error) {
	var accountID string
	if req.AccountID != nil {
		accountID = strings.TrimSpace(*req.AccountID)
	}
	parentAccountID := normalizeOptionalString(req.ParentAccountID)

	var acc *domain.Account
	if accountID == "" {
		var inferredCurrency string
		if parentAccountID != nil {
			parent, err := s.accounts.GetAccount(ctx, userID, *parentAccountID)
			if err != nil {
				return nil, err
			}
			if parent.AccountType != "bank" && parent.AccountType != "wallet" {
				return nil, ValidationError("parent account must be bank or wallet", map[string]any{"field": "parent_account_id"})
			}
			inferredCurrency = parent.Currency
		}

		brokerCurrency := strings.ToUpper(strings.TrimSpace(toString(req.BrokerAccountCurrency)))
		if brokerCurrency == "" {
			brokerCurrency = inferredCurrency
		}

		brokerName := strings.TrimSpace(toString(req.BrokerAccountName))
		if brokerName == "" {
			brokerName = "Broker"
		}

		createdAcc, err := s.accounts.CreateAccount(ctx, userID, CreateAccountRequest{
			Name:          brokerName,
			AccountType:   "broker",
			Currency:      brokerCurrency,
			AccountNumber: normalizeOptionalString(req.BrokerAccountNumber),
			Color:         normalizeOptionalString(req.BrokerAccountColor),
		})
		if err != nil {
			return nil, err
		}
		accountID = createdAcc.ID
		acc = createdAcc
	} else {
		cur, err := s.accounts.GetAccount(ctx, userID, accountID)
		if err != nil {
			return nil, err
		}
		acc = cur
	}

	if acc.AccountType != "broker" {
		return nil, ValidationError("account_id must be an account of type broker", map[string]any{"field": "account_id"})
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	item := domain.InvestmentAccount{
		ID:          id,
		AccountID:   accountID,
		FeeSettings: req.FeeSettings,
		TaxSettings: req.TaxSettings,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateInvestmentAccount(ctx, userID, item); err != nil {
		if errors.Is(err, domain.ErrInvestmentForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if isUniqueViolation(err) {
			// Make this endpoint effectively idempotent for a broker account.
			items, listErr := s.repo.ListInvestmentAccounts(ctx, userID)
			if listErr != nil {
				return nil, err
			}
			for _, existing := range items {
				if existing.AccountID == accountID {
					out := existing
					return &out, nil
				}
			}
		}
		return nil, err
	}
	created, err := s.repo.GetInvestmentAccount(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *investmentService) GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*domain.InvestmentAccount, error) {
	id := strings.TrimSpace(investmentAccountID)
	if id == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}
	item, err := s.repo.GetInvestmentAccount(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrInvestmentAccountNotFound) {
			return nil, NotFoundErrorWithCause("investment account not found", nil, err)
		}
		return nil, err
	}
	return item, nil
}

func (s *investmentService) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	// Backfill: accounts of type `broker` created via Accounts UI should appear here.
	// Investment accounts are a 1-1 extension table, so ensure the extension exists.
	accounts, err := s.accounts.ListAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.ListInvestmentAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	existingByAccountID := make(map[string]struct{}, len(existing))
	for _, ia := range existing {
		existingByAccountID[ia.AccountID] = struct{}{}
	}

	createdAny := false
	for _, acc := range accounts {
		if acc.AccountType != "broker" || acc.Status != "active" {
			continue
		}
		if _, ok := existingByAccountID[acc.ID]; ok {
			continue
		}

		now := time.Now().UTC()
		item := domain.InvestmentAccount{
			ID:        uuid.NewString(),
			AccountID: acc.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.repo.CreateInvestmentAccount(ctx, userID, item); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, fmt.Errorf("backfill investment account for broker account %s: %w", acc.ID, err)
		}
		createdAny = true
	}

	if createdAny {
		return s.repo.ListInvestmentAccounts(ctx, userID)
	}
	return existing, nil
}

func (s *investmentService) GetSecurity(ctx context.Context, _ string, securityID string) (*domain.Security, error) {
	id := strings.TrimSpace(securityID)
	if id == "" {
		return nil, ValidationError("securityId is required", map[string]any{"field": "securityId"})
	}
	item, err := s.repo.GetSecurity(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrSecurityNotFound) {
			return nil, NotFoundErrorWithCause("security not found", nil, err)
		}
		return nil, err
	}
	return item, nil
}

func (s *investmentService) ListSecurities(ctx context.Context, _ string) ([]domain.Security, error) {
	return s.repo.ListSecurities(ctx)
}

func (s *investmentService) ListSecurityPrices(ctx context.Context, _ string, securityID string, from *string, to *string) ([]domain.SecurityPriceDaily, error) {
	id := strings.TrimSpace(securityID)
	if id == "" {
		return nil, ValidationError("securityId is required", map[string]any{"field": "securityId"})
	}
	if from != nil && strings.TrimSpace(*from) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*from)); err != nil {
			return nil, ValidationError("from is invalid", map[string]any{"field": "from"})
		}
	}
	if to != nil && strings.TrimSpace(*to) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*to)); err != nil {
			return nil, ValidationError("to is invalid", map[string]any{"field": "to"})
		}
	}
	return s.repo.ListSecurityPrices(ctx, id, normalizeOptionalString(from), normalizeOptionalString(to))
}

func (s *investmentService) ListSecurityEvents(ctx context.Context, _ string, securityID string, from *string, to *string) ([]domain.SecurityEvent, error) {
	id := strings.TrimSpace(securityID)
	if id == "" {
		return nil, ValidationError("securityId is required", map[string]any{"field": "securityId"})
	}
	if from != nil && strings.TrimSpace(*from) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*from)); err != nil {
			return nil, ValidationError("from is invalid", map[string]any{"field": "from"})
		}
	}
	if to != nil && strings.TrimSpace(*to) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*to)); err != nil {
			return nil, ValidationError("to is invalid", map[string]any{"field": "to"})
		}
	}
	return s.repo.ListSecurityEvents(ctx, id, normalizeOptionalString(from), normalizeOptionalString(to))
}

func (s *investmentService) CreateTrade(ctx context.Context, userID string, brokerAccountID string, req CreateTradeRequest) (*domain.Trade, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}

	ia, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		if errors.Is(err, domain.ErrInvestmentAccountNotFound) {
			return nil, NotFoundErrorWithCause("investment account not found", nil, err)
		}
		return nil, err
	}

	securityID := strings.TrimSpace(req.SecurityID)
	if securityID == "" {
		return nil, ValidationError("security_id is required", map[string]any{"field": "security_id"})
	}
	if _, err := s.repo.GetSecurity(ctx, securityID); err != nil {
		if errors.Is(err, domain.ErrSecurityNotFound) {
			return nil, NotFoundErrorWithCause("security not found", nil, err)
		}
		return nil, err
	}

	side := strings.TrimSpace(req.Side)
	if side != "buy" && side != "sell" {
		return nil, ValidationError("side is invalid", map[string]any{"field": "side"})
	}

	quantity := strings.TrimSpace(req.Quantity)
	if quantity == "" {
		return nil, ValidationError("quantity is required", map[string]any{"field": "quantity"})
	}
	if !isValidDecimal(quantity) {
		return nil, ValidationError("quantity must be a decimal string", map[string]any{"field": "quantity"})
	}

	price := strings.TrimSpace(req.Price)
	if price == "" {
		return nil, ValidationError("price is required", map[string]any{"field": "price"})
	}
	if !isValidDecimal(price) {
		return nil, ValidationError("price must be a decimal string", map[string]any{"field": "price"})
	}

	fees := "0"
	if req.Fees != nil {
		fees = strings.TrimSpace(*req.Fees)
		if fees == "" {
			fees = "0"
		}
	}
	if !isValidDecimal(fees) {
		return nil, ValidationError("fees must be a decimal string", map[string]any{"field": "fees"})
	}

	taxes := "0"
	if req.Taxes != nil {
		taxes = strings.TrimSpace(*req.Taxes)
		if taxes == "" {
			taxes = "0"
		}
	}
	if !isValidDecimal(taxes) {
		return nil, ValidationError("taxes must be a decimal string", map[string]any{"field": "taxes"})
	}

	occurredAt, _, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	feeTxID := normalizeOptionalString(req.FeeTransactionID)
	taxTxID := normalizeOptionalString(req.TaxTransactionID)

	// Auto-create fee/tax transactions if requested amounts > 0 and no explicit transaction ids provided.
	if feeTxID == nil {
		if amt, ok := new(big.Rat).SetString(fees); ok && amt.Cmp(new(big.Rat)) > 0 {
			desc := "Trade fee"
			if sec, err := s.repo.GetSecurity(ctx, securityID); err == nil {
				desc = "Trade fee: " + sec.Symbol
			}
			occurredDate := occurredAt.UTC().Format("2006-01-02")
			externalRef := deriveTradeExternalRef(req.ClientID, id, "fee")
			tx, err := s.tx.Create(ctx, userID, CreateTransactionRequest{
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
			occurredDate := occurredAt.UTC().Format("2006-01-02")
			externalRef := deriveTradeExternalRef(req.ClientID, id, "tax")
			tx, err := s.tx.Create(ctx, userID, CreateTransactionRequest{
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

	item := domain.Trade{
		ID:               id,
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

	if err := s.repo.CreateTrade(ctx, userID, item); err != nil {
		if errors.Is(err, domain.ErrInvestmentForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		return nil, err
	}

	// Return it by scanning list (MVP style).
	items, err := s.repo.ListTrades(ctx, userID, bid)
	if err != nil {
		return &item, nil
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return &item, nil
}

func (s *investmentService) ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]domain.Trade, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}
	return s.repo.ListTrades(ctx, userID, bid)
}

func (s *investmentService) ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]domain.Holding, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}
	return s.repo.ListHoldings(ctx, userID, bid)
}

func (s *investmentService) UpsertSecurityEventElection(ctx context.Context, userID string, brokerAccountID string, req UpsertSecurityEventElectionRequest) (*domain.SecurityEventElection, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}
	// Ensure access.
	_, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		if errors.Is(err, domain.ErrInvestmentAccountNotFound) {
			return nil, NotFoundErrorWithCause("investment account not found", nil, err)
		}
		return nil, err
	}

	eventID := strings.TrimSpace(req.SecurityEventID)
	if eventID == "" {
		return nil, ValidationError("security_event_id is required", map[string]any{"field": "security_event_id"})
	}
	event, err := s.repo.GetSecurityEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, domain.ErrSecurityEventNotFound) {
			return nil, NotFoundErrorWithCause("security event not found", nil, err)
		}
		return nil, err
	}

	elected := strings.TrimSpace(req.ElectedQuantity)
	if elected == "" {
		return nil, ValidationError("elected_quantity is required", map[string]any{"field": "elected_quantity"})
	}
	if !isValidDecimal(elected) {
		return nil, ValidationError("elected_quantity must be a decimal string", map[string]any{"field": "elected_quantity"})
	}

	status := "draft"
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
	}
	if status != "draft" && status != "confirmed" && status != "cancelled" {
		return nil, ValidationError("status is invalid", map[string]any{"field": "status"})
	}

	// Entitlement date precedence: ex_date -> record_date -> effective_date.
	entitlementDate := firstNonNilString(event.ExDate, event.RecordDate, event.EffectiveDate)
	if entitlementDate == "" {
		return nil, ValidationError("security event has no entitlement date", nil)
	}

	// Snapshot holding quantity using current holding row (MVP approximation).
	holdingQty := "0"
	if h, err := s.repo.GetHolding(ctx, userID, bid, event.SecurityID); err == nil {
		holdingQty = h.Quantity
	}

	entitled := computeEntitledQuantity(holdingQty, event.RatioNumerator, event.RatioDenominator)

	// Clamp/validate: elected <= entitled.
	if cmpDecimalStrings(elected, entitled) > 0 {
		return nil, ValidationError("elected_quantity must be <= entitled_quantity", map[string]any{"field": "elected_quantity"})
	}

	now := time.Now().UTC()
	id := uuid.NewString()
	var confirmedAt *time.Time
	if status == "confirmed" {
		confirmedAt = &now
	}

	item := domain.SecurityEventElection{
		ID:                           id,
		UserID:                       userID,
		BrokerAccountID:              bid,
		SecurityEventID:              eventID,
		SecurityID:                   event.SecurityID,
		EntitlementDate:              entitlementDate,
		HoldingQuantityAtEntitlement: holdingQty,
		EntitledQuantity:             entitled,
		ElectedQuantity:              elected,
		Status:                       status,
		ConfirmedAt:                  confirmedAt,
		Note:                         normalizeOptionalString(req.Note),
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}

	out, err := s.repo.UpsertSecurityEventElection(ctx, userID, item)
	if err != nil {
		if errors.Is(err, domain.ErrInvestmentForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		return nil, err
	}
	return out, nil
}

func (s *investmentService) ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]domain.SecurityEventElection, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, ValidationError("investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
	}
	if status != nil {
		v := strings.TrimSpace(*status)
		if v != "" && v != "draft" && v != "confirmed" && v != "cancelled" {
			return nil, ValidationError("status is invalid", map[string]any{"field": "status"})
		}
	}
	return s.repo.ListSecurityEventElections(ctx, userID, bid, normalizeOptionalString(status))
}

func deriveTradeExternalRef(clientID *string, tradeID string, kind string) *string {
	if kind != "fee" && kind != "tax" {
		return nil
	}
	if clientID != nil {
		v := strings.TrimSpace(*clientID)
		if v != "" {
			out := "trade:" + v + ":" + kind
			return &out
		}
	}
	out := "trade:" + tradeID + ":" + kind
	return &out
}

func firstNonNilString(a, b, c *string) string {
	if a != nil {
		v := strings.TrimSpace(*a)
		if v != "" {
			return v
		}
	}
	if b != nil {
		v := strings.TrimSpace(*b)
		if v != "" {
			return v
		}
	}
	if c != nil {
		v := strings.TrimSpace(*c)
		if v != "" {
			return v
		}
	}
	return ""
}

func computeEntitledQuantity(holdingQty string, ratioNum *string, ratioDen *string) string {
	// Default: entitled == holding
	if ratioNum == nil || ratioDen == nil {
		return holdingQty
	}
	ns := strings.TrimSpace(*ratioNum)
	ds := strings.TrimSpace(*ratioDen)
	if ns == "" || ds == "" {
		return holdingQty
	}
	q, ok := new(big.Rat).SetString(strings.TrimSpace(holdingQty))
	if !ok {
		return "0"
	}
	n, ok := new(big.Rat).SetString(ns)
	if !ok {
		return holdingQty
	}
	d, ok := new(big.Rat).SetString(ds)
	if !ok || d.Cmp(new(big.Rat)) == 0 {
		return holdingQty
	}
	q.Mul(q, n)
	q.Quo(q, d)
	// Persist as 8-decimal numeric string to match schema.
	return q.FloatString(8)
}

func cmpDecimalStrings(a, b string) int {
	ra, okA := new(big.Rat).SetString(strings.TrimSpace(a))
	rb, okB := new(big.Rat).SetString(strings.TrimSpace(b))
	if !okA || !okB {
		return 0
	}
	return ra.Cmp(rb)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
