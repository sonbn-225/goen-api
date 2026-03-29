package mocks

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.UserWithPassword) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*domain.UserWithPassword, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithPassword), args.Error(1)
}

func (m *MockUserRepository) FindUserByPhone(ctx context.Context, phone string) (*domain.UserWithPassword, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithPassword), args.Error(1)
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error) {
	args := m.Called(ctx, userID, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserProfile(ctx context.Context, userID string, params domain.UpdateUserParams) (*domain.User, error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
