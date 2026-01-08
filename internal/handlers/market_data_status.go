package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type MarketDataRateLimit struct {
	PerMinute         int     `json:"per_minute"`
	PerHour           int     `json:"per_hour"`
	UsedMinute        int     `json:"used_minute"`
	UsedHour          int     `json:"used_hour"`
	RemainingMinute   int     `json:"remaining_minute"`
	RemainingHour     int     `json:"remaining_hour"`
	MinuteResetInSecs float64 `json:"minute_reset_in_seconds"`
	HourResetInSecs   float64 `json:"hour_reset_in_seconds"`
}

type MarketDataSyncState struct {
	SyncKey            string     `json:"sync_key"`
	MinIntervalSeconds int        `json:"min_interval_seconds"`
	LastStartedAt      *time.Time `json:"last_started_at"`
	LastSuccessAt      *time.Time `json:"last_success_at"`
	LastFailureAt      *time.Time `json:"last_failure_at"`
	LastStatus         string     `json:"last_status"`
	LastError          *string    `json:"last_error"`
	NextDueAt          *time.Time `json:"next_due_at"`
	CooldownSeconds    int        `json:"cooldown_seconds"`
}

type SecurityMarketDataStatusResponse struct {
	SecurityID string               `json:"security_id"`
	Prices     *MarketDataSyncState `json:"prices_daily"`
	Events     *MarketDataSyncState `json:"security_events"`
	RateLimit  *MarketDataRateLimit `json:"rate_limit,omitempty"`
}

func loadSyncState(ctx context.Context, d Deps, syncKey string) (*MarketDataSyncState, error) {
	if d.DB == nil {
		return nil, nil
	}
	pool, err := d.DB.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var (
		minIntervalSeconds int
		lastStartedAt      *time.Time
		lastSuccessAt      *time.Time
		lastFailureAt      *time.Time
		lastStatus         string
		lastError          *string
	)

	err = pool.QueryRow(ctx, `
		SELECT min_interval_seconds, last_started_at, last_success_at, last_failure_at, last_status, last_error
		FROM market_data_sync_states
		WHERE sync_key = $1
	`, syncKey).Scan(&minIntervalSeconds, &lastStartedAt, &lastSuccessAt, &lastFailureAt, &lastStatus, &lastError)
	if err != nil {
		// If missing row, just return nil (never run).
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	st := &MarketDataSyncState{
		SyncKey:            syncKey,
		MinIntervalSeconds: minIntervalSeconds,
		LastStartedAt:      lastStartedAt,
		LastSuccessAt:      lastSuccessAt,
		LastFailureAt:      lastFailureAt,
		LastStatus:         lastStatus,
		LastError:          lastError,
	}

	if lastSuccessAt != nil {
		next := lastSuccessAt.Add(time.Duration(minIntervalSeconds) * time.Second)
		st.NextDueAt = &next
		cd := int(time.Until(next).Seconds())
		if cd < 0 {
			cd = 0
		}
		st.CooldownSeconds = cd
	} else {
		st.NextDueAt = nil
		st.CooldownSeconds = 0
	}

	return st, nil
}

func fetchRateLimit(ctx context.Context, d Deps) (*MarketDataRateLimit, error) {
	if d.Cfg == nil || d.Cfg.MarketDataStatusURL == "" {
		return nil, nil
	}
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.Cfg.MarketDataStatusURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed struct {
		Status    string               `json:"status"`
		RateLimit *MarketDataRateLimit `json:"rate_limit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.RateLimit, nil
}

// GetSecurityMarketDataStatus godoc
// @Summary Market data status for a security
// @Description Returns last sync timestamps, cooldown to next allowed sync, and current worker rate-limit remaining (best-effort).
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} SecurityMarketDataStatusResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/market-data/status [get]
func GetSecurityMarketDataStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		// Validate security exists (permissions)
		if _, err := d.InvestmentService.GetSecurity(r.Context(), uid, securityID); err != nil {
			if err == domain.ErrSecurityNotFound {
				apierror.Write(w, http.StatusNotFound, "not_found", "security not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		prices, err := loadSyncState(ctx, d, "vnstock.prices_daily:"+securityID)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		events, err := loadSyncState(ctx, d, "vnstock.security_events:"+securityID)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		rateLimit, _ := fetchRateLimit(ctx, d) // best-effort

		writeJSON(w, http.StatusOK, SecurityMarketDataStatusResponse{
			SecurityID: securityID,
			Prices:     prices,
			Events:     events,
			RateLimit:  rateLimit,
		})
	}
}
