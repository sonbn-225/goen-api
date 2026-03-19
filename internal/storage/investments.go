package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
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
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Only allow creating extension for an accessible broker account.
	// Require write permission (owner/editor) because this is a write action.
	cmd, err := pool.Exec(ctx, `
		INSERT INTO investment_accounts (
			id, account_id, fee_settings, tax_settings, created_at, updated_at
		)
		SELECT $1,$2,$3,$4,$5,$6
		WHERE EXISTS (
			SELECT 1
			FROM accounts a
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE a.id = $2
			  AND a.account_type = 'broker'
			  AND a.deleted_at IS NULL
			  AND ua.user_id = $7
			  AND ua.status = 'active'
			  AND ua.permission IN ('owner','editor')
		)
	`, ia.ID, ia.AccountID, ia.FeeSettings, ia.TaxSettings, ia.CreatedAt, ia.UpdatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*domain.InvestmentAccount, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
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

	var out domain.InvestmentAccount
	var feeSettings any
	var taxSettings any
	if err := row.Scan(&out.ID, &out.AccountID, &out.Currency, &feeSettings, &taxSettings, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrInvestmentAccountNotFound
		}
		return nil, err
	}
	if feeSettings != nil {
		out.FeeSettings = feeSettings
	}
	if taxSettings != nil {
		out.TaxSettings = taxSettings
	}
	return &out, nil
}

func (r *InvestmentRepo) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
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

	out := []domain.InvestmentAccount{}
	for rows.Next() {
		var ia domain.InvestmentAccount
		var feeSettings any
		var taxSettings any
		if err := rows.Scan(&ia.ID, &ia.AccountID, &ia.Currency, &feeSettings, &taxSettings, &ia.CreatedAt, &ia.UpdatedAt); err != nil {
			return nil, err
		}
		if feeSettings != nil {
			ia.FeeSettings = feeSettings
		}
		if taxSettings != nil {
			ia.TaxSettings = taxSettings
		}
		out = append(out, ia)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) UpdateInvestmentAccountSettings(ctx context.Context, userID string, investmentAccountID string, feeSettings any, taxSettings any) (*domain.InvestmentAccount, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Write requires owner/editor on underlying broker cash ledger.
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

	var out domain.InvestmentAccount
	var fee any
	var tax any
	if err := row.Scan(&out.ID, &out.AccountID, &fee, &tax, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrInvestmentForbidden
		}
		return nil, err
	}
	if fee != nil {
		out.FeeSettings = fee
	}
	if tax != nil {
		out.TaxSettings = tax
	}

	// Fill currency via existing accessor (keeps output consistent with GetInvestmentAccount).
	acc, err := r.GetInvestmentAccount(ctx, userID, out.ID)
	if err == nil && acc != nil {
		acc.FeeSettings = out.FeeSettings
		acc.TaxSettings = out.TaxSettings
		acc.CreatedAt = out.CreatedAt
		acc.UpdatedAt = out.UpdatedAt
		return acc, nil
	}
	return &out, nil
}

func (r *InvestmentRepo) GetSecurity(ctx context.Context, securityID string) (*domain.Security, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
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
			return nil, apperrors.ErrSecurityNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *InvestmentRepo) ListSecurities(ctx context.Context) ([]domain.Security, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
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
		return nil, apperrors.ErrDatabaseNotReady
	}
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
		return nil, apperrors.ErrDatabaseNotReady
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
		return nil, apperrors.ErrDatabaseNotReady
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
			return nil, apperrors.ErrSecurityEventNotFound
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
		return nil, apperrors.ErrDatabaseNotReady
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
			return nil, apperrors.ErrInvestmentForbidden
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
		return nil, apperrors.ErrDatabaseNotReady
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
		  AND ($3::text IS NULL OR e.status = $3::security_event_election_status)
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
		return apperrors.ErrDatabaseNotReady
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
		return apperrors.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]domain.Trade, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
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

