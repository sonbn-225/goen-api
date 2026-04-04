package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type InvestmentRepo struct {
	db *database.Postgres
}

func NewInvestmentRepo(db *database.Postgres) *InvestmentRepo {
	return &InvestmentRepo{db: db}
}

func (r *InvestmentRepo) GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*entity.InvestmentAccount, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT ia.id, ia.account_id, a.currency, ia.fee_settings, ia.tax_settings, ia.created_at, ia.updated_at
		FROM investment_accounts ia
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ia.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, investmentAccountID, userID)

	var out entity.InvestmentAccount
	if err := row.Scan(&out.ID, &out.AccountID, &out.Currency, &out.FeeSettings, &out.TaxSettings, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("investment account not found")
		}
		return nil, err
	}
	return &out, nil
}

func (r *InvestmentRepo) ListInvestmentAccounts(ctx context.Context, userID string) ([]entity.InvestmentAccount, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT ia.id, ia.account_id, a.currency, ia.fee_settings, ia.tax_settings, ia.created_at, ia.updated_at
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

	var results []entity.InvestmentAccount
	for rows.Next() {
		var ia entity.InvestmentAccount
		if err := rows.Scan(&ia.ID, &ia.AccountID, &ia.Currency, &ia.FeeSettings, &ia.TaxSettings, &ia.CreatedAt, &ia.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, ia)
	}
	return results, nil
}

