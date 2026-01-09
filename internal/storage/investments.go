package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type InvestmentRepo struct {
	db *Postgres
}

func NewInvestmentRepo(db *Postgres) *InvestmentRepo {
	return &InvestmentRepo{db: db}
}

func (r *InvestmentRepo) CreateInvestmentAccount(ctx context.Context, userID string, ia domain.InvestmentAccount) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Only allow creating extension for an accessible broker account.
	// Require write permission (owner/editor) because this is a write action.
	cmd, err := pool.Exec(ctx, `
		INSERT INTO investment_accounts (
			id, account_id, broker_name, sync_enabled, sync_settings, created_at, updated_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7
		WHERE EXISTS (
			SELECT 1
			FROM accounts a
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE a.id = $2
			  AND a.account_type = 'broker'
			  AND a.deleted_at IS NULL
			  AND ua.user_id = $9
			  AND ua.status = 'active'
			  AND ua.permission IN ('owner','editor')
		)
	`, ia.ID, ia.AccountID, ia.BrokerName, ia.SyncEnabled, ia.SyncSettings, ia.CreatedAt, ia.UpdatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*domain.InvestmentAccount, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT ia.id, ia.account_id, ia.broker_name, a.currency, ia.sync_enabled, ia.sync_settings, ia.created_at, ia.updated_at
		FROM investment_accounts ia
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ia.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, investmentAccountID, userID)

	var out domain.InvestmentAccount
	var syncSettings any
	if err := row.Scan(&out.ID, &out.AccountID, &out.BrokerName, &out.Currency, &out.SyncEnabled, &syncSettings, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvestmentAccountNotFound
		}
		return nil, err
	}
	if syncSettings != nil {
		out.SyncSettings = syncSettings
	}
	return &out, nil
}

