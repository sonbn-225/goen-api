package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    interfaces.UserRepository
	refreshRepo interfaces.RefreshTokenRepository
	s3          *storage.S3Client
	cfg         *config.Config
}

func NewAuthService(
	userRepo interfaces.UserRepository,
	refreshRepo interfaces.RefreshTokenRepository,
	s3 *storage.S3Client,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		s3:          s3,
		cfg:         cfg,
	}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)
	username := strings.ToLower(strings.TrimSpace(req.Username))

	if email == "" && phone == "" {
		return nil, apperr.BadRequest("invalid_input", "email or phone is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := entity.UserWithPassword{
		User: entity.User{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
			},
			Email:       &email,
			Phone:       &phone,
			DisplayName: &req.DisplayName,
			Username:    username,
			Status:      "active",
			Settings:    s.defaultUserSettings("en"), // Hardcoded for now
		},
		PasswordHash: string(hash),
	}
	u.PublicShareURL = s.generatePublicShareURL(u.Username)

	accessToken, expiresIn, err := s.generateAccessToken(u.User)
	if err != nil {
		return nil, err
	}

	refreshTokenEntity, refreshToken := s.newRefreshToken(u.ID)
	if err := s.userRepo.CreateUserWithRefreshToken(ctx, u, *refreshTokenEntity); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         s.mapToUserResponse(u.User),
	}, nil
}

func (s *AuthService) Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error) {
	login := strings.ToLower(strings.TrimSpace(req.Login))

	var u *entity.UserWithPassword
	var err error

	if strings.Contains(login, "@") {
		u, err = s.userRepo.FindUserByEmail(ctx, login)
	} else if len(login) > 0 && login[0] != '+' && !isNumeric(login) {
		u, err = s.userRepo.FindUserByUsername(ctx, login)
	} else {
		u, err = s.userRepo.FindUserByPhone(ctx, login)
	}

	if err != nil {
		return nil, apperr.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperr.ErrInvalidCredentials
	}

	accessToken, expiresIn, err := s.generateAccessToken(u.User)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndSaveRefreshToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         s.mapToUserResponse(u.User),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// 1. Verify refresh token exists in DB
	rt, err := s.refreshRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, apperr.Unauthorized("invalid refresh token")
	}

	// 2. Check expiry
	if utils.Now().After(rt.ExpiresAt) {
		_ = s.refreshRepo.DeleteByToken(ctx, refreshToken)
		return nil, apperr.Unauthorized("refresh token expired")
	}

	// 3. Get user
	u, err := s.userRepo.FindUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// 4. Generate new pair (Rotation)
	_ = s.refreshRepo.DeleteByToken(ctx, refreshToken)

	newAccessToken, expiresIn, err := s.generateAccessToken(*u)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateAndSaveRefreshToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         s.mapToUserResponse(*u),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshRepo.DeleteByToken(ctx, refreshToken)
}

func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	u, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := s.mapToUserResponse(*u)
	return &res, nil
}

func (s *AuthService) UpdateMySettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*dto.UserResponse, error) {
	u, err := s.userRepo.UpdateUserSettings(ctx, userID, patch)
	if err != nil {
		return nil, err
	}
	res := s.mapToUserResponse(*u)
	return &res, nil
}

func (s *AuthService) UploadAvatar(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UserResponse, error) {
	if s.s3 == nil {
		return nil, apperr.Internal("storage not configured")
	}

	key, err := s.s3.UploadAvatar(ctx, userID.String(), file)
	if err != nil {
		return nil, err
	}

	url := s.s3.AvatarURL(s.cfg.PublicBaseURL, key)
	u, err := s.userRepo.UpdateUserProfile(ctx, userID, entity.UpdateUserParams{
		AvatarURL: &url,
	})
	if err != nil {
		return nil, err
	}
	res := s.mapToUserResponse(*u)
	return &res, nil
}

func (s *AuthService) GetMyAvatars(ctx context.Context, userID uuid.UUID) ([]dto.MediaResponse, error) {
	if s.s3 == nil {
		return nil, apperr.Internal("storage not configured")
	}

	prefix := fmt.Sprintf("avatars/%s/", userID.String())
	objects, err := s.s3.ListObjects(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var res []dto.MediaResponse
	for _, obj := range objects {
		res = append(res, dto.MediaResponse{
			Key:       obj.Key,
			URL:       s.s3.AvatarURL(s.cfg.PublicBaseURL, obj.Key),
			Size:      obj.Size,
			UpdatedAt: obj.LastModified,
		})
	}

	return res, nil
}

func (s *AuthService) UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, email, phone, username *string) (*dto.UserResponse, error) {
	params := entity.UpdateUserParams{}
	if displayName != nil {
		v := strings.TrimSpace(*displayName)
		params.DisplayName = &v
	}
	if email != nil {
		v := strings.ToLower(strings.TrimSpace(*email))
		params.Email = &v
	}
	if phone != nil {
		v := strings.TrimSpace(*phone)
		params.Phone = &v
	}
	if username != nil {
		v := strings.ToLower(strings.TrimSpace(*username))
		params.Username = &v
	}

	u, err := s.userRepo.UpdateUserProfile(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	res := s.mapToUserResponse(*u)
	return &res, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	u, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	var uWithPass *entity.UserWithPassword
	if u.Email != nil {
		uWithPass, err = s.userRepo.FindUserByEmail(ctx, *u.Email)
	} else if u.Phone != nil {
		uWithPass, err = s.userRepo.FindUserByPhone(ctx, *u.Phone)
	} else {
		uWithPass, err = s.userRepo.FindUserByUsername(ctx, u.Username)
	}

	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(uWithPass.PasswordHash), []byte(currentPassword)); err != nil {
		return apperr.BadRequest("invalid_password", "invalid current password")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashStr := string(newHash)
	_, err = s.userRepo.UpdateUserProfile(ctx, userID, entity.UpdateUserParams{
		PasswordHash: &hashStr,
	})
	return err
}

func (s *AuthService) generateAccessToken(user entity.User) (string, int, error) {
	// Access tokens are short-lived
	exp := utils.Now().Add(time.Duration(s.cfg.JWTAccessTTL) * time.Minute)
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"iat": utils.Now().Unix(),
		"exp": exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, s.cfg.JWTAccessTTL * 60, nil
}

func (s *AuthService) mapToUserResponse(user entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		Phone:          user.Phone,
		DisplayName:    user.DisplayName,
		AvatarURL:      user.AvatarURL,
		Username:       user.Username,
		PublicShareURL: user.PublicShareURL,
		Settings:       user.Settings,
		Status:         user.Status,
	}
}

func (s *AuthService) generateAndSaveRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rt, token := s.newRefreshToken(userID)

	if err := s.refreshRepo.Create(ctx, rt); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) newRefreshToken(userID uuid.UUID) (*entity.RefreshToken, string) {
	token := uuid.NewString()
	now := utils.Now()

	return &entity.RefreshToken{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{ID: utils.NewID()},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}, token
}

func (s *AuthService) generatePublicShareURL(username string) *string {
	if username == "" {
		return nil
	}
	url := fmt.Sprintf("%s/u/%s", s.cfg.PublicBaseURL, username)
	return &url
}

func (s *AuthService) defaultUserSettings(lang string) map[string]any {
	return map[string]any{
		"locale":                "en-US",
		"default_currency":      "VND",
		"timezone":              "Asia/Ho_Chi_Minh",
		"rotating_savings_term": "hui",
	}
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
