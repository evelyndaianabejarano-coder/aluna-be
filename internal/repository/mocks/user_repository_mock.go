package mocks

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) Create(user *models.User) error {
	return m.Called(user).Error(0)
}

func (m *UserRepository) Update(user *models.User) error {
	return m.Called(user).Error(0)
}

func (m *UserRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *UserRepository) List(offset, limit int) ([]models.User, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}
