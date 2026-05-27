package mocks

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type WaitlistRepository struct {
	mock.Mock
}

func (m *WaitlistRepository) Add(entry *models.Waitlist) error {
	return m.Called(entry).Error(0)
}

func (m *WaitlistRepository) Remove(alumnoID, claseID uuid.UUID) error {
	return m.Called(alumnoID, claseID).Error(0)
}

func (m *WaitlistRepository) GetFirst(claseID uuid.UUID) (*models.Waitlist, error) {
	args := m.Called(claseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Waitlist), args.Error(1)
}

func (m *WaitlistRepository) HasEntry(alumnoID, claseID uuid.UUID) (bool, error) {
	args := m.Called(alumnoID, claseID)
	return args.Bool(0), args.Error(1)
}

func (m *WaitlistRepository) CountByClaseID(claseID uuid.UUID) (int64, error) {
	args := m.Called(claseID)
	return args.Get(0).(int64), args.Error(1)
}
