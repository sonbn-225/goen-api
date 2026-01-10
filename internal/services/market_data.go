package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

type RefreshOneResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

type RefreshManyResponse struct {
	Stream     string   `json:"stream"`
	Enqueued   int      `json:"enqueued"`
	MessageIDs []string `json:"message_ids"`
	NotFound   []string `json:"not_found_symbols,omitempty"`
}

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

type SecurityMarketDataStatus struct {
	SecurityID string               `json:"security_id"`
	Prices     *MarketDataSyncState `json:"prices_daily"`
	Events     *MarketDataSyncState `json:"security_events"`
	RateLimit  *MarketDataRateLimit `json:"rate_limit,omitempty"`
}

type GlobalMarketDataStatus struct {
	MarketSync *MarketDataSyncState `json:"market_sync"`
	RateLimit  *MarketDataRateLimit `json:"rate_limit,omitempty"`
}

type MarketDataService interface {
	EnqueueSecurityPricesDaily(ctx context.Context, userID, securityID string, force, full, from, to string) (RefreshOneResponse, error)
	EnqueueSecurityEvents(ctx context.Context, userID, securityID string, force string) (RefreshOneResponse, error)
	EnqueueMarketSync(ctx context.Context, userID string, includePrices, includeEvents bool, force string, full bool) (RefreshOneResponse, error)
	EnqueueBySymbol(ctx context.Context, userID, symbol string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error)
	EnqueueBySymbols(ctx context.Context, userID string, symbols []string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error)
	GetSecurityStatus(ctx context.Context, userID, securityID string) (SecurityMarketDataStatus, error)
	GetGlobalStatus(ctx context.Context, userID string) (GlobalMarketDataStatus, error)
}

type marketDataService struct {
	cfg       *config.Config
	db        *storage.Postgres
	redis     *storage.Redis
	investSvc InvestmentService
}

func NewMarketDataService(cfg *config.Config, db *storage.Postgres, redis *storage.Redis, investSvc InvestmentService) MarketDataService {
	return &marketDataService{cfg: cfg, db: db, redis: redis, investSvc: investSvc}
}

func mapInvestmentError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrSecurityNotFound) {
		return NotFoundError("security not found", nil)
	}
	if errors.Is(err, domain.ErrInvestmentForbidden) {
		return ForbiddenError("forbidden")
	}
	return err
}

func (s *marketDataService) EnqueueSecurityPricesDaily(ctx context.Context, userID, securityID string, force, full, from, to string) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, DependencyUnavailableError("redis is not configured")
	}

	// Permissions / existence check
	if _, err := s.investSvc.GetSecurity(ctx, userID, securityID); err != nil {
		return RefreshOneResponse{}, mapInvestmentError(err)
	}

	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.prices_daily",
		"security_id":          securityID,
		"requested_by_user_id": userID,
	}
	if force != "" {
		values["force"] = force
	}
	if full != "" {
		values["full"] = full
	}
	if from != "" {
		values["from"] = from
	}
	if to != "" {
		values["to"] = to
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return RefreshOneResponse{}, err
	}
	return RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *marketDataService) EnqueueSecurityEvents(ctx context.Context, userID, securityID string, force string) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, DependencyUnavailableError("redis is not configured")
	}

	if _, err := s.investSvc.GetSecurity(ctx, userID, securityID); err != nil {
		return RefreshOneResponse{}, mapInvestmentError(err)
	}

	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.security_events",
		"security_id":          securityID,
		"requested_by_user_id": userID,
	}
	if force != "" {
		values["force"] = force
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return RefreshOneResponse{}, err
	}
	return RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *marketDataService) EnqueueMarketSync(ctx context.Context, userID string, includePrices, includeEvents bool, force string, full bool) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, DependencyUnavailableError("redis is not configured")
	}
	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.market_sync",
		"include_prices":       boolTo01(includePrices),
		"include_events":       boolTo01(includeEvents),
		"requested_by_user_id": userID,
	}
	if force != "" {
		values["force"] = force
	}
	if full && includePrices {
		values["full"] = "1"
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return RefreshOneResponse{}, err
	}
	return RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *marketDataService) EnqueueBySymbol(ctx context.Context, userID, symbol string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
	if s.redis == nil {
		return RefreshManyResponse{}, DependencyUnavailableError("redis is not configured")
	}
	if s.db == nil {
		return RefreshManyResponse{}, DependencyUnavailableError("postgres is not configured")
	}

	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return RefreshManyResponse{}, ValidationError("symbol is required", map[string]any{"field": "symbol"})
	}

	idsBySymbol, err := s.loadSecurityIDsBySymbols(ctx, []string{sym})
	if err != nil {
		return RefreshManyResponse{}, err
	}
	securityID, found := idsBySymbol[sym]
	if !found || securityID == "" {
		return RefreshManyResponse{}, NotFoundError("security symbol not found", map[string]any{"symbol": sym})
	}

	return s.enqueueJobsForSecurityID(ctx, userID, securityID, includePrices, includeEvents, force)
}

