package mocks

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type TokenService struct {
	mock.Mock
}

func (m *TokenService) GenerateAccessToken(userID uuid.UUID, role models.Role) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *TokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *TokenService) ValidateToken(tokenString string) (*services.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Claims), args.Error(1)
}

func (m *TokenService) ValidateRefreshToken(tokenString string) (*services.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Claims), args.Error(1)
}

func (m *TokenService) RevokeRefreshToken(tokenString string) error {
	return m.Called(tokenString).Error(0)
}
