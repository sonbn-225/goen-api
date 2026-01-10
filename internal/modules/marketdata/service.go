package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// RefreshOneResponse contains enqueue result for single job.
type RefreshOneResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

// RefreshManyResponse contains enqueue result for multiple jobs.
type RefreshManyResponse struct {
	Stream     string   `json:"stream"`
	Enqueued   int      `json:"enqueued"`
	MessageIDs []string `json:"message_ids"`
	NotFound   []string `json:"not_found_symbols,omitempty"`
}

// RateLimit represents rate limit state.
type RateLimit struct {
	PerMinute         int     `json:"per_minute"`
	PerHour           int     `json:"per_hour"`
	UsedMinute        int     `json:"used_minute"`
	UsedHour          int     `json:"used_hour"`
	RemainingMinute   int     `json:"remaining_minute"`
	RemainingHour     int     `json:"remaining_hour"`
	MinuteResetInSecs float64 `json:"minute_reset_in_seconds"`
	HourResetInSecs   float64 `json:"hour_reset_in_seconds"`
}

// SyncState represents sync state for a resource.
type SyncState struct {
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

// SecurityStatus contains market data sync status for a security.
type SecurityStatus struct {
	SecurityID string     `json:"security_id"`
	Prices     *SyncState `json:"prices_daily"`
	Events     *SyncState `json:"security_events"`
	RateLimit  *RateLimit `json:"rate_limit,omitempty"`
}

// GlobalStatus contains global market data sync status.
type GlobalStatus struct {
	MarketSync *SyncState `json:"market_sync"`
	RateLimit  *RateLimit `json:"rate_limit,omitempty"`
}

// Service handles market data business logic.
type Service struct {
	cfg       *config.Config
	repo      Repository
	redis     *storage.Redis
	investSvc InvestmentServiceInterface
}

// NewService creates a new market data service.
func NewService(cfg *config.Config, repo Repository, redis *storage.Redis, investSvc InvestmentServiceInterface) *Service {
	return &Service{cfg: cfg, repo: repo, redis: redis, investSvc: investSvc}
}

func (s *Service) mapInvestmentError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, apperrors.ErrSecurityNotFound) {
		return apperrors.Wrap(apperrors.KindNotFound, "security not found", err)
	}
	if errors.Is(err, apperrors.ErrInvestmentForbidden) {
		return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
	}
	return err
}

// EnqueueSecurityPricesDaily enqueues a job to refresh daily prices.
func (s *Service) EnqueueSecurityPricesDaily(ctx context.Context, userID, securityID, force, full, from, to string) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, apperrors.DependencyUnavailable("redis is not configured")
	}

	if _, err := s.investSvc.GetSecurity(ctx, securityID); err != nil {
		return RefreshOneResponse{}, s.mapInvestmentError(err)
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

// EnqueueSecurityEvents enqueues a job to refresh security events.
func (s *Service) EnqueueSecurityEvents(ctx context.Context, userID, securityID, force string) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, apperrors.DependencyUnavailable("redis is not configured")
	}

	if _, err := s.investSvc.GetSecurity(ctx, securityID); err != nil {
		return RefreshOneResponse{}, s.mapInvestmentError(err)
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

// EnqueueMarketSync enqueues a market-wide sync job.
func (s *Service) EnqueueMarketSync(ctx context.Context, userID string, includePrices, includeEvents bool, force string, full bool) (RefreshOneResponse, error) {
	if s.redis == nil {
		return RefreshOneResponse{}, apperrors.DependencyUnavailable("redis is not configured")
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

// EnqueueBySymbol enqueues jobs for a single symbol.
func (s *Service) EnqueueBySymbol(ctx context.Context, userID, symbol string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
	if s.redis == nil {
		return RefreshManyResponse{}, apperrors.DependencyUnavailable("redis is not configured")
	}
	if s.repo == nil {
		return RefreshManyResponse{}, apperrors.DependencyUnavailable("market data repository is not configured")
	}

	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return RefreshManyResponse{}, apperrors.Validation("symbol is required", map[string]any{"field": "symbol"})
	}

	idsBySymbol, err := s.repo.LoadSecurityIDsBySymbols(ctx, []string{sym})
	if err != nil {
		return RefreshManyResponse{}, err
	}
	securityID, found := idsBySymbol[sym]
	if !found || securityID == "" {
		return RefreshManyResponse{}, apperrors.NotFound("security symbol not found", map[string]any{"symbol": sym})
	}

	return s.enqueueJobsForSecurityID(ctx, userID, securityID, includePrices, includeEvents, force)
}

// EnqueueBySymbols enqueues jobs for multiple symbols.
func (s *Service) EnqueueBySymbols(ctx context.Context, userID string, symbols []string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
	if s.redis == nil {
		return RefreshManyResponse{}, apperrors.DependencyUnavailable("redis is not configured")
	}
	if s.repo == nil {
		return RefreshManyResponse{}, apperrors.DependencyUnavailable("market data repository is not configured")
	}
	if len(symbols) == 0 {
		return RefreshManyResponse{}, apperrors.Validation("symbols is required", map[string]any{"field": "symbols"})
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
		return RefreshManyResponse{}, apperrors.Validation("symbols is required", map[string]any{"field": "symbols"})
	}

	idsBySymbol, err := s.repo.LoadSecurityIDsBySymbols(ctx, cleaned)
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

// GetSecurityStatus returns market data sync status for a security.
func (s *Service) GetSecurityStatus(ctx context.Context, userID, securityID string) (SecurityStatus, error) {
	if _, err := s.investSvc.GetSecurity(ctx, securityID); err != nil {
		return SecurityStatus{}, s.mapInvestmentError(err)
	}

	prices, err := s.repo.LoadSyncState(ctx, "vnstock.prices_daily:"+securityID)
	if err != nil {
		return SecurityStatus{}, err
	}
	events, err := s.repo.LoadSyncState(ctx, "vnstock.security_events:"+securityID)
	if err != nil {
		return SecurityStatus{}, err
	}

	rateLimit, _ := s.fetchRateLimit(ctx)

	return SecurityStatus{SecurityID: securityID, Prices: prices, Events: events, RateLimit: rateLimit}, nil
}

// GetGlobalStatus returns global market data sync status.
func (s *Service) GetGlobalStatus(ctx context.Context, _ string) (GlobalStatus, error) {
	marketSync, err := s.repo.LoadSyncState(ctx, "vnstock.market_sync")
	if err != nil {
		return GlobalStatus{}, err
	}
	rateLimit, _ := s.fetchRateLimit(ctx)
	return GlobalStatus{MarketSync: marketSync, RateLimit: rateLimit}, nil
}

func (s *Service) enqueueJobsForSecurityID(ctx context.Context, userID, securityID string, includePrices, includeEvents bool, force string) (RefreshManyResponse, error) {
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

func (s *Service) fetchRateLimit(ctx context.Context) (*RateLimit, error) {
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
		Status    string     `json:"status"`
		RateLimit *RateLimit `json:"rate_limit"`
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
