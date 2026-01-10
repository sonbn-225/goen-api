package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/mocks"
	"github.com/sonbn-225/goen-api/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Signup(t *testing.T) {
	// Setup Config
	cfg := &config.Config{
		JWTSecret:           "test-secret",
		JWTAccessTTLMinutes: 60,
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserRepository)
		service := services.NewAuthService(mockRepo, cfg)

		req := services.SignupRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}

		// Expect CreateUser to be called once with any UserWithPassword object
		mockRepo.On("CreateUser", context.Background(), mock.MatchedBy(func(u domain.UserWithPassword) bool {
			return *u.Email == "test@example.com" && u.PasswordHash != "" && u.PasswordHash != "password123"
		})).Return(nil)

		// Act
		resp, err := service.Signup(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "test@example.com", *resp.User.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Fail_UserAlreadyExists", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserRepository)
		service := services.NewAuthService(mockRepo, cfg)

		req := services.SignupRequest{
			Email:    "exists@example.com",
			Password: "password123",
		}

		mockRepo.On("CreateUser", context.Background(), mock.Anything).Return(domain.ErrUserAlreadyExists)

		// Act
		resp, err := service.Signup(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		var se *services.ServiceError
		if assert.True(t, errors.As(err, &se)) {
			assert.Equal(t, services.ErrorKindConflict, se.Kind)
			assert.Equal(t, "user already exists", se.Message)
			assert.True(t, errors.Is(err, domain.ErrUserAlreadyExists))
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("Fail_Validation_ShortPassword", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserRepository)
		service := services.NewAuthService(mockRepo, cfg)

		req := services.SignupRequest{
			Email:    "valid@example.com",
			Password: "short", // < 8 chars
		}

		// Act
		resp, err := service.Signup(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		var se *services.ServiceError
		if assert.True(t, errors.As(err, &se)) {
			assert.Equal(t, services.ErrorKindValidation, se.Kind)
			assert.Contains(t, se.Message, "password must be at least 8 characters")
		}

		// Ensure repo was NOT called
		mockRepo.AssertNotCalled(t, "CreateUser")
	})
}
