package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/investment"
)

type InvestmentRepository struct {
	db *pgxpool.Pool
}

func NewInvestmentRepository(db *pgxpool.Pool) *InvestmentRepository {
	return &InvestmentRepository{db: db}
}

func (r *InvestmentRepository) GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*investment.InvestmentAccount, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "investment", "operation", "get_investment_account", "user_id", userID, "investment_account_id", investmentAccountID)

	row := r.db.QueryRow(ctx, `
		SELECT ia.id, ia.account_id, a.currency, ia.fee_settings, ia.tax_settings, ia.created_at, ia.updated_at
		FROM investment_accounts ia
		JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ia.id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
	`, investmentAccountID, userID)

	item, err := scanInvestmentAccount(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_investment_get_account_failed", "error", err)
		return nil, err
	}
	return item, nil
}

func (r *InvestmentRepository) ListInvestmentAccounts(ctx context.Context, userID string) ([]investment.InvestmentAccount, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "investment", "operation", "list_investment_accounts", "user_id", userID)

	rows, err := r.db.Query(ctx, `
		SELECT ia.id, ia.account_id, a.currency, ia.fee_settings, ia.tax_settings, ia.created_at, ia.updated_at
		FROM investment_accounts ia
		JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		ORDER BY ia.created_at DESC
	`, userID)
	if err != nil {
		logger.Error("repo_investment_list_accounts_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]investment.InvestmentAccount, 0)
	for rows.Next() {
		item, err := scanInvestmentAccount(rows)
		if err != nil {
			logger.Error("repo_investment_list_accounts_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		logger.Error("repo_investment_list_accounts_failed", "error", err)
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, feeSettings any, taxSettings any) (*investment.InvestmentAccount, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "investment", "operation", "update_investment_account_settings", "user_id", userID, "investment_account_id", investmentAccountID)

	now := time.Now().UTC()
	row := r.db.QueryRow(ctx, `
		WITH ok AS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $1
			  AND ua.user_id = $2
			  AND ua.status = 'active'
			  AND ua.permission IN ('owner', 'editor')
		)
		UPDATE investment_accounts
		SET fee_settings = COALESCE($3, fee_settings),
		    tax_settings = COALESCE($4, tax_settings),
		    updated_at = $5
		WHERE id = $1
		  AND EXISTS (SELECT 1 FROM ok)
		RETURNING id, account_id, NULL::text AS currency, fee_settings, tax_settings, created_at, updated_at
	`, investmentAccountID, userID, feeSettings, taxSettings, now)

	item, err := scanInvestmentAccount(row)
	if err != nil {
		if isNoRows(err) {
			existing, getErr := r.GetInvestmentAccount(ctx, userID, investmentAccountID)
			if getErr != nil {
				logger.Error("repo_investment_update_account_failed", "error", getErr)
				return nil, getErr
			}
			if existing == nil {
				return nil, nil
			}
			return nil, apperrors.New(apperrors.KindForbidden, "no permission to update investment account settings")
		}
		logger.Error("repo_investment_update_account_failed", "error", err)
		return nil, err
	}

	fresh, err := r.GetInvestmentAccount(ctx, userID, item.ID)
	if err != nil {
		logger.Error("repo_investment_update_account_failed", "error", err)
		return nil, err
	}
	return fresh, nil
}

func (r *InvestmentRepository) GetSecurity(ctx context.Context, securityID string) (*investment.Security, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, symbol, name, asset_class::text, currency, created_at, updated_at
		FROM securities
		WHERE id = $1
	`, securityID)

	item, err := scanSecurity(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *InvestmentRepository) ListSecurities(ctx context.Context) ([]investment.Security, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, symbol, name, asset_class::text, currency, created_at, updated_at
		FROM securities
		ORDER BY symbol ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]investment.Security, 0)
	for rows.Next() {
		item, err := scanSecurity(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]investment.SecurityPriceDaily, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, security_id, price_date, open::text, high::text, low::text, close::text, volume::text, created_at, updated_at
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

	items := make([]investment.SecurityPriceDaily, 0)
	for rows.Next() {
		item, err := scanSecurityPrice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]investment.SecurityEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			security_id,
			event_type::text,
			to_char(ex_date, 'YYYY-MM-DD'),
			to_char(record_date, 'YYYY-MM-DD'),
			to_char(pay_date, 'YYYY-MM-DD'),
			to_char(effective_date, 'YYYY-MM-DD'),
			cash_amount_per_share::text,
			ratio_numerator::text,
			ratio_denominator::text,
			subscription_price::text,
			currency,
			vnstock_event_id,
			note,
			created_at,
			updated_at
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

	items := make([]investment.SecurityEvent, 0)
	for rows.Next() {
		item, err := scanSecurityEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) CreateTrade(ctx context.Context, userID string, trade investment.Trade) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "investment", "operation", "create_trade", "user_id", userID, "trade_id", trade.ID)

	commandTag, err := r.db.Exec(ctx, `
		INSERT INTO trades (
			id,
			client_id,
			broker_account_id,
			security_id,
			fee_transaction_id,
			tax_transaction_id,
			side,
			quantity,
			price,
			fees,
			taxes,
			occurred_at,
			note,
			created_at,
			updated_at
		)
		SELECT
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::trade_side,
			$8::numeric,
			$9::numeric,
			$10::numeric,
			$11::numeric,
			$12,
			$13,
			$14,
			$15
		WHERE EXISTS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $3
			  AND ua.user_id = $16
			  AND ua.status = 'active'
			  AND ua.permission IN ('owner', 'editor')
		)
	`,
		trade.ID,
		trade.ClientID,
		trade.BrokerAccountID,
		trade.SecurityID,
		trade.FeeTransactionID,
		trade.TaxTransactionID,
		trade.Side,
		trade.Quantity,
		trade.Price,
		trade.Fees,
		trade.Taxes,
		trade.OccurredAt,
		trade.Note,
		trade.CreatedAt,
		trade.UpdatedAt,
		userID,
	)
	if err != nil {
		logger.Error("repo_investment_create_trade_failed", "error", err)
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return apperrors.New(apperrors.KindForbidden, "no permission to create trade for this investment account")
	}
	return nil
}

func (r *InvestmentRepository) ListTrades(ctx context.Context, userID, brokerAccountID string) ([]investment.Trade, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			t.id,
			t.client_id,
			t.broker_account_id,
			t.security_id,
			t.fee_transaction_id,
			t.tax_transaction_id,
			t.side::text,
			t.quantity::text,
			t.price::text,
			t.fees::text,
			t.taxes::text,
			t.occurred_at,
			t.note,
			t.created_at,
			t.updated_at
		FROM trades t
		JOIN investment_accounts ia ON ia.id = t.broker_account_id
		JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.broker_account_id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		ORDER BY t.occurred_at DESC, t.id DESC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]investment.Trade, 0)
	for rows.Next() {
		item, err := scanTrade(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]investment.Holding, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			h.id,
			h.broker_account_id,
			h.security_id,
			h.quantity::text,
			h.cost_basis_total::text,
			h.avg_cost::text,
			h.market_price::text,
			h.market_value::text,
			h.unrealized_pnl::text,
			h.as_of,
			h.source_of_truth::text,
			h.created_at,
			h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		ORDER BY h.security_id ASC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]investment.Holding, 0)
	for rows.Next() {
		item, err := scanHolding(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InvestmentRepository) GetHolding(ctx context.Context, userID, brokerAccountID, securityID string) (*investment.Holding, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
			h.id,
			h.broker_account_id,
			h.security_id,
			h.quantity::text,
			h.cost_basis_total::text,
			h.avg_cost::text,
			h.market_price::text,
			h.market_value::text,
			h.unrealized_pnl::text,
			h.as_of,
			h.source_of_truth::text,
			h.created_at,
			h.updated_at
		FROM holdings h
		JOIN investment_accounts ia ON ia.id = h.broker_account_id
		JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.broker_account_id = $1
		  AND h.security_id = $2
		  AND ua.user_id = $3
		  AND ua.status = 'active'
	`, brokerAccountID, securityID, userID)

	item, err := scanHolding(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *InvestmentRepository) UpsertHolding(ctx context.Context, userID string, holding investment.Holding) (*investment.Holding, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO holdings (
			id,
			broker_account_id,
			security_id,
			quantity,
			cost_basis_total,
			avg_cost,
			market_price,
			market_value,
			unrealized_pnl,
			as_of,
			source_of_truth,
			created_at,
			updated_at
		)
		SELECT
			$1,
			$2,
			$3,
			$4::numeric,
			$5::numeric,
			$6::numeric,
			$7::numeric,
			$8::numeric,
			$9::numeric,
			$10,
			$11::holding_source_of_truth,
			$12,
			$13
		WHERE EXISTS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id AND a.deleted_at IS NULL
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $2
			  AND ua.user_id = $14
			  AND ua.status = 'active'
			  AND ua.permission IN ('owner', 'editor')
		)
		ON CONFLICT (broker_account_id, security_id)
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			cost_basis_total = EXCLUDED.cost_basis_total,
			avg_cost = EXCLUDED.avg_cost,
			market_price = EXCLUDED.market_price,
			market_value = EXCLUDED.market_value,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			as_of = EXCLUDED.as_of,
			source_of_truth = EXCLUDED.source_of_truth,
			updated_at = EXCLUDED.updated_at
		RETURNING
			id,
			broker_account_id,
			security_id,
			quantity::text,
			cost_basis_total::text,
			avg_cost::text,
			market_price::text,
			market_value::text,
			unrealized_pnl::text,
			as_of,
			source_of_truth::text,
			created_at,
			updated_at
	`,
		holding.ID,
		holding.BrokerAccountID,
		holding.SecurityID,
		holding.Quantity,
		holding.CostBasisTotal,
		holding.AvgCost,
		holding.MarketPrice,
		holding.MarketValue,
		holding.UnrealizedPnL,
		holding.AsOf,
		holding.SourceOfTruth,
		holding.CreatedAt,
		holding.UpdatedAt,
		userID,
	)

	item, err := scanHolding(row)
	if err != nil {
		if isNoRows(err) {
			return nil, apperrors.New(apperrors.KindForbidden, "no permission to update holding")
		}
		return nil, err
	}
	return item, nil
}

type investmentScanner interface {
	Scan(dest ...any) error
}

func scanInvestmentAccount(scanner investmentScanner) (*investment.InvestmentAccount, error) {
	var item investment.InvestmentAccount
	var currency sql.NullString
	var feeSettings any
	var taxSettings any
	if err := scanner.Scan(&item.ID, &item.AccountID, &currency, &feeSettings, &taxSettings, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if currency.Valid {
		item.Currency = currency.String
	}
	if feeSettings != nil {
		item.FeeSettings = feeSettings
	}
	if taxSettings != nil {
		item.TaxSettings = taxSettings
	}
	return &item, nil
}

func scanSecurity(scanner investmentScanner) (*investment.Security, error) {
	var item investment.Security
	var name sql.NullString
	var assetClass sql.NullString
	var currency sql.NullString
	if err := scanner.Scan(&item.ID, &item.Symbol, &name, &assetClass, &currency, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if name.Valid {
		item.Name = &name.String
	}
	if assetClass.Valid {
		item.AssetClass = &assetClass.String
	}
	if currency.Valid {
		item.Currency = &currency.String
	}
	return &item, nil
}

func scanSecurityPrice(scanner investmentScanner) (*investment.SecurityPriceDaily, error) {
	var item investment.SecurityPriceDaily
	var priceDate time.Time
	var open, high, low, close, volume sql.NullString
	if err := scanner.Scan(&item.ID, &item.SecurityID, &priceDate, &open, &high, &low, &close, &volume, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.PriceDate = priceDate.Format("2006-01-02")
	if open.Valid {
		item.Open = &open.String
	}
	if high.Valid {
		item.High = &high.String
	}
	if low.Valid {
		item.Low = &low.String
	}
	if close.Valid {
		item.Close = close.String
	}
	if volume.Valid {
		item.Volume = &volume.String
	}
	return &item, nil
}

func scanSecurityEvent(scanner investmentScanner) (*investment.SecurityEvent, error) {
	var item investment.SecurityEvent
	var exDate, recordDate, payDate, effectiveDate sql.NullString
	var cashAmountPerShare, ratioNumerator, ratioDenominator, subscriptionPrice sql.NullString
	var currency, vnstockEventID, note sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.SecurityID,
		&item.EventType,
		&exDate,
		&recordDate,
		&payDate,
		&effectiveDate,
		&cashAmountPerShare,
		&ratioNumerator,
		&ratioDenominator,
		&subscriptionPrice,
		&currency,
		&vnstockEventID,
		&note,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if exDate.Valid {
		item.ExDate = &exDate.String
	}
	if recordDate.Valid {
		item.RecordDate = &recordDate.String
	}
	if payDate.Valid {
		item.PayDate = &payDate.String
	}
	if effectiveDate.Valid {
		item.EffectiveDate = &effectiveDate.String
	}
	if cashAmountPerShare.Valid {
		item.CashAmountPerShare = &cashAmountPerShare.String
	}
	if ratioNumerator.Valid {
		item.RatioNumerator = &ratioNumerator.String
	}
	if ratioDenominator.Valid {
		item.RatioDenominator = &ratioDenominator.String
	}
	if subscriptionPrice.Valid {
		item.SubscriptionPrice = &subscriptionPrice.String
	}
	if currency.Valid {
		item.Currency = &currency.String
	}
	if vnstockEventID.Valid {
		item.VnstockEventID = &vnstockEventID.String
	}
	if note.Valid {
		item.Note = &note.String
	}

	return &item, nil
}

func scanTrade(scanner investmentScanner) (*investment.Trade, error) {
	var item investment.Trade
	var clientID, feeTransactionID, taxTransactionID, note sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&clientID,
		&item.BrokerAccountID,
		&item.SecurityID,
		&feeTransactionID,
		&taxTransactionID,
		&item.Side,
		&item.Quantity,
		&item.Price,
		&item.Fees,
		&item.Taxes,
		&item.OccurredAt,
		&note,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if clientID.Valid {
		item.ClientID = &clientID.String
	}
	if feeTransactionID.Valid {
		item.FeeTransactionID = &feeTransactionID.String
	}
	if taxTransactionID.Valid {
		item.TaxTransactionID = &taxTransactionID.String
	}
	if note.Valid {
		item.Note = &note.String
	}
	return &item, nil
}

func scanHolding(scanner investmentScanner) (*investment.Holding, error) {
	var item investment.Holding
	var costBasisTotal, avgCost, marketPrice, marketValue, unrealizedPnL sql.NullString
	var sourceOfTruth string

	if err := scanner.Scan(
		&item.ID,
		&item.BrokerAccountID,
		&item.SecurityID,
		&item.Quantity,
		&costBasisTotal,
		&avgCost,
		&marketPrice,
		&marketValue,
		&unrealizedPnL,
		&item.AsOf,
		&sourceOfTruth,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if costBasisTotal.Valid {
		item.CostBasisTotal = &costBasisTotal.String
	}
	if avgCost.Valid {
		item.AvgCost = &avgCost.String
	}
	if marketPrice.Valid {
		item.MarketPrice = &marketPrice.String
	}
	if marketValue.Valid {
		item.MarketValue = &marketValue.String
	}
	if unrealizedPnL.Valid {
		item.UnrealizedPnL = &unrealizedPnL.String
	}
	item.SourceOfTruth = strings.TrimSpace(sourceOfTruth)
	return &item, nil
}

var _ investment.Repository = (*InvestmentRepository)(nil)