func (r *InvestmentRepo) UpdateInvestmentAccountSettings(ctx context.Context, userID string, investmentAccountID string, feeSettings any, taxSettings any) (*entity.InvestmentAccount, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		WITH ok AS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
		UPDATE investment_accounts
		SET fee_settings = COALESCE($3, fee_settings),
		    tax_settings = COALESCE($4, tax_settings),
		    updated_at = NOW()
		WHERE id = $1 AND EXISTS (SELECT 1 FROM ok)
		RETURNING id, account_id, fee_settings, tax_settings, created_at, updated_at
	`, investmentAccountID, userID, feeSettings, taxSettings)

	var out entity.InvestmentAccount
	if err := row.Scan(&out.ID, &out.AccountID, &out.FeeSettings, &out.TaxSettings, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("forbidden: account access required")
		}
		return nil, err
	}

	// Refetch to get currency
	return r.GetInvestmentAccount(ctx, userID, out.ID)
}

func (r *InvestmentRepo) GetSecurity(ctx context.Context, securityID string) (*entity.Security, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, symbol, name, asset_class, currency, created_at, updated_at
		FROM securities
		WHERE id = $1
	`, securityID)

	var s entity.Security
	if err := row.Scan(&s.ID, &s.Symbol, &s.Name, &s.AssetClass, &s.Currency, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("security not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *InvestmentRepo) ListSecurities(ctx context.Context) ([]entity.Security, error) {
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

	var results []entity.Security
	for rows.Next() {
		var s entity.Security
		if err := rows.Scan(&s.ID, &s.Symbol, &s.Name, &s.AssetClass, &s.Currency, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *InvestmentRepo) ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]entity.SecurityPriceDaily, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, security_id, price_date,
		       open * 1000, high * 1000, low * 1000, close * 1000,
		       volume, created_at, updated_at
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

	var results []entity.SecurityPriceDaily
	for rows.Next() {
		var p entity.SecurityPriceDaily
		var priceDate time.Time
		var open, high, low, volume sql.NullString
		if err := rows.Scan(
			&p.ID, &p.SecurityID, &priceDate, &open, &high, &low, &p.Close, &volume, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.PriceDate = priceDate.Format("2006-01-02")
		if open.Valid {
			p.Open = &open.String
		}
		if high.Valid {
			p.High = &high.String
		}
		if low.Valid {
			p.Low = &low.String
		}
		if volume.Valid {
			p.Volume = &volume.String
		}
		results = append(results, p)
	}
	return results, nil
}

func (r *InvestmentRepo) ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]entity.SecurityEvent, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, security_id, event_type, ex_date, record_date, pay_date, effective_date,
		       cash_amount_per_share, ratio_numerator, ratio_denominator, subscription_price,
		       currency, note, created_at, updated_at
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

	var results []entity.SecurityEvent
	for rows.Next() {
		var e entity.SecurityEvent
		var exDate, recordDate, payDate, effectiveDate sql.NullTime
		var cashPerShare, ratioNum, ratioDen, subPrice sql.NullString
		if err := rows.Scan(
			&e.ID, &e.SecurityID, &e.EventType, &exDate, &recordDate, &payDate, &effectiveDate,
			&cashPerShare, &ratioNum, &ratioDen, &subPrice, &e.Currency, &e.Note, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if exDate.Valid {
			d := exDate.Time.Format("2006-01-02")
			e.ExDate = &d
		}
		if recordDate.Valid {
			d := recordDate.Time.Format("2006-01-02")
			e.RecordDate = &d
		}
		if payDate.Valid {
			d := payDate.Time.Format("2006-01-02")
			e.PayDate = &d
		}
		if effectiveDate.Valid {
			d := effectiveDate.Time.Format("2006-01-02")
			e.EffectiveDate = &d
		}
		if cashPerShare.Valid {
			e.CashAmountPerShare = &cashPerShare.String
		}
		if ratioNum.Valid {
			e.RatioNumerator = &ratioNum.String
		}
		if ratioDen.Valid {
			e.RatioDenominator = &ratioDen.String
		}
		if subPrice.Valid {
			e.SubscriptionPrice = &subPrice.String
		}
		results = append(results, e)
	}
	return results, nil
}

func (r *InvestmentRepo) GetSecurityEvent(ctx context.Context, securityEventID string) (*entity.SecurityEvent, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, security_id, event_type, ex_date, record_date, pay_date, effective_date,
		       cash_amount_per_share, ratio_numerator, ratio_denominator, subscription_price,
		       currency, note, created_at, updated_at
		FROM security_events
		WHERE id = $1
	`, securityEventID)

	var e entity.SecurityEvent
	var exDate, recordDate, payDate, effectiveDate sql.NullTime
	var cashPerShare, ratioNum, ratioDen, subPrice sql.NullString
	if err := row.Scan(
		&e.ID, &e.SecurityID, &e.EventType, &exDate, &recordDate, &payDate, &effectiveDate,
		&cashPerShare, &ratioNum, &ratioDen, &subPrice, &e.Currency, &e.Note, &e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("security event not found")
		}
		return nil, err
	}
	if exDate.Valid {
		d := exDate.Time.Format("2006-01-02")
		e.ExDate = &d
	}
	if recordDate.Valid {
		d := recordDate.Time.Format("2006-01-02")
		e.RecordDate = &d
	}
	if payDate.Valid {
		d := payDate.Time.Format("2006-01-02")
		e.PayDate = &d
	}
	if effectiveDate.Valid {
		d := effectiveDate.Time.Format("2006-01-02")
		e.EffectiveDate = &d
	}
	if cashPerShare.Valid {
		e.CashAmountPerShare = &cashPerShare.String
	}
	if ratioNum.Valid {
		e.RatioNumerator = &ratioNum.String
	}
	if ratioDen.Valid {
		e.RatioDenominator = &ratioDen.String
	}
	if subPrice.Valid {
		e.SubscriptionPrice = &subPrice.String
	}
	return &e, nil
}

func (r *InvestmentRepo) UpsertSecurityEventElection(ctx context.Context, userID string, e entity.SecurityEventElection) (*entity.SecurityEventElection, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
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

	var out entity.SecurityEventElection
	if err := row.Scan(
		&out.ID, &out.UserID, &out.BrokerAccountID, &out.SecurityEventID, &out.SecurityID,
		&out.EntitlementDate, &out.HoldingQuantityAtEntitlement, &out.EntitledQuantity, &out.ElectedQuantity,
		&out.Status, &out.ConfirmedAt, &out.Note, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("forbidden: account access required")
		}
		return nil, err
	}
	return &out, nil
}

func (r *InvestmentRepo) ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]entity.SecurityEventElection, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

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
	`, brokerAccountID, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.SecurityEventElection
	for rows.Next() {
		var e entity.SecurityEventElection
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.BrokerAccountID, &e.SecurityEventID, &e.SecurityID,
			&e.EntitlementDate, &e.HoldingQuantityAtEntitlement, &e.EntitledQuantity, &e.ElectedQuantity,
			&e.Status, &e.ConfirmedAt, &e.Note, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	return results, nil
}

func (r *InvestmentRepo) CreateTrade(ctx context.Context, userID string, t entity.Trade) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

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
		return errors.New("forbidden: account access required")
	}
	return nil
}