func (r *InvestmentRepo) GetTrade(ctx context.Context, userID string, tradeID string) (*domain.Trade, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT t.id, t.client_id, t.broker_account_id, t.security_id,
		       t.fee_transaction_id, t.tax_transaction_id,
		       t.side::text, t.quantity::text, t.price::text, t.fees::text, t.taxes::text,
		       t.occurred_at, t.note, t.created_at, t.updated_at
		FROM trades t
		JOIN investment_accounts ia ON ia.id = t.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, tradeID, userID)

	var t domain.Trade
	if err := row.Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTradeNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *InvestmentRepo) DeleteTrade(ctx context.Context, userID string, tradeID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
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
		return apperrors.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) DeleteShareLotsByTradeID(ctx context.Context, userID string, tradeID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM share_lots
		WHERE buy_trade_id = $1 AND EXISTS (
			SELECT 1
			FROM share_lots l
			JOIN investment_accounts ia ON ia.id = l.broker_account_id
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE l.buy_trade_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, tradeID, userID)
	return err
}

func (r *InvestmentRepo) DeleteRealizedLogsByTradeID(ctx context.Context, userID string, tradeID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM realized_trade_logs
		WHERE sell_trade_id = $1 AND EXISTS (
			SELECT 1
			FROM realized_trade_logs l
			JOIN investment_accounts ia ON ia.id = l.broker_account_id
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE l.sell_trade_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, tradeID, userID)
	return err
}

