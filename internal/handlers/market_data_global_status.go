package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
)

type GlobalMarketDataStatusResponse struct {
	MarketSync *MarketDataSyncState `json:"market_sync"`
	RateLimit  *MarketDataRateLimit `json:"rate_limit,omitempty"`
}

func loadSyncStateByKey(ctx context.Context, d Deps, syncKey string) (*MarketDataSyncState, error) {
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

func fetchRateLimitFromStatusURL(ctx context.Context, d Deps) (*MarketDataRateLimit, error) {
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

// GetGlobalMarketDataStatus godoc
// @Summary Global market-data worker status
// @Description Returns global worker rate-limit remaining and market sync last timestamps.
// @Tags investments
// @Produce json
// @Success 200 {object} GlobalMarketDataStatusResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /market-data/vnstock/status [get]
func GetGlobalMarketDataStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		marketSync, err := loadSyncStateByKey(ctx, d, "vnstock.market_sync")
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		rateLimit, _ := fetchRateLimitFromStatusURL(ctx, d) // best-effort

		writeJSON(w, http.StatusOK, GlobalMarketDataStatusResponse{MarketSync: marketSync, RateLimit: rateLimit})
	}
}