func (r *InvestmentRepo) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT ia.id, ia.account_id, ia.broker_name, a.currency, ia.sync_enabled, ia.sync_settings, ia.created_at, ia.updated_at
		FROM investment_accounts ia
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY ia.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.InvestmentAccount{}
	for rows.Next() {
		var ia domain.InvestmentAccount
		var syncSettings any
		if err := rows.Scan(&ia.ID, &ia.AccountID, &ia.BrokerName, &ia.Currency, &ia.SyncEnabled, &syncSettings, &ia.CreatedAt, &ia.UpdatedAt); err != nil {
			return nil, err
		}
		if syncSettings != nil {
			ia.SyncSettings = syncSettings
		}
		out = append(out, ia)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) CreateSecurity(ctx context.Context, s domain.Security) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO securities (id, symbol, name, asset_class, currency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, s.ID, s.Symbol, s.Name, s.AssetClass, s.Currency, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *InvestmentRepo) GetSecurity(ctx context.Context, securityID string) (*domain.Security, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, symbol, name, asset_class, currency, created_at, updated_at
		FROM securities
		WHERE id = $1
	`, securityID)

	var s domain.Security
	if err := row.Scan(&s.ID, &s.Symbol, &s.Name, &s.AssetClass, &s.Currency, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSecurityNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *InvestmentRepo) ListSecurities(ctx context.Context) ([]domain.Security, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, symbol, name, asset_class, currency, created_at, updated_at
		FROM securities
		ORDER BY symbol ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Security{}
	for rows.Next() {
		var s domain.Security
		if err := rows.Scan(&s.ID, &s.Symbol, &s.Name, &s.AssetClass, &s.Currency, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]domain.SecurityPriceDaily, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, security_id, price_date, open, high, low, close, volume, created_at, updated_at
		FROM security_price_dailies
		WHERE security_id = $1
		  AND ($2::date IS NULL OR price_date >= $2::date)
		  AND ($3::date IS NULL OR price_date <= $3::date)
		ORDER BY price_date ASC
	`, securityID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.SecurityPriceDaily{}
	for rows.Next() {
		var p domain.SecurityPriceDaily
		var priceDate time.Time
		var open, high, low, close, volume sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.SecurityID,
			&priceDate,
			&open,
			&high,
			&low,
			&close,
			&volume,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.PriceDate = priceDate.Format("2006-01-02")
		p.Open = nullStringPtr(open)
		p.High = nullStringPtr(high)
		p.Low = nullStringPtr(low)
		if close.Valid {
			p.Close = close.String
		}
		p.Volume = nullStringPtr(volume)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]domain.SecurityEvent, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, security_id, event_type, ex_date, record_date, pay_date, effective_date,
		       cash_amount_per_share, ratio_numerator, ratio_denominator, subscription_price,
		       currency, vnstock_event_id, note, created_at, updated_at
		FROM security_events
		WHERE security_id = $1
		  AND ($2::date IS NULL OR COALESCE(effective_date, ex_date, record_date, pay_date) >= $2::date)
		  AND ($3::date IS NULL OR COALESCE(effective_date, ex_date, record_date, pay_date) <= $3::date)
		ORDER BY COALESCE(effective_date, ex_date, record_date, pay_date) ASC, created_at ASC
	`, securityID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.SecurityEvent{}
	for rows.Next() {
		var e domain.SecurityEvent
		var exDate, recordDate, payDate, effectiveDate sql.NullTime
		var cashPerShare, ratioNum, ratioDen, subPrice sql.NullString
		if err := rows.Scan(
			&e.ID,
			&e.SecurityID,
			&e.EventType,
			&exDate,
			&recordDate,
			&payDate,
			&effectiveDate,
			&cashPerShare,
			&ratioNum,
			&ratioDen,
			&subPrice,
			&e.Currency,
			&e.VnstockEventID,
			&e.Note,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}

		e.ExDate = nullTimeToDatePtr(exDate)
		e.RecordDate = nullTimeToDatePtr(recordDate)
		e.PayDate = nullTimeToDatePtr(payDate)
		e.EffectiveDate = nullTimeToDatePtr(effectiveDate)
		e.CashAmountPerShare = nullStringPtr(cashPerShare)
		e.RatioNumerator = nullStringPtr(ratioNum)
		e.RatioDenominator = nullStringPtr(ratioDen)
		e.SubscriptionPrice = nullStringPtr(subPrice)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) GetSecurityEvent(ctx context.Context, securityEventID string) (*domain.SecurityEvent, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, security_id, event_type, ex_date, record_date, pay_date, effective_date,
		       cash_amount_per_share, ratio_numerator, ratio_denominator, subscription_price,
		       currency, vnstock_event_id, note, created_at, updated_at
		FROM security_events
		WHERE id = $1
	`, securityEventID)

	var e domain.SecurityEvent
	var exDate, recordDate, payDate, effectiveDate sql.NullTime
	var cashPerShare, ratioNum, ratioDen, subPrice sql.NullString
	if err := row.Scan(
		&e.ID,
		&e.SecurityID,
		&e.EventType,
		&exDate,
		&recordDate,
		&payDate,
		&effectiveDate,
		&cashPerShare,
		&ratioNum,
		&ratioDen,
		&subPrice,
		&e.Currency,
		&e.VnstockEventID,
		&e.Note,
		&e.CreatedAt,
		&e.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSecurityEventNotFound
		}
		return nil, err
	}
	e.ExDate = nullTimeToDatePtr(exDate)
	e.RecordDate = nullTimeToDatePtr(recordDate)
	e.PayDate = nullTimeToDatePtr(payDate)
	e.EffectiveDate = nullTimeToDatePtr(effectiveDate)
	e.CashAmountPerShare = nullStringPtr(cashPerShare)
	e.RatioNumerator = nullStringPtr(ratioNum)
	e.RatioDenominator = nullStringPtr(ratioDen)
	e.SubscriptionPrice = nullStringPtr(subPrice)
	return &e, nil
}

