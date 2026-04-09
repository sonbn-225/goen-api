package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// UserRepository định nghĩa lớp truy cập dữ liệu cho tài khoản người dùng và hồ sơ cá nhân.
type UserRepository interface {
	// --- Nhóm 1: Truy vấn Người dùng (Flexible Tx) ---

	// FindUserByEmailTx tìm kiếm người dùng theo địa chỉ email đã đăng ký.
	FindUserByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*entity.UserWithPassword, error)
	// FindUserByPhoneTx tìm kiếm người dùng theo số điện thoại.
	FindUserByPhoneTx(ctx context.Context, tx pgx.Tx, phone string) (*entity.UserWithPassword, error)
	// FindUserByUsernameTx tìm kiếm người dùng theo định danh đăng nhập duy nhất.
	FindUserByUsernameTx(ctx context.Context, tx pgx.Tx, username string) (*entity.UserWithPassword, error)
	// FindUserByIDTx tìm kiếm hồ sơ người dùng theo UUID nội bộ.
	FindUserByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.User, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateUserWithRefreshTokenTx lưu một người dùng mới cùng với token làm mới ban đầu một cách nguyên tử.
	CreateUserWithRefreshTokenTx(ctx context.Context, tx pgx.Tx, user entity.UserWithPassword, refreshToken entity.RefreshToken) error
	// UpdateUserSettingsTx cập nhật các cài đặt mới.
	UpdateUserSettingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, patch map[string]any) (*entity.User, error)
	// UpdateUserProfileTx chỉnh sửa thông tin metadata hồ sơ.
	UpdateUserProfileTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, params entity.UpdateUserParams) (*entity.User, error)
}

// RefreshTokenRepository định nghĩa lớp truy cập dữ liệu để lưu trữ các token phiên làm việc.
type RefreshTokenRepository interface {
	// --- Nhóm 1: Truy vấn Token (Flexible Tx) ---

	// GetByTokenTx lấy thông tin chi tiết của token làm mới (hỗ trợ transaction).
	GetByTokenTx(ctx context.Context, tx pgx.Tx, token string) (*entity.RefreshToken, error)

	// --- Nhóm 2: Thao tác Token (Transactional) ---

	// CreateTx lưu trữ một token làm mới mới.
	CreateTx(ctx context.Context, tx pgx.Tx, token *entity.RefreshToken) error
	// DeleteByTokenTx xóa một token làm mới cụ thể.
	DeleteByTokenTx(ctx context.Context, tx pgx.Tx, token string) error
}

// AuthService định nghĩa lớp nghiệp vụ cho xác thực, quản lý phiên và cài đặt hồ sơ.
type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	UpdateMySettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*dto.UserResponse, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UserResponse, error)
	GetMyAvatars(ctx context.Context, userID uuid.UUID) ([]dto.MediaResponse, error)
	UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, email, phone, username *string) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}