func (r *InvestmentRepo) GetTrade(ctx context.Context, userID string, tradeID string) (*entity.Trade, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT t.id, t.client_id, t.broker_account_id, t.security_id,
		       t.fee_transaction_id, t.tax_transaction_id,
		       t.side, t.quantity::text, t.price::text, t.fees::text, t.taxes::text,
		       t.occurred_at, t.note, t.created_at, t.updated_at
		FROM trades t
		JOIN investment_accounts ia ON ia.id = t.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, tradeID, userID)

	var t entity.Trade
	if err := row.Scan(
		&t.ID, &t.ClientID, &t.BrokerAccountID, &t.SecurityID, &t.FeeTransactionID, &t.TaxTransactionID,
		&t.Side, &t.Quantity, &t.Price, &t.Fees, &t.Taxes, &t.OccurredAt, &t.Note, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trade not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *InvestmentRepo) ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]entity.Trade, error) {
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

	var results []entity.Trade
	for rows.Next() {
		var t entity.Trade
		if err := rows.Scan(
			&t.ID, &t.ClientID, &t.BrokerAccountID, &t.SecurityID, &t.FeeTransactionID, &t.TaxTransactionID,
			&t.Side, &t.Quantity, &t.Price, &t.Fees, &t.Taxes, &t.OccurredAt, &t.Note, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (r *InvestmentRepo) DeleteTrade(ctx context.Context, userID string, tradeID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		DELETE FROM trades
		WHERE id = $1 AND EXISTS (
			SELECT 1
			FROM trades t
			JOIN investment_accounts ia ON ia.id = t.broker_account_id
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE t.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, tradeID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("forbidden: account access required")
	}
	return nil
}

func (r *InvestmentRepo) ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			h.id, h.broker_account_id, h.security_id, h.quantity::text, 
			h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, 
			h.market_value::text, h.unrealized_pnl::text, h.as_of, h.created_at, h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY h.security_id ASC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Holding
	for rows.Next() {
		var h entity.Holding
		if err := rows.Scan(
			&h.ID, &h.BrokerAccountID, &h.SecurityID, &h.Quantity, &h.CostBasisTotal, &h.AvgCost,
			&h.MarketPrice, &h.MarketValue, &h.UnrealizedPnL, &h.AsOf, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, nil
}

func (r *InvestmentRepo) GetHolding(ctx context.Context, userID string, brokerAccountID string, securityID string) (*entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT 
			h.id, h.broker_account_id, h.security_id, h.quantity::text, 
			h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, 
			h.market_value::text, h.unrealized_pnl::text, h.as_of, h.created_at, h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1 AND h.security_id = $2 AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, brokerAccountID, securityID, userID)

	var h entity.Holding
	if err := row.Scan(
		&h.ID, &h.BrokerAccountID, &h.SecurityID, &h.Quantity, &h.CostBasisTotal, &h.AvgCost,
		&h.MarketPrice, &h.MarketValue, &h.UnrealizedPnL, &h.AsOf, &h.CreatedAt, &h.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

func (r *InvestmentRepo) UpsertHolding(ctx context.Context, userID string, h entity.Holding) (*entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	row := pool.QueryRow(ctx, `
		INSERT INTO holdings (
			id, broker_account_id, security_id, quantity, cost_basis_total, avg_cost, 
			market_price, market_value, unrealized_pnl, as_of, created_at, updated_at
		) VALUES ($1,$2,$3,$4::numeric,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10,$11,$12)
		ON CONFLICT (broker_account_id, security_id) 
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			cost_basis_total = EXCLUDED.cost_basis_total,
			avg_cost = EXCLUDED.avg_cost,
			market_price = EXCLUDED.market_price,
			market_value = EXCLUDED.market_value,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			as_of = EXCLUDED.as_of,
			updated_at = EXCLUDED.updated_at
		RETURNING id, broker_account_id, security_id, quantity::text, cost_basis_total::text, avg_cost::text, market_price::text, market_value::text, unrealized_pnl::text, as_of, created_at, updated_at
	`, h.ID, h.BrokerAccountID, h.SecurityID, h.Quantity, h.CostBasisTotal, h.AvgCost, h.MarketPrice, h.MarketValue, h.UnrealizedPnL, h.AsOf, h.CreatedAt, h.UpdatedAt)

	var out entity.Holding
	if err := row.Scan(
		&out.ID, &out.BrokerAccountID, &out.SecurityID, &out.Quantity, &out.CostBasisTotal, &out.AvgCost,
		&out.MarketPrice, &out.MarketValue, &out.UnrealizedPnL, &out.AsOf, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvestmentRepo) ListShareLots(ctx context.Context, userID string, brokerAccountID string, securityID string) ([]entity.ShareLot, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			l.id, l.broker_account_id, l.security_id, l.quantity::text, 
			l.acquisition_date, l.cost_basis_per_share::text, l.provenance, 
			l.status, l.buy_trade_id, l.created_at, l.updated_at
		FROM share_lots l
		JOIN investment_accounts ia ON ia.id = l.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.broker_account_id = $1 AND l.security_id = $2 AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY l.acquisition_date ASC, l.created_at ASC
	`, brokerAccountID, securityID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.ShareLot
	for rows.Next() {
		var l entity.ShareLot
		var acqDate time.Time
		if err := rows.Scan(
			&l.ID, &l.BrokerAccountID, &l.SecurityID, &l.Quantity, &acqDate,
			&l.CostBasisPer, &l.Provenance, &l.Status, &l.BuyTradeID, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.AcquisitionDate = acqDate.Format("2006-01-02")
		results = append(results, l)
	}
	return results, nil
}

func (r *InvestmentRepo) CreateShareLot(ctx context.Context, userID string, lot entity.ShareLot) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO share_lots (
			id, broker_account_id, security_id, quantity, acquisition_date, 
			cost_basis_per_share, provenance, status, buy_trade_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4::numeric,$5,$6::numeric,$7,$8,$9,$10,$11)
	`, lot.ID, lot.BrokerAccountID, lot.SecurityID, lot.Quantity, lot.AcquisitionDate, lot.CostBasisPer, lot.Provenance, lot.Status, lot.BuyTradeID, lot.CreatedAt, lot.UpdatedAt)
	return err
}

func (r *InvestmentRepo) UpdateShareLotQuantity(ctx context.Context, userID string, lotID string, quantity string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	status := "active"
	if quantity == "0" || quantity == "0.00" || quantity == "0.00000000" {
		status = "closed"
	}

	_, err = pool.Exec(ctx, `
		UPDATE share_lots 
		SET quantity = $1::numeric, status = $2, updated_at = NOW()
		WHERE id = $3
	`, quantity, status, lotID)
	return err
}

func (r *InvestmentRepo) DeleteShareLotsByTradeID(ctx context.Context, userID string, tradeID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM share_lots WHERE buy_trade_id = $1`, tradeID)
	return err
}

func (r *InvestmentRepo) CreateRealizedTradeLog(ctx context.Context, userID string, log entity.RealizedTradeLog) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO realized_trade_logs (
			id, broker_account_id, security_id, sell_trade_id, source_share_lot_id, 
			quantity, acquisition_date, cost_basis_total, sell_price, proceeds, 
			realized_pnl, provenance, created_at
		) VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12,$13)
	`, log.ID, log.BrokerAccountID, log.SecurityID, log.SellTradeID, log.SourceShareLot, log.Quantity, log.AcquisitionDate, log.CostBasisTotal, log.SellPrice, log.Proceeds, log.RealizedPnL, log.Provenance, log.CreatedAt)
	return err
}

func (r *InvestmentRepo) ListRealizedLogsByTradeID(ctx context.Context, userID string, tradeID string) ([]entity.RealizedTradeLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			id, broker_account_id, security_id, sell_trade_id, source_share_lot_id, 
			quantity::text, acquisition_date, cost_basis_total::text, sell_price::text, 
			proceeds::text, realized_pnl::text, provenance, created_at
		FROM realized_trade_logs
		WHERE sell_trade_id = $1
	`, tradeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.RealizedTradeLog
	for rows.Next() {
		var l entity.RealizedTradeLog
		var acqDate time.Time
		if err := rows.Scan(
			&l.ID, &l.BrokerAccountID, &l.SecurityID, &l.SellTradeID, &l.SourceShareLot,
			&l.Quantity, &acqDate, &l.CostBasisTotal, &l.SellPrice, &l.Proceeds,
			&l.RealizedPnL, &l.Provenance, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.AcquisitionDate = acqDate.Format("2006-01-02")
		results = append(results, l)
	}
	return results, nil
}

func (r *InvestmentRepo) DeleteRealizedLogsByTradeID(ctx context.Context, userID string, tradeID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM realized_trade_logs WHERE sell_trade_id = $1`, tradeID)
	return err
}

func (r *InvestmentRepo) ListRealizedLogs(ctx context.Context, userID string, brokerAccountID string) ([]entity.RealizedTradeLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			l.id, l.broker_account_id, l.security_id, l.sell_trade_id, l.source_share_lot_id, 
			l.quantity::text, l.acquisition_date, l.cost_basis_total::text, l.sell_price::text, 
			l.proceeds::text, l.realized_pnl::text, l.provenance, l.created_at
		FROM realized_trade_logs l
		JOIN investment_accounts ia ON ia.id = l.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.broker_account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY l.created_at DESC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.RealizedTradeLog
	for rows.Next() {
		var l entity.RealizedTradeLog
		var acqDate time.Time
		if err := rows.Scan(
			&l.ID, &l.BrokerAccountID, &l.SecurityID, &l.SellTradeID, &l.SourceShareLot,
			&l.Quantity, &acqDate, &l.CostBasisTotal, &l.SellPrice, &l.Proceeds,
			&l.RealizedPnL, &l.Provenance, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.AcquisitionDate = acqDate.Format("2006-01-02")
		results = append(results, l)
	}
	return results, nil
}
