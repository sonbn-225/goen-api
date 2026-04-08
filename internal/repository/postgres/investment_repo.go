package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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

func (r *InvestmentRepo) UpsertSecurityEventElection(ctx context.Context, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	row := pool.QueryRow(ctx, `
		WITH ok AS (
			SELECT 1
			FROM accounts a
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
		INSERT INTO security_event_elections (
			id, user_id, account_id, security_event_id, security_id,
			entitlement_date, holding_quantity_at_entitlement_date, entitled_quantity, elected_quantity,
			status, confirmed_at, note, created_at, updated_at
		)
		SELECT $3,$4,$1,$5,$6,$7,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13,$14,$15
		WHERE EXISTS (SELECT 1 FROM ok)
		ON CONFLICT (account_id, security_event_id)
		DO UPDATE SET
			elected_quantity = EXCLUDED.elected_quantity,
			status = EXCLUDED.status,
			confirmed_at = EXCLUDED.confirmed_at,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, account_id, security_event_id, security_id,
		          entitlement_date, holding_quantity_at_entitlement_date::text, entitled_quantity::text, elected_quantity::text,
		          status, confirmed_at, note, created_at, updated_at
	`, e.AccountID, userID, e.ID, e.UserID, e.SecurityEventID, e.SecurityID, e.EntitlementDate, e.HoldingQuantityAtEntitlement, e.EntitledQuantity, e.ElectedQuantity, e.Status, e.ConfirmedAt, e.Note, e.CreatedAt, e.UpdatedAt)

	var out entity.SecurityEventElection
	if err := row.Scan(
		&out.ID, &out.UserID, &out.AccountID, &out.SecurityEventID, &out.SecurityID,
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

func (r *InvestmentRepo) ListSecurityEventElections(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT e.id, e.user_id, e.account_id, e.security_event_id, e.security_id,
		       e.entitlement_date, e.holding_quantity_at_entitlement_date::text, e.entitled_quantity::text, e.elected_quantity::text,
		       e.status, e.confirmed_at, e.note, e.created_at, e.updated_at
		FROM security_event_elections e
		JOIN accounts a ON a.id = e.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE e.account_id = $1
		  AND ua.user_id = $2
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		  AND ($3::text IS NULL OR e.status = $3)
		ORDER BY e.updated_at DESC
	`, accountID, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.SecurityEventElection
	for rows.Next() {
		var e entity.SecurityEventElection
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.AccountID, &e.SecurityEventID, &e.SecurityID,
			&e.EntitlementDate, &e.HoldingQuantityAtEntitlement, &e.EntitledQuantity, &e.ElectedQuantity,
			&e.Status, &e.ConfirmedAt, &e.Note, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	return results, nil
}


func (r *InvestmentRepo) CreateTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Trade) error {
	cmd, err := tx.Exec(ctx, `
		INSERT INTO trades (
			id, account_id, security_id, fee_transaction_id, tax_transaction_id,
			side, quantity, price, fees, taxes, occurred_at, note, created_at, updated_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13,$14
		WHERE EXISTS (
			SELECT 1
			FROM accounts a
			JOIN user_accounts ua ON ua.account_id = a.id
			WHERE a.id = $2 AND ua.user_id = $15 AND ua.status = 'active' AND a.deleted_at IS NULL
			  AND ua.permission IN ('owner','editor')
		)
	`, t.ID, t.AccountID, t.SecurityID, t.FeeTransactionID, t.TaxTransactionID, t.Side, t.Quantity, t.Price, t.Fees, t.Taxes, t.OccurredAt, t.Note, t.CreatedAt, t.UpdatedAt, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("forbidden: account access required")
	}
	return nil
}

func (r *InvestmentRepo) DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) error {
	return DeleteTransactionTx(ctx, tx, userID, transactionID)
}

func (r *InvestmentRepo) GetTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT t.id, t.account_id, t.security_id,
		       t.fee_transaction_id, t.tax_transaction_id,
		       t.side, t.quantity::text, t.price::text, t.fees::text, t.taxes::text,
		       t.occurred_at, t.note, t.created_at, t.updated_at
		FROM trades t
		JOIN accounts a ON a.id = t.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, tradeID, userID)

	var t entity.Trade
	if err := row.Scan(
		&t.ID, &t.AccountID, &t.SecurityID, &t.FeeTransactionID, &t.TaxTransactionID,
		&t.Side, &t.Quantity, &t.Price, &t.Fees, &t.Taxes, &t.OccurredAt, &t.Note, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trade not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *InvestmentRepo) ListTrades(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Trade, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT t.id, t.account_id, t.security_id, t.fee_transaction_id, t.tax_transaction_id,
		       t.side, t.quantity::text, t.price::text, t.fees::text, t.taxes::text, t.occurred_at,
		       t.note, t.created_at, t.updated_at
		FROM trades t
		JOIN accounts a ON a.id = t.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY t.occurred_at DESC, t.created_at DESC
	`, accountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Trade
	for rows.Next() {
		var t entity.Trade
		if err := rows.Scan(
			&t.ID, &t.AccountID, &t.SecurityID, &t.FeeTransactionID, &t.TaxTransactionID,
			&t.Side, &t.Quantity, &t.Price, &t.Fees, &t.Taxes, &t.OccurredAt, &t.Note, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (r *InvestmentRepo) DeleteTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		DELETE FROM trades
		WHERE id = $1 AND EXISTS (
			SELECT 1
			FROM trades t
			JOIN accounts a ON a.id = t.account_id
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

func (r *InvestmentRepo) ListHoldings(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			h.id, h.account_id, h.security_id, h.quantity::text, 
			h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, 
			h.market_value::text, h.unrealized_pnl::text, h.as_of, h.created_at, h.updated_at
		FROM holdings h
		JOIN accounts a ON a.id = h.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY h.security_id ASC
	`, accountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Holding
	for rows.Next() {
		var h entity.Holding
		if err := rows.Scan(
			&h.ID, &h.AccountID, &h.SecurityID, &h.Quantity, &h.CostBasisTotal, &h.AvgCost,
			&h.MarketPrice, &h.MarketValue, &h.UnrealizedPnL, &h.AsOf, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, nil
}

func (r *InvestmentRepo) GetHolding(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT 
			h.id, h.account_id, h.security_id, h.quantity::text, 
			h.cost_basis_total::text, h.avg_cost::text, h.market_price::text, 
			h.market_value::text, h.unrealized_pnl::text, h.as_of, h.created_at, h.updated_at
		FROM holdings h
		JOIN accounts a ON a.id = h.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE h.account_id = $1 AND h.security_id = $2 AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, accountID, securityID, userID)

	var h entity.Holding
	if err := row.Scan(
		&h.ID, &h.AccountID, &h.SecurityID, &h.Quantity, &h.CostBasisTotal, &h.AvgCost,
		&h.MarketPrice, &h.MarketValue, &h.UnrealizedPnL, &h.AsOf, &h.CreatedAt, &h.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

func (r *InvestmentRepo) UpsertHolding(ctx context.Context, userID uuid.UUID, h entity.Holding) (*entity.Holding, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	row := pool.QueryRow(ctx, `
		INSERT INTO holdings (
			id, account_id, security_id, quantity, cost_basis_total, avg_cost, 
			market_price, market_value, unrealized_pnl, as_of, created_at, updated_at
		) VALUES ($1,$2,$3,$4::numeric,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10,$11,$12)
		ON CONFLICT (account_id, security_id) 
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			cost_basis_total = EXCLUDED.cost_basis_total,
			avg_cost = EXCLUDED.avg_cost,
			market_price = EXCLUDED.market_price,
			market_value = EXCLUDED.market_value,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			as_of = EXCLUDED.as_of,
			updated_at = EXCLUDED.updated_at
		RETURNING id, account_id, security_id, quantity::text, cost_basis_total::text, avg_cost::text, market_price::text, market_value::text, unrealized_pnl::text, as_of, created_at, updated_at
	`, h.ID, h.AccountID, h.SecurityID, h.Quantity, h.CostBasisTotal, h.AvgCost, h.MarketPrice, h.MarketValue, h.UnrealizedPnL, h.AsOf, h.CreatedAt, h.UpdatedAt)

	var out entity.Holding
	if err := row.Scan(
		&out.ID, &out.AccountID, &out.SecurityID, &out.Quantity, &out.CostBasisTotal, &out.AvgCost,
		&out.MarketPrice, &out.MarketValue, &out.UnrealizedPnL, &out.AsOf, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvestmentRepo) ListShareLots(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			l.id, l.account_id, l.security_id, l.quantity::text, 
			l.acquisition_date, l.cost_basis_per_share::text, l.provenance, 
			l.status, l.buy_trade_id, l.created_at, l.updated_at
		FROM share_lots l
		JOIN accounts a ON a.id = l.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.account_id = $1 AND l.security_id = $2 AND ua.user_id = $3 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY l.acquisition_date ASC, l.created_at ASC
	`, accountID, securityID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.ShareLot
	for rows.Next() {
		var l entity.ShareLot
		var acqDate time.Time
		if err := rows.Scan(
			&l.ID, &l.AccountID, &l.SecurityID, &l.Quantity, &acqDate,
			&l.CostBasisPer, &l.Provenance, &l.Status, &l.BuyTradeID, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.AcquisitionDate = acqDate.Format("2006-01-02")
		results = append(results, l)
	}
	return results, nil
}


func (r *InvestmentRepo) CreateShareLotTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lot entity.ShareLot) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO share_lots (
			id, account_id, security_id, quantity, acquisition_date, 
			cost_basis_per_share, provenance, status, buy_trade_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4::numeric,$5,$6::numeric,$7,$8,$9,$10,$11)
	`, lot.ID, lot.AccountID, lot.SecurityID, lot.Quantity, lot.AcquisitionDate, lot.CostBasisPer, lot.Provenance, lot.Status, lot.BuyTradeID, lot.CreatedAt, lot.UpdatedAt)
	return err
}


func (r *InvestmentRepo) UpdateShareLotQuantityTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lotID uuid.UUID, quantity string) error {
	status := "active"
	if quantity == "0" || quantity == "0.00" || quantity == "0.00000000" {
		status = "closed"
	}

	_, err := tx.Exec(ctx, `
		UPDATE share_lots 
		SET quantity = $1::numeric, status = $2, updated_at = NOW()
		WHERE id = $3
	`, quantity, status, lotID)
	return err
}

func (r *InvestmentRepo) DeleteShareLotsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM share_lots WHERE buy_trade_id = $1`, tradeID)
	return err
}


func (r *InvestmentRepo) CreateRealizedTradeLogTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, log entity.RealizedTradeLog) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO realized_trade_logs (
			id, account_id, security_id, sell_trade_id, source_share_lot_id, 
			quantity, acquisition_date, cost_basis_total, sell_price, proceeds, 
			realized_pnl, provenance, created_at
		) VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12,$13)
	`, log.ID, log.AccountID, log.SecurityID, log.SellTradeID, log.SourceShareLot, log.Quantity, log.AcquisitionDate, log.CostBasisTotal, log.SellPrice, log.Proceeds, log.RealizedPnL, log.Provenance, log.CreatedAt)
	return err
}

func (r *InvestmentRepo) ListRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			id, account_id, security_id, sell_trade_id, source_share_lot_id, 
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
			&l.ID, &l.AccountID, &l.SecurityID, &l.SellTradeID, &l.SourceShareLot,
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

func (r *InvestmentRepo) DeleteRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM realized_trade_logs WHERE sell_trade_id = $1`, tradeID)
	return err
}

func (r *InvestmentRepo) ListRealizedLogs(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.RealizedTradeLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			l.id, l.account_id, l.security_id, l.sell_trade_id, l.source_share_lot_id, 
			l.quantity::text, l.acquisition_date, l.cost_basis_total::text, l.sell_price::text, 
			l.proceeds::text, l.realized_pnl::text, l.provenance, l.created_at
		FROM realized_trade_logs l
		JOIN accounts a ON a.id = l.account_id
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE l.account_id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY l.created_at DESC
	`, accountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.RealizedTradeLog
	for rows.Next() {
		var l entity.RealizedTradeLog
		var acqDate time.Time
		if err := rows.Scan(
			&l.ID, &l.AccountID, &l.SecurityID, &l.SellTradeID, &l.SourceShareLot,
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
