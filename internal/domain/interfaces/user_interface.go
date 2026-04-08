package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// UserRepository định nghĩa lớp truy cập dữ liệu cho tài khoản người dùng và hồ sơ cá nhân.
type UserRepository interface {
	// CreateUserWithRefreshToken lưu một người dùng mới cùng với token làm mới ban đầu một cách nguyên tử.
	CreateUserWithRefreshToken(ctx context.Context, user entity.UserWithPassword, refreshToken entity.RefreshToken) error
	// FindUserByEmail tìm kiếm người dùng theo địa chỉ email đã đăng ký.
	FindUserByEmail(ctx context.Context, email string) (*entity.UserWithPassword, error)
	// FindUserByPhone tìm kiếm người dùng theo số điện thoại.
	FindUserByPhone(ctx context.Context, phone string) (*entity.UserWithPassword, error)
	// FindUserByUsername tìm kiếm người dùng theo định danh đăng nhập duy nhất.
	FindUserByUsername(ctx context.Context, username string) (*entity.UserWithPassword, error)
	// FindUserByID tìm kiếm hồ sơ người dùng theo UUID nội bộ.
	FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// UpdateUserSettings cập nhật các cài đặt mới vào JSON cài đặt của người dùng.
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*entity.User, error)
	// UpdateUserProfile chỉnh sửa thông tin metadata hồ sơ người dùng.
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, params entity.UpdateUserParams) (*entity.User, error)
}

// RefreshTokenRepository định nghĩa lớp truy cập dữ liệu để lưu trữ các token phiên làm việc.
type RefreshTokenRepository interface {
	// Create lưu trữ một token làm mới mới.
	Create(ctx context.Context, token *entity.RefreshToken) error
	// GetByToken lấy thông tin chi tiết của token làm mới để xác thực hoặc làm mới phiên.
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	// DeleteByToken xóa một token làm mới cụ thể (đăng xuất).
	DeleteByToken(ctx context.Context, token string) error
	// DeleteAllByUserID xóa tất cả các phiên làm việc của một người dùng (phản ứng khi có sự cố bảo mật).
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// AuthService định nghĩa lớp nghiệp vụ cho xác thực, quản lý phiên và cài đặt hồ sơ.
type AuthService interface {
	// Signup tạo tài khoản người dùng mới và trả về phiên làm việc ban đầu.
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	// Signin xác thực thông tin đăng nhập và cấp các token mới.
	Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error)
	// Refresh cấp một token truy cập mới bằng cách sử dụng token làm mới hợp lệ.
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	// Logout vô hiệu hóa một phiên làm việc của người dùng.
	Logout(ctx context.Context, refreshToken string) error
	// GetMe trả về hồ sơ của người dùng đang đăng nhập.
	GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	// UpdateMySettings cho phép người dùng hiện tại chỉnh sửa các tùy chọn ứng dụng của họ.
	UpdateMySettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*dto.UserResponse, error)
	// UploadAvatar xử lý việc lưu trữ và liên kết ảnh đại diện.
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UserResponse, error)
	// GetMyAvatars trả về danh sách các ảnh đại diện đã tải lên.
	GetMyAvatars(ctx context.Context, userID uuid.UUID) ([]dto.MediaResponse, error)
	// UpdateMyProfile chỉnh sửa thông tin metadata hồ sơ (email, tên hiển thị).
	UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, email, phone, username *string) (*dto.UserResponse, error)
	// ChangePassword xử lý việc cập nhật mật khẩu bảo mật.
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}