func (r *InvestmentRepo) UpsertSecurityEventElection(ctx context.Context, userID string, e domain.SecurityEventElection) (*domain.SecurityEventElection, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the broker account belongs to the user (read access is enough for upsert itself; but elections are write).
	// Require write permission on underlying broker cash ledger account.
	row := pool.QueryRow(ctx, `
		WITH ok AS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
		INSERT INTO security_event_elections (
			id, user_id, broker_account_id, security_event_id, security_id,
			entitlement_date, holding_quantity_at_entitlement_date, entitled_quantity, elected_quantity,
			status, confirmed_at, note, created_at, updated_at
		)
		SELECT $3,$4,$1,$5,$6,$7,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13,$14,$15
		WHERE EXISTS (SELECT 1 FROM ok)
		ON CONFLICT (broker_account_id, security_event_id)
		DO UPDATE SET
			elected_quantity = EXCLUDED.elected_quantity,
			status = EXCLUDED.status,
			confirmed_at = EXCLUDED.confirmed_at,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, broker_account_id, security_event_id, security_id,
		          entitlement_date, holding_quantity_at_entitlement_date::text, entitled_quantity::text, elected_quantity::text,
		          status, confirmed_at, note, created_at, updated_at
	`, e.BrokerAccountID, userID, e.ID, e.UserID, e.SecurityEventID, e.SecurityID, e.EntitlementDate, e.HoldingQuantityAtEntitlement, e.EntitledQuantity, e.ElectedQuantity, e.Status, e.ConfirmedAt, e.Note, e.CreatedAt, e.UpdatedAt)

	var out domain.SecurityEventElection
	var holdingQty, entitledQty, electedQty string
	if err := row.Scan(
		&out.ID,
		&out.UserID,
		&out.BrokerAccountID,
		&out.SecurityEventID,
		&out.SecurityID,
		&out.EntitlementDate,
		&holdingQty,
		&entitledQty,
		&electedQty,
		&out.Status,
		&out.ConfirmedAt,
		&out.Note,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvestmentForbidden
		}
		return nil, err
	}
	out.HoldingQuantityAtEntitlement = holdingQty
	out.EntitledQuantity = entitledQty
	out.ElectedQuantity = electedQty
	return &out, nil
}

