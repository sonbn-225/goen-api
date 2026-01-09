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
	AccountID       *string `json:"account_id,omitempty"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
	BrokerName      *string `json:"broker_name,omitempty"`
	SyncEnabled     *bool   `json:"sync_enabled,omitempty"`
	SyncSettings    any     `json:"sync_settings,omitempty"`
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
	if accountID == "" && parentAccountID == nil {
		return nil, errors.New("either account_id or parent_account_id is required")
	}

	var acc *domain.Account
	if accountID == "" {
		parent, err := s.accounts.GetAccount(ctx, userID, *parentAccountID)
		if err != nil {
			return nil, err
		}
		if parent.AccountType != "bank" && parent.AccountType != "wallet" {
			return nil, errors.New("parent account must be bank or wallet")
		}

		createdAcc, err := s.accounts.CreateAccount(ctx, userID, CreateAccountRequest{
			Name:            "Broker",
			AccountType:     "broker",
			Currency:        parent.Currency,
			ParentAccountID: parentAccountID,
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
		return nil, errors.New("account_id must be an account of type broker")
	}

	syncEnabled := false
	if req.SyncEnabled != nil {
		syncEnabled = *req.SyncEnabled
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	item := domain.InvestmentAccount{
		ID:           id,
		AccountID:    accountID,
		BrokerName:   normalizeOptionalString(req.BrokerName),
		Currency:     acc.Currency,
		SyncEnabled:  syncEnabled,
		SyncSettings: req.SyncSettings,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateInvestmentAccount(ctx, userID, item); err != nil {
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
		return nil, errors.New("investmentAccountId is required")
	}
	return s.repo.GetInvestmentAccount(ctx, userID, id)
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
		brokerName := strings.TrimSpace(acc.Name)
		var brokerNamePtr *string
		if brokerName != "" {
			brokerNamePtr = &brokerName
		}

		item := domain.InvestmentAccount{
			ID:          uuid.NewString(),
			AccountID:   acc.ID,
			BrokerName:  brokerNamePtr,
			Currency:    acc.Currency,
			SyncEnabled: false,
			CreatedAt:   now,
			UpdatedAt:   now,
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
		return nil, errors.New("securityId is required")
	}
	return s.repo.GetSecurity(ctx, id)
}

func (s *investmentService) ListSecurities(ctx context.Context, _ string) ([]domain.Security, error) {
	return s.repo.ListSecurities(ctx)
}

func (s *investmentService) ListSecurityPrices(ctx context.Context, _ string, securityID string, from *string, to *string) ([]domain.SecurityPriceDaily, error) {
	id := strings.TrimSpace(securityID)
	if id == "" {
		return nil, errors.New("securityId is required")
	}
	if from != nil && strings.TrimSpace(*from) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*from)); err != nil {
			return nil, errors.New("from is invalid")
		}
	}
	if to != nil && strings.TrimSpace(*to) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*to)); err != nil {
			return nil, errors.New("to is invalid")
		}
	}
	return s.repo.ListSecurityPrices(ctx, id, normalizeOptionalString(from), normalizeOptionalString(to))
}

func (s *investmentService) ListSecurityEvents(ctx context.Context, _ string, securityID string, from *string, to *string) ([]domain.SecurityEvent, error) {
	id := strings.TrimSpace(securityID)
	if id == "" {
		return nil, errors.New("securityId is required")
	}
	if from != nil && strings.TrimSpace(*from) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*from)); err != nil {
			return nil, errors.New("from is invalid")
		}
	}
	if to != nil && strings.TrimSpace(*to) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*to)); err != nil {
			return nil, errors.New("to is invalid")
		}
	}
	return s.repo.ListSecurityEvents(ctx, id, normalizeOptionalString(from), normalizeOptionalString(to))
}

func (s *investmentService) CreateTrade(ctx context.Context, userID string, brokerAccountID string, req CreateTradeRequest) (*domain.Trade, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, errors.New("investmentAccountId is required")
	}

	ia, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	securityID := strings.TrimSpace(req.SecurityID)
	if securityID == "" {
		return nil, errors.New("security_id is required")
	}
	if _, err := s.repo.GetSecurity(ctx, securityID); err != nil {
		return nil, err
	}

	side := strings.TrimSpace(req.Side)
	if side != "buy" && side != "sell" {
		return nil, errors.New("side is invalid")
	}

	quantity := strings.TrimSpace(req.Quantity)
	if quantity == "" {
		return nil, errors.New("quantity is required")
	}
	if !isValidDecimal(quantity) {
		return nil, errors.New("quantity must be a decimal string")
	}

	price := strings.TrimSpace(req.Price)
	if price == "" {
		return nil, errors.New("price is required")
	}
	if !isValidDecimal(price) {
		return nil, errors.New("price must be a decimal string")
	}

	fees := "0"
	if req.Fees != nil {
		fees = strings.TrimSpace(*req.Fees)
		if fees == "" {
			fees = "0"
		}
	}
	if !isValidDecimal(fees) {
		return nil, errors.New("fees must be a decimal string")
	}

	taxes := "0"
	if req.Taxes != nil {
		taxes = strings.TrimSpace(*req.Taxes)
		if taxes == "" {
			taxes = "0"
		}
	}
	if !isValidDecimal(taxes) {
		return nil, errors.New("taxes must be a decimal string")
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
		return nil, errors.New("investmentAccountId is required")
	}
	return s.repo.ListTrades(ctx, userID, bid)
}

func (s *investmentService) ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]domain.Holding, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, errors.New("investmentAccountId is required")
	}
	return s.repo.ListHoldings(ctx, userID, bid)
}

func (s *investmentService) UpsertSecurityEventElection(ctx context.Context, userID string, brokerAccountID string, req UpsertSecurityEventElectionRequest) (*domain.SecurityEventElection, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, errors.New("investmentAccountId is required")
	}
	// Ensure access.
	_, err := s.repo.GetInvestmentAccount(ctx, userID, bid)
	if err != nil {
		return nil, err
	}

	eventID := strings.TrimSpace(req.SecurityEventID)
	if eventID == "" {
		return nil, errors.New("security_event_id is required")
	}
	event, err := s.repo.GetSecurityEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	elected := strings.TrimSpace(req.ElectedQuantity)
	if elected == "" {
		return nil, errors.New("elected_quantity is required")
	}
	if !isValidDecimal(elected) {
		return nil, errors.New("elected_quantity must be a decimal string")
	}

	status := "draft"
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
	}
	if status != "draft" && status != "confirmed" && status != "cancelled" {
		return nil, errors.New("status is invalid")
	}

	// Entitlement date precedence: ex_date -> record_date -> effective_date.
	entitlementDate := firstNonNilString(event.ExDate, event.RecordDate, event.EffectiveDate)
	if entitlementDate == "" {
		return nil, errors.New("security event has no entitlement date")
	}

	// Snapshot holding quantity using current holding row (MVP approximation).
	holdingQty := "0"
	if h, err := s.repo.GetHolding(ctx, userID, bid, event.SecurityID); err == nil {
		holdingQty = h.Quantity
	}

	entitled := computeEntitledQuantity(holdingQty, event.RatioNumerator, event.RatioDenominator)

	// Clamp/validate: elected <= entitled.
	if cmpDecimalStrings(elected, entitled) > 0 {
		return nil, errors.New("elected_quantity must be <= entitled_quantity")
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

	return s.repo.UpsertSecurityEventElection(ctx, userID, item)
}

func (s *investmentService) ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]domain.SecurityEventElection, error) {
	bid := strings.TrimSpace(brokerAccountID)
	if bid == "" {
		return nil, errors.New("investmentAccountId is required")
	}
	if status != nil {
		v := strings.TrimSpace(*status)
		if v != "" && v != "draft" && v != "confirmed" && v != "cancelled" {
			return nil, errors.New("status is invalid")
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
