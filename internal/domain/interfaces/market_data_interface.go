package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type MarketDataRepository interface {
	LoadSecurityIDsBySymbols(ctx context.Context, symbols []string) (map[string]string, error)
	LoadSyncState(ctx context.Context, syncKey string) (*entity.SyncState, error)
}

type MarketDataService interface {
	EnqueueSecurityPricesDaily(ctx context.Context, userID string, req dto.RefreshPriceRequest) (dto.RefreshOneResponse, error)
	EnqueueSecurityEvents(ctx context.Context, userID string, req dto.RefreshEventRequest) (dto.RefreshOneResponse, error)
	EnqueueMarketSync(ctx context.Context, userID string, req dto.MarketSyncRequest) (dto.RefreshOneResponse, error)
	EnqueueBySymbols(ctx context.Context, userID string, req dto.RefreshSymbolsRequest) (dto.RefreshManyResponse, error)
	GetSecurityStatus(ctx context.Context, userID, securityID string) (dto.SecurityStatus, error)
	GetGlobalStatus(ctx context.Context) (dto.GlobalStatus, error)
}