func (r *InvestmentRepo) ListRealizedLogsByTradeID(ctx context.Context, userID, tradeID string) ([]domain.RealizedTradeLog, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.broker_account_id, l.security_id, l.sell_trade_id, l.source_share_lot_id,
		       l.quantity::text, l.acquisition_date::text, l.cost_basis_total::text,
		       l.sell_price::text, l.proceeds::text, l.realized_pnl::text, l.provenance::text, l.created_at
		FROM realized_trade_logs l
		JOIN investment_accounts ia ON ia.id = l.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.sell_trade_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, tradeID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.RealizedTradeLog{}
	for rows.Next() {
		var l domain.RealizedTradeLog
		if err := rows.Scan(
			&l.ID,
			&l.BrokerAccountID,
			&l.SecurityID,
			&l.SellTradeID,
			&l.SourceShareLot,
			&l.Quantity,
			&l.AcquisitionDate,
			&l.CostBasisTotal,
			&l.SellPrice,
			&l.Proceeds,
			&l.RealizedPnL,
			&l.Provenance,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]domain.Holding, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT h.id, h.broker_account_id, h.security_id, h.quantity::text,
		       h.cost_basis_total::text, h.avg_cost::text,
		       mv.market_price::text, mv.market_value::text, mv.unrealized_pnl::text,
		       h.as_of, h.source_of_truth, h.created_at, h.updated_at
		FROM holdings h
		LEFT JOIN LATERAL (
			SELECT spd.close * 1000 AS latest_close
			FROM security_price_dailies spd
			WHERE spd.security_id = h.security_id
			ORDER BY spd.price_date DESC
			LIMIT 1
		) lp ON TRUE
		LEFT JOIN LATERAL (
			SELECT
				COALESCE(lp.latest_close, h.market_price) AS market_price,
				CASE
					WHEN COALESCE(lp.latest_close, h.market_price) IS NULL THEN NULL
					ELSE ROUND(h.quantity * COALESCE(lp.latest_close, h.market_price), 2)
				END AS market_value,
				CASE
					WHEN h.cost_basis_total IS NULL OR COALESCE(lp.latest_close, h.market_price) IS NULL THEN NULL
					ELSE ROUND(h.quantity * COALESCE(lp.latest_close, h.market_price), 2) - h.cost_basis_total
				END AS unrealized_pnl
		) mv ON TRUE
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
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT h.id, h.broker_account_id, h.security_id, h.quantity::text,
		       h.cost_basis_total::text, h.avg_cost::text,
		       mv.market_price::text, mv.market_value::text, mv.unrealized_pnl::text,
		       h.as_of, h.source_of_truth, h.created_at, h.updated_at
		FROM holdings h
		LEFT JOIN LATERAL (
			SELECT spd.close * 1000 AS latest_close
			FROM security_price_dailies spd
			WHERE spd.security_id = h.security_id
			ORDER BY spd.price_date DESC
			LIMIT 1
		) lp ON TRUE
		LEFT JOIN LATERAL (
			SELECT
				COALESCE(lp.latest_close, h.market_price) AS market_price,
				CASE
					WHEN COALESCE(lp.latest_close, h.market_price) IS NULL THEN NULL
					ELSE ROUND(h.quantity * COALESCE(lp.latest_close, h.market_price), 2)
				END AS market_value,
				CASE
					WHEN h.cost_basis_total IS NULL OR COALESCE(lp.latest_close, h.market_price) IS NULL THEN NULL
					ELSE ROUND(h.quantity * COALESCE(lp.latest_close, h.market_price), 2) - h.cost_basis_total
				END AS unrealized_pnl
		) mv ON TRUE
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
			return nil, apperrors.ErrHoldingNotFound
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

func (r *InvestmentRepo) UpsertHolding(ctx context.Context, userID string, h domain.Holding) (*domain.Holding, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
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
		INSERT INTO holdings (
			id, broker_account_id, security_id, quantity, cost_basis_total, avg_cost, as_of, source_of_truth, created_at, updated_at
		)
		SELECT $3,$1,$4,$5::numeric,$6::numeric,$7::numeric,$8,$9,$10,$11
		WHERE EXISTS (SELECT 1 FROM ok)
		ON CONFLICT (broker_account_id, security_id)
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			cost_basis_total = EXCLUDED.cost_basis_total,
			avg_cost = EXCLUDED.avg_cost,
			source_of_truth = EXCLUDED.source_of_truth,
			updated_at = EXCLUDED.updated_at
		WHERE EXISTS (SELECT 1 FROM ok)
		RETURNING id, broker_account_id, security_id, quantity::text,
		          cost_basis_total::text, avg_cost::text, market_price::text, market_value::text, unrealized_pnl::text,
		          as_of, source_of_truth, created_at, updated_at
	`, h.BrokerAccountID, userID, h.ID, h.SecurityID, h.Quantity, h.CostBasisTotal, h.AvgCost, h.AsOf, h.SourceOfTruth, h.CreatedAt, h.UpdatedAt)

	var out domain.Holding
	var costBasisTotal, avgCost, marketPrice, marketValue, unrealizedPnL sql.NullString
	if err := row.Scan(
		&out.ID,
		&out.BrokerAccountID,
		&out.SecurityID,
		&out.Quantity,
		&costBasisTotal,
		&avgCost,
		&marketPrice,
		&marketValue,
		&unrealizedPnL,
		&out.AsOf,
		&out.SourceOfTruth,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrInvestmentForbidden
		}
		return nil, err
	}

	out.CostBasisTotal = nullStringPtr(costBasisTotal)
	out.AvgCost = nullStringPtr(avgCost)
	out.MarketPrice = nullStringPtr(marketPrice)
	out.MarketValue = nullStringPtr(marketValue)
	out.UnrealizedPnL = nullStringPtr(unrealizedPnL)
	return &out, nil
}

func (r *InvestmentRepo) ListShareLots(ctx context.Context, userID string, brokerAccountID string, securityID string) ([]domain.ShareLot, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.broker_account_id, l.security_id, l.quantity::text,
		       l.acquisition_date::text, l.cost_basis_per_share::text, l.provenance::text, l.status::text,
		       l.buy_trade_id, l.created_at, l.updated_at
		FROM share_lots l
		JOIN investment_accounts ia ON ia.id = l.broker_account_id
		JOIN accounts a ON a.id = ia.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.broker_account_id = $1 AND l.security_id = $2
		  AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY l.acquisition_date ASC, l.created_at ASC
	`, brokerAccountID, securityID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ShareLot{}
	for rows.Next() {
		var lot domain.ShareLot
		if err := rows.Scan(
			&lot.ID,
			&lot.BrokerAccountID,
			&lot.SecurityID,
			&lot.Quantity,
			&lot.AcquisitionDate,
			&lot.CostBasisPer,
			&lot.Provenance,
			&lot.Status,
			&lot.BuyTradeID,
			&lot.CreatedAt,
			&lot.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, lot)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) CreateShareLot(ctx context.Context, userID string, lot domain.ShareLot) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		INSERT INTO share_lots (
			id, broker_account_id, security_id, quantity, acquisition_date, cost_basis_per_share,
			provenance, status, buy_trade_id, created_at, updated_at
		)
		SELECT $1,$2,$3,$4::numeric,$5::date,$6::numeric,$7::lot_provenance,$8::lot_status,$9,$10,$11
		WHERE EXISTS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $2 AND ua.user_id = $12 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, lot.ID, lot.BrokerAccountID, lot.SecurityID, lot.Quantity, lot.AcquisitionDate, lot.CostBasisPer, lot.Provenance, lot.Status, lot.BuyTradeID, lot.CreatedAt, lot.UpdatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) UpdateShareLotQuantity(ctx context.Context, userID string, lotID string, quantity string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		WITH ok AS (
			SELECT 1
			FROM share_lots l
			JOIN investment_accounts ia ON ia.id = l.broker_account_id
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE l.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
		UPDATE share_lots
		SET quantity = $3::numeric,
		    updated_at = NOW()
		WHERE id = $1 AND EXISTS (SELECT 1 FROM ok)
	`, lotID, userID, quantity)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrInvestmentForbidden
	}
	return nil
}

func (r *InvestmentRepo) CreateRealizedTradeLog(ctx context.Context, userID string, log domain.RealizedTradeLog) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		INSERT INTO realized_trade_logs (
			id, broker_account_id, security_id, sell_trade_id, source_share_lot_id, quantity,
			acquisition_date, cost_basis_total, sell_price, proceeds, realized_pnl, provenance, created_at
		)
		SELECT $1,$2,$3,$4,$5,$6::numeric,$7::date,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12::lot_provenance,$13
		WHERE EXISTS (
			SELECT 1
			FROM investment_accounts ia
			JOIN accounts a ON a.id = ia.account_id
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE ia.id = $2 AND ua.user_id = $14 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, log.ID, log.BrokerAccountID, log.SecurityID, log.SellTradeID, log.SourceShareLot, log.Quantity, log.AcquisitionDate, log.CostBasisTotal, log.SellPrice, log.Proceeds, log.RealizedPnL, log.Provenance, log.CreatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrInvestmentForbidden
	}
	return nil
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
func (r *InvestmentRepo) ListRealizedLogs(ctx context.Context, userID string, brokerAccountID string) ([]domain.RealizedTradeLog, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.broker_account_id, l.security_id, l.sell_trade_id, l.source_share_lot_id,
		       l.quantity::text, l.acquisition_date::text, l.cost_basis_total::text,
		       l.sell_price::text, l.proceeds::text, l.realized_pnl::text, l.provenance::text, l.created_at
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

	out := []domain.RealizedTradeLog{}
	for rows.Next() {
		var l domain.RealizedTradeLog
		if err := rows.Scan(
			&l.ID,
			&l.BrokerAccountID,
			&l.SecurityID,
			&l.SellTradeID,
			&l.SourceShareLot,
			&l.Quantity,
			&l.AcquisitionDate,
			&l.CostBasisTotal,
			&l.SellPrice,
			&l.Proceeds,
			&l.RealizedPnL,
			&l.Provenance,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *InvestmentRepo) ListDividends(ctx context.Context, userID string, brokerAccountID string) ([]domain.Transaction, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// We look for transactions linked to this investment account's ledger
	// which are tagged as dividends (via ExternalRef pattern).
	rows, err := pool.Query(ctx, `
		SELECT t.id, t.client_id, t.external_ref, t.type, t.occurred_at,
		       to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
		       t.amount::text,
		       (
				   SELECT li.note
				   FROM transaction_line_items li
				   WHERE li.transaction_id = t.id
				   ORDER BY li.id
				   LIMIT 1
		       ) AS description,
		       t.account_id, t.status, t.created_at, t.updated_at
		FROM transactions t
		JOIN investment_accounts ia ON ia.account_id = t.account_id
		JOIN user_accounts ua ON ua.account_id = ia.account_id
		WHERE ia.id = $1 AND ua.user_id = $2 AND ua.status = 'active'
		  AND t.deleted_at IS NULL
		  AND t.type = 'income'
		  AND t.external_ref LIKE 'event:%'
		ORDER BY t.occurred_at DESC
	`, brokerAccountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Transaction{}
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID, &t.ClientID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate,
			&t.Amount, &t.Description, &t.AccountID, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
