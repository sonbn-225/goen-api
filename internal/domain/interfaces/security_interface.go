package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// SecurityRepository định nghĩa lớp truy cập dữ liệu cho thông tin gốc về đầu tư (mã niêm yết, giá cả, sự kiện).
type SecurityRepository interface {
	// --- Nhóm 1: Truy vấn Metadata Toàn cầu (Read-only Optimized) ---

	// GetSecurity lấy thông tin metadata của một mã chứng khoán (tên, mã niêm yết, loại).
	GetSecurity(ctx context.Context, securityID uuid.UUID) (*entity.Security, error)
	// ListSecurities trả về tất cả các mã chứng khoán được hỗ trợ bởi nền tảng.
	ListSecurities(ctx context.Context) ([]entity.Security, error)
	// ListSecurityPrices trả về dữ liệu giá hàng ngày lịch sử.
	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityPriceDaily, error)
	// ListSecurityEvents trả về lịch sử các sự kiện doanh nghiệp (chia tách, cổ tức).
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityEvent, error)
	// GetSecurityEvent lấy chi tiết cho một sự kiện doanh nghiệp cụ thể.
	GetSecurityEvent(ctx context.Context, securityEventID uuid.UUID) (*entity.SecurityEvent, error)
}

// SecurityService định nghĩa lớp nghiệp vụ để truy cập dữ liệu gốc đầu tư toàn cầu.
type SecurityService interface {
	// GetSecurity trả về thông tin metadata chứng khoán đã được định dạng.
	GetSecurity(ctx context.Context, securityID uuid.UUID) (*dto.SecurityResponse, error)
	// ListSecurities trả về danh sách tất cả các tài sản có thể theo dõi.
	ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error)

	// ListSecurityPrices trả về giá lịch sử đã được định dạng để vẽ biểu đồ.
	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityPriceDailyResponse, error)
	// ListSecurityEvents trả về các sự kiện doanh nghiệp lịch sử đã được định dạng.
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityEventResponse, error)
}
