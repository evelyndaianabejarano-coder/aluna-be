package mocks

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ReservationRepository struct {
	mock.Mock
}

func (m *ReservationRepository) FindByID(id uuid.UUID) (*models.Reservation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reservation), args.Error(1)
}

func (m *ReservationRepository) FindByAlumnoID(alumnoID uuid.UUID, offset, limit int) ([]models.Reservation, int64, error) {
	args := m.Called(alumnoID, offset, limit)
	return args.Get(0).([]models.Reservation), args.Get(1).(int64), args.Error(2)
}

func (m *ReservationRepository) FindByClaseID(claseID uuid.UUID) ([]models.Reservation, error) {
	args := m.Called(claseID)
	return args.Get(0).([]models.Reservation), args.Error(1)
}

func (m *ReservationRepository) Create(r *models.Reservation) error {
	return m.Called(r).Error(0)
}

func (m *ReservationRepository) Cancel(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *ReservationRepository) HasActiveReservation(alumnoID, claseID uuid.UUID) (bool, error) {
	args := m.Called(alumnoID, claseID)
	return args.Bool(0), args.Error(1)
}

func (m *ReservationRepository) CountActiveByClaseID(claseID uuid.UUID) (int64, error) {
	args := m.Called(claseID)
	return args.Get(0).(int64), args.Error(1)
}
