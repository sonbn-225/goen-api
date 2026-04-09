package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type SecurityRepo struct {
	BaseRepo
}

func NewSecurityRepo(db *database.Postgres) *SecurityRepo {
	return &SecurityRepo{BaseRepo: *NewBaseRepo(db)}
}

func (r *SecurityRepo) GetSecurityTx(ctx context.Context, tx pgx.Tx, securityID uuid.UUID) (*entity.Security, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, `
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

func (r *SecurityRepo) ListSecuritiesTx(ctx context.Context, tx pgx.Tx) ([]entity.Security, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
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

func (r *SecurityRepo) ListSecurityPricesTx(ctx context.Context, tx pgx.Tx, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityPriceDaily, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
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

func (r *SecurityRepo) ListSecurityEventsTx(ctx context.Context, tx pgx.Tx, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityEvent, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
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
		if ratioDenominator := ratioDen.String; ratioDen.Valid {
			e.RatioDenominator = &ratioDenominator
		}
		if subscriptionPrice := subPrice.String; subPrice.Valid {
			e.SubscriptionPrice = &subscriptionPrice
		}
		results = append(results, e)
	}
	return results, nil
}

func (r *SecurityRepo) GetSecurityEventTx(ctx context.Context, tx pgx.Tx, securityEventID uuid.UUID) (*entity.SecurityEvent, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, `
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
	if ratioDenominator := ratioDen.String; ratioDen.Valid {
		e.RatioDenominator = &ratioDenominator
	}
	if subscriptionPrice := subPrice.String; subPrice.Valid {
		e.SubscriptionPrice = &subscriptionPrice
	}
	return &e, nil
}
