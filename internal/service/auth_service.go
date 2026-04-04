package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo interfaces.UserRepository
	s3       *storage.S3Client
	cfg      *config.Config
}

func NewAuthService(userRepo interfaces.UserRepository, s3 *storage.S3Client, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		s3:       s3,
		cfg:      cfg,
	}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)
	username := strings.ToLower(strings.TrimSpace(req.Username))

	if email == "" && phone == "" {
		return nil, errors.New("email or phone is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := entity.UserWithPassword{
		User: entity.User{
			ID:          uuid.NewString(),
			Email:       &email,
			Phone:       &phone,
			DisplayName: &req.DisplayName,
			Username:    username,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
			Settings:    s.defaultUserSettings("en"), // Hardcoded for now
		},
		PasswordHash: string(hash),
	}
	u.PublicShareURL = s.generatePublicShareURL(u.Username)

	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	token, exp, err := s.generateToken(u.User)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   exp,
		User:        u.User,
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
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, exp, err := s.generateToken(u.User)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   exp,
		User:        u.User,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, userID string) (*dto.AuthResponse, error) {
	u, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	token, exp, err := s.generateToken(*u)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   exp,
		User:        *u,
	}, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*entity.User, error) {
	return s.userRepo.FindUserByID(ctx, userID)
}

func (s *AuthService) UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*entity.User, error) {
	return s.userRepo.UpdateUserSettings(ctx, userID, patch)
}

func (s *AuthService) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (*entity.User, error) {
	if s.s3 == nil {
		return nil, errors.New("storage not configured")
	}

	key, err := s.s3.UploadAvatar(ctx, userID, file)
	if err != nil {
		return nil, err
	}

	url := s.s3.AvatarURL(s.cfg.PublicBaseURL, key)
	return s.userRepo.UpdateUserProfile(ctx, userID, entity.UpdateUserParams{
		AvatarURL: &url,
	})
}

func (s *AuthService) UpdateMyProfile(ctx context.Context, userID string, displayName, email, phone, username *string) (*entity.User, error) {
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

	return s.userRepo.UpdateUserProfile(ctx, userID, params)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
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
		return errors.New("invalid current password")
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

func (s *AuthService) generateToken(user entity.User) (string, int, error) {
	exp := time.Now().Add(time.Duration(s.cfg.JWTAccessTTL) * time.Minute)
	claims := jwt.MapClaims{
		"sub": user.ID,
		"iat": time.Now().Unix(),
		"exp": exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, s.cfg.JWTAccessTTL * 60, nil
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