func (r *InvestmentRepo) ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]domain.SecurityEventElection, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	normStatus := normalizeOptionalString(status)
	rows, err := pool.Query(ctx, `
		SELECT e.id, e.user_id, e.broker_account_id, e.security_event_id, e.security_id,
		       e.entitlement_date, e.holding_quantity_at_entitlement_date::text, e.entitled_quantity::text, e.elected_quantity::text,
		       e.status, e.confirmed_at, e.note, e.created_at, e.updated_at
		FROM security_event_elections e
		JOIN investment_accounts ia ON ia.id = e.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE e.broker_account_id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		  AND ($3::text IS NULL OR e.status = $3)
		ORDER BY e.updated_at DESC
	`, brokerAccountID, userID, normStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.SecurityEventElection{}
	for rows.Next() {
		var e domain.SecurityEventElection
		var holdingQty, entitledQty, electedQty string
		if err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.BrokerAccountID,
			&e.SecurityEventID,
			&e.SecurityID,
			&e.EntitlementDate,
			&holdingQty,
			&entitledQty,
			&electedQty,
			&e.Status,
			&e.ConfirmedAt,
			&e.Note,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		e.HoldingQuantityAtEntitlement = holdingQty
		e.EntitledQuantity = entitledQty
		e.ElectedQuantity = electedQty
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) CreateTrade(ctx context.Context, userID string, t domain.Trade) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Write requires owner/editor on underlying broker cash ledger.
	cmd, err := pool.Exec(ctx, `
		INSERT INTO trades (
			id, client_id, broker_account_id, security_id, fee_transaction_id, tax_transaction_id,
			side, quantity, price, fees, taxes, occurred_at, note, created_at, updated_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12,$13,$14,$15
		WHERE EXISTS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $3 AND ua.user_id = $16 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, t.ID, t.ClientID, t.BrokerAccountID, t.SecurityID, t.FeeTransactionID, t.TaxTransactionID, t.Side, t.Quantity, t.Price, t.Fees, t.Taxes, t.OccurredAt, t.Note, t.CreatedAt, t.UpdatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]domain.Trade, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT t.id, t.client_id, t.broker_account_id, t.security_id, t.fee_transaction_id, t.tax_transaction_id,
		       t.side, t.quantity::text, t.price::text, t.fees::text, t.taxes::text, t.occurred_at,
		       t.note, t.created_at, t.updated_at
		FROM trades t
		JOIN investment_accounts ia ON ia.id = t.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.broker_account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY t.occurred_at DESC, t.created_at DESC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Trade{}
	for rows.Next() {
		var t domain.Trade
		if err := rows.Scan(
			&t.ID,
			&t.ClientID,
			&t.BrokerAccountID,
			&t.SecurityID,
			&t.FeeTransactionID,
			&t.TaxTransactionID,
			&t.Side,
			&t.Quantity,
			&t.Price,
			&t.Fees,
			&t.Taxes,
			&t.OccurredAt,
			&t.Note,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]domain.Holding, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT h.id, h.broker_account_id, h.security_id, h.quantity::text,
		       h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, h.market_value::text, h.unrealized_pnl::text,
		       h.as_of, h.source_of_truth, h.created_at, h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY h.updated_at DESC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Holding{}
	for rows.Next() {
		var h domain.Holding
		var costBasisTotal, avgCost, marketPrice, marketValue, unrealizedPnL sql.NullString
		if err := rows.Scan(
			&h.ID,
			&h.BrokerAccountID,
			&h.SecurityID,
			&h.Quantity,
			&costBasisTotal,
			&avgCost,
			&marketPrice,
			&marketValue,
			&unrealizedPnL,
			&h.AsOf,
			&h.SourceOfTruth,
			&h.CreatedAt,
			&h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		h.CostBasisTotal = nullStringPtr(costBasisTotal)
		h.AvgCost = nullStringPtr(avgCost)
		h.MarketPrice = nullStringPtr(marketPrice)
		h.MarketValue = nullStringPtr(marketValue)
		h.UnrealizedPnL = nullStringPtr(unrealizedPnL)
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) GetHolding(ctx context.Context, userID string, brokerAccountID string, securityID string) (*domain.Holding, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT h.id, h.broker_account_id, h.security_id, h.quantity::text,
		       h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, h.market_value::text, h.unrealized_pnl::text,
		       h.as_of, h.source_of_truth, h.created_at, h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1 AND h.security_id = $2
		  AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, brokerAccountID, securityID, userID)

	var h domain.Holding
	var costBasisTotal, avgCost, marketPrice, marketValue, unrealizedPnL sql.NullString
	if err := row.Scan(
		&h.ID,
		&h.BrokerAccountID,
		&h.SecurityID,
		&h.Quantity,
		&costBasisTotal,
		&avgCost,
		&marketPrice,
		&marketValue,
		&unrealizedPnL,
		&h.AsOf,
		&h.SourceOfTruth,
		&h.CreatedAt,
		&h.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHoldingNotFound
		}
		return nil, err
	}

	h.CostBasisTotal = nullStringPtr(costBasisTotal)
	h.AvgCost = nullStringPtr(avgCost)
	h.MarketPrice = nullStringPtr(marketPrice)
	h.MarketValue = nullStringPtr(marketValue)
	h.UnrealizedPnL = nullStringPtr(unrealizedPnL)
	return &h, nil
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := strings.TrimSpace(ns.String)
	if v == "" {
		return nil
	}
	return &v
}

func nullTimeToDatePtr(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	v := nt.Time.UTC().Format("2006-01-02")
	return &v
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
