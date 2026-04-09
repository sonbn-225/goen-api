package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// MarketDataRepository định nghĩa lớp truy cập dữ liệu cho giá chứng khoán bên ngoài và các sự kiện doanh nghiệp.
type MarketDataRepository interface {
	// --- Nhóm 1: Truy vấn Dữ liệu Thị trường (Flexible Tx) ---

	// LoadSecurityIDsBySymbolsTx phân giải các UUID nội bộ từ các mã niêm yết (ticker symbols).
	LoadSecurityIDsBySymbolsTx(ctx context.Context, tx pgx.Tx, symbols []string) (map[string]uuid.UUID, error)
	// LoadSyncStateTx lấy thời điểm cuối cùng dữ liệu thị trường được đồng bộ hóa cho một khóa cụ thể.
	LoadSyncStateTx(ctx context.Context, tx pgx.Tx, syncKey string) (*entity.SyncState, error)
}

// MarketDataService định nghĩa nghiệp vụ cho việc đồng bộ hóa dữ liệu thị trường chạy ngầm.
type MarketDataService interface {
	// EnqueueSecurityPricesDaily lập lịch một tác vụ chạy ngầm để lấy dữ liệu giá OHLC hàng ngày lịch sử.
	EnqueueSecurityPricesDaily(ctx context.Context, userID uuid.UUID, req dto.RefreshPriceRequest) (dto.RefreshOneResponse, error)
	// EnqueueSecurityEvents lập lịch một tác vụ để lấy thông tin chia tách và cổ tức.
	EnqueueSecurityEvents(ctx context.Context, userID uuid.UUID, req dto.RefreshEventRequest) (dto.RefreshOneResponse, error)
	// EnqueueMarketSync điều phối một đợt cập nhật toàn diện các mã niêm yết và giá cả.
	EnqueueMarketSync(ctx context.Context, userID uuid.UUID, req dto.MarketSyncRequest) (dto.RefreshOneResponse, error)
	// EnqueueBySymbols kích hoạt cập nhật cho một nhóm các mã niêm yết cụ thể.
	EnqueueBySymbols(ctx context.Context, userID uuid.UUID, req dto.RefreshSymbolsRequest) (dto.RefreshManyResponse, error)
	// GetSecurityStatus trả về trạng thái đồng bộ hóa cho một mã chứng khoán cụ thể.
	GetSecurityStatus(ctx context.Context, userID, securityID uuid.UUID) (dto.SecurityStatus, error)
	// GetGlobalStatus trả về sức khỏe và trạng thái của các trình đồng bộ hóa chạy ngầm.
	GetGlobalStatus(ctx context.Context) (dto.GlobalStatus, error)
}