func (s *marketDataService) EnqueueBySymbols(ctx context.Context, userID string, symbols []string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
	if s.redis == nil {
		return RefreshManyResponse{}, DependencyUnavailableError("redis is not configured")
	}
	if s.db == nil {
		return RefreshManyResponse{}, DependencyUnavailableError("postgres is not configured")
	}
	if len(symbols) == 0 {
		return RefreshManyResponse{}, ValidationError("symbols is required", map[string]any{"field": "symbols"})
	}

	cleaned := make([]string, 0, len(symbols))
	seen := map[string]struct{}{}
	for _, s0 := range symbols {
		sym := strings.ToUpper(strings.TrimSpace(s0))
		if sym == "" {
			continue
		}
		if _, ok := seen[sym]; ok {
			continue
		}
		seen[sym] = struct{}{}
		cleaned = append(cleaned, sym)
	}
	if len(cleaned) == 0 {
		return RefreshManyResponse{}, ValidationError("symbols is required", map[string]any{"field": "symbols"})
	}

	idsBySymbol, err := s.loadSecurityIDsBySymbols(ctx, cleaned)
	if err != nil {
		return RefreshManyResponse{}, err
	}

	stream := "goen:market_data:jobs"
	messageIDs := []string{}
	enqueued := 0
	notFound := []string{}

	for _, sym := range cleaned {
		securityID, ok := idsBySymbol[sym]
		if !ok || securityID == "" {
			notFound = append(notFound, sym)
			continue
		}

		res, err := s.enqueueJobsForSecurityID(ctx, userID, securityID, includePrices, includeEvents, force)
		if err != nil {
			return RefreshManyResponse{}, err
		}
		messageIDs = append(messageIDs, res.MessageIDs...)
		enqueued += res.Enqueued
	}

	return RefreshManyResponse{Stream: stream, Enqueued: enqueued, MessageIDs: messageIDs, NotFound: notFound}, nil
}

func (s *marketDataService) GetSecurityStatus(ctx context.Context, userID, securityID string) (SecurityMarketDataStatus, error) {
	if _, err := s.investSvc.GetSecurity(ctx, userID, securityID); err != nil {
		return SecurityMarketDataStatus{}, mapInvestmentError(err)
	}

	// If DB isn't configured, we still return a valid response, just without sync states.
	prices, err := s.loadSyncState(ctx, "vnstock.prices_daily:"+securityID)
	if err != nil {
		return SecurityMarketDataStatus{}, err
	}
	events, err := s.loadSyncState(ctx, "vnstock.security_events:"+securityID)
	if err != nil {
		return SecurityMarketDataStatus{}, err
	}

	rateLimit, _ := s.fetchRateLimit(ctx) // best-effort

	return SecurityMarketDataStatus{SecurityID: securityID, Prices: prices, Events: events, RateLimit: rateLimit}, nil
}

func (s *marketDataService) GetGlobalStatus(ctx context.Context, _ string) (GlobalMarketDataStatus, error) {
	marketSync, err := s.loadSyncStateByKey(ctx, "vnstock.market_sync")
	if err != nil {
		return GlobalMarketDataStatus{}, err
	}
	rateLimit, _ := s.fetchRateLimit(ctx) // best-effort
	return GlobalMarketDataStatus{MarketSync: marketSync, RateLimit: rateLimit}, nil
}

func (s *marketDataService) enqueueJobsForSecurityID(ctx context.Context, userID, securityID string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
	stream := "goen:market_data:jobs"
	messageIDs := []string{}
	enqueued := 0

	if includePrices {
		values := map[string]any{
			"job_type":             "vnstock.prices_daily",
			"security_id":          securityID,
			"requested_by_user_id": userID,
		}
		if force != "" {
			values["force"] = force
		}
		id, err := s.redis.XAdd(ctx, stream, values)
		if err != nil {
			return RefreshManyResponse{}, err
		}
		messageIDs = append(messageIDs, id)
		enqueued++
	}

	if includeEvents {
		values := map[string]any{
			"job_type":             "vnstock.security_events",
			"security_id":          securityID,
			"requested_by_user_id": userID,
		}
		if force != "" {
			values["force"] = force
		}
		id, err := s.redis.XAdd(ctx, stream, values)
		if err != nil {
			return RefreshManyResponse{}, err
		}
		messageIDs = append(messageIDs, id)
		enqueued++
	}

	return RefreshManyResponse{Stream: stream, Enqueued: enqueued, MessageIDs: messageIDs}, nil
}

func (s *marketDataService) loadSecurityIDsBySymbols(ctx context.Context, symbols []string) (map[string]string, error) {
	if s.db == nil {
		return nil, nil
	}
	pool, err := s.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, nil
	}

	rows, err := pool.Query(ctx, `
		SELECT symbol, id
		FROM securities
		WHERE symbol = ANY($1)
	`, symbols)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var sym, id string
		if err := rows.Scan(&sym, &id); err != nil {
			return nil, err
		}
		out[strings.ToUpper(strings.TrimSpace(sym))] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *marketDataService) loadSyncState(ctx context.Context, syncKey string) (*MarketDataSyncState, error) {
	if s.db == nil {
		return nil, nil
	}
	pool, err := s.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, nil
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
	}

	return st, nil
}

func (s *marketDataService) loadSyncStateByKey(ctx context.Context, syncKey string) (*MarketDataSyncState, error) {
	return s.loadSyncState(ctx, syncKey)
}

func (s *marketDataService) fetchRateLimit(ctx context.Context) (*MarketDataRateLimit, error) {
	if s.cfg == nil || s.cfg.MarketDataStatusURL == "" {
		return nil, nil
	}
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.MarketDataStatusURL, nil)
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

func boolTo01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
