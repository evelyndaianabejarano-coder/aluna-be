package mocks

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ClassRepository struct {
	mock.Mock
}

func (m *ClassRepository) List(filters repository.ClassFilters) ([]models.Class, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Class), args.Get(1).(int64), args.Error(2)
}

func (m *ClassRepository) FindByID(id uuid.UUID) (*models.Class, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Class), args.Error(1)
}

func (m *ClassRepository) Create(class *models.Class) error {
	return m.Called(class).Error(0)
}

func (m *ClassRepository) Update(class *models.Class) error {
	return m.Called(class).Error(0)
}

func (m *ClassRepository) Cancel(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *ClassRepository) CountActiveReservations(classID uuid.UUID) (int64, error) {
	args := m.Called(classID)
	return args.Get(0).(int64), args.Error(1)
}
