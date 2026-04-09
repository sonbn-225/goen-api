package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type MarketDataService struct {
	cfg       *config.Config
	repo      interfaces.MarketDataRepository
	redis  *database.Redis
	secSvc interfaces.SecurityService
}

func NewMarketDataService(
	cfg *config.Config,
	repo interfaces.MarketDataRepository,
	redis *database.Redis,
	secSvc interfaces.SecurityService,
) *MarketDataService {
	return &MarketDataService{
		cfg:    cfg,
		repo:   repo,
		redis:  redis,
		secSvc: secSvc,
	}
}

func (s *MarketDataService) EnqueueSecurityPricesDaily(ctx context.Context, userID uuid.UUID, req dto.RefreshPriceRequest) (dto.RefreshOneResponse, error) {
	if s.redis == nil {
		return dto.RefreshOneResponse{}, errors.New("redis is not configured")
	}

	if _, err := s.secSvc.GetSecurity(ctx, req.SecurityID); err != nil {
		return dto.RefreshOneResponse{}, err
	}

	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.prices_daily",
		"security_id":          req.SecurityID.String(),
		"requested_by_user_id": userID.String(),
	}
	if req.Force != nil {
		values["force"] = *req.Force
	}
	if req.Full != nil {
		values["full"] = *req.Full
	}
	if req.From != nil {
		values["from"] = *req.From
	}
	if req.To != nil {
		values["to"] = *req.To
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return dto.RefreshOneResponse{}, err
	}
	return dto.RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *MarketDataService) EnqueueSecurityEvents(ctx context.Context, userID uuid.UUID, req dto.RefreshEventRequest) (dto.RefreshOneResponse, error) {
	if s.redis == nil {
		return dto.RefreshOneResponse{}, errors.New("redis is not configured")
	}

	if _, err := s.secSvc.GetSecurity(ctx, req.SecurityID); err != nil {
		return dto.RefreshOneResponse{}, err
	}

	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.security_events",
		"security_id":          req.SecurityID.String(),
		"requested_by_user_id": userID.String(),
	}
	if req.Force != nil {
		values["force"] = *req.Force
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return dto.RefreshOneResponse{}, err
	}
	return dto.RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *MarketDataService) EnqueueMarketSync(ctx context.Context, userID uuid.UUID, req dto.MarketSyncRequest) (dto.RefreshOneResponse, error) {
	if s.redis == nil {
		return dto.RefreshOneResponse{}, errors.New("redis is not configured")
	}

	stream := "goen:market_data:jobs"
	values := map[string]any{
		"job_type":             "vnstock.market_sync",
		"include_prices":       boolTo01(req.IncludePrices),
		"include_events":       boolTo01(req.IncludeEvents),
		"requested_by_user_id": userID.String(),
	}
	if req.Force != nil {
		values["force"] = *req.Force
	}
	if req.Full && req.IncludePrices {
		values["full"] = "1"
	}

	id, err := s.redis.XAdd(ctx, stream, values)
	if err != nil {
		return dto.RefreshOneResponse{}, err
	}
	return dto.RefreshOneResponse{Stream: stream, MessageID: id}, nil
}

func (s *MarketDataService) EnqueueBySymbols(ctx context.Context, userID uuid.UUID, req dto.RefreshSymbolsRequest) (dto.RefreshManyResponse, error) {
	if s.redis == nil {
		return dto.RefreshManyResponse{}, errors.New("redis is not configured")
	}

	cleaned := []string{}
	for _, sym := range req.Symbols {
		s0 := strings.ToUpper(strings.TrimSpace(sym))
		if s0 != "" {
			cleaned = append(cleaned, s0)
		}
	}

	idsBySymbol, err := s.repo.LoadSecurityIDsBySymbolsTx(ctx, nil, cleaned)
	if err != nil {
		return dto.RefreshManyResponse{}, err
	}

	stream := "goen:market_data:jobs"
	messageIDs := []string{}
	enqueued := 0
	notFound := []string{}

	for _, sym := range cleaned {
		securityID, ok := idsBySymbol[sym]
		if !ok || securityID == uuid.Nil {
			notFound = append(notFound, sym)
			continue
		}

		if req.IncludePrices {
			v := map[string]any{
				"job_type": "vnstock.prices_daily", "security_id": securityID.String(), "requested_by_user_id": userID.String(),
			}
			if req.Force != nil {
				v["force"] = *req.Force
			}
			id, _ := s.redis.XAdd(ctx, stream, v)
			messageIDs = append(messageIDs, id)
			enqueued++
		}
		if req.IncludeEvents {
			v := map[string]any{
				"job_type": "vnstock.security_events", "security_id": securityID.String(), "requested_by_user_id": userID.String(),
			}
			if req.Force != nil {
				v["force"] = *req.Force
			}
			id, _ := s.redis.XAdd(ctx, stream, v)
			messageIDs = append(messageIDs, id)
			enqueued++
		}
	}

	return dto.RefreshManyResponse{Stream: stream, Enqueued: enqueued, MessageIDs: messageIDs, NotFound: notFound}, nil
}

func (s *MarketDataService) GetSecurityStatus(ctx context.Context, userID, securityID uuid.UUID) (dto.SecurityStatus, error) {
	if _, err := s.secSvc.GetSecurity(ctx, securityID); err != nil {
		return dto.SecurityStatus{}, err
	}

	prices, _ := s.repo.LoadSyncStateTx(ctx, nil, "vnstock.prices_daily:"+securityID.String())
	events, _ := s.repo.LoadSyncStateTx(ctx, nil, "vnstock.security_events:"+securityID.String())
	rateLimit, _ := s.fetchRateLimit(ctx)

	return dto.SecurityStatus{SecurityID: securityID, Prices: prices, Events: events, RateLimit: rateLimit}, nil
}

func (s *MarketDataService) GetGlobalStatus(ctx context.Context) (dto.GlobalStatus, error) {
	marketSync, _ := s.repo.LoadSyncStateTx(ctx, nil, "vnstock.market_sync")
	rateLimit, _ := s.fetchRateLimit(ctx)
	return dto.GlobalStatus{MarketSync: marketSync, RateLimit: rateLimit}, nil
}

func (s *MarketDataService) fetchRateLimit(ctx context.Context) (*entity.RateLimit, error) {
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
		Status    string            `json:"status"`
		RateLimit *entity.RateLimit `json:"rate_limit"`
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
