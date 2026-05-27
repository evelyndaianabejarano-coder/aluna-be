package repository

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationRepository interface {
	FindByID(id uuid.UUID) (*models.Reservation, error)
	FindByAlumnoID(alumnoID uuid.UUID, offset, limit int) ([]models.Reservation, int64, error)
	FindByClaseID(claseID uuid.UUID) ([]models.Reservation, error)
	Create(r *models.Reservation) error
	Cancel(id uuid.UUID) error
	HasActiveReservation(alumnoID, claseID uuid.UUID) (bool, error)
	CountActiveByClaseID(claseID uuid.UUID) (int64, error)
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindByID(id uuid.UUID) (*models.Reservation, error) {
	var res models.Reservation
	err := r.db.
		Preload("Alumno").
		Preload("Clase").
		First(&res, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *reservationRepository) FindByAlumnoID(alumnoID uuid.UUID, offset, limit int) ([]models.Reservation, int64, error) {
	q := r.db.Model(&models.Reservation{}).Where("alumno_id = ?", alumnoID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var reservations []models.Reservation
	err := q.
		Preload("Clase").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reservations).Error

	return reservations, total, err
}

func (r *reservationRepository) FindByClaseID(claseID uuid.UUID) ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.
		Preload("Alumno").
		Where("clase_id = ? AND estado IN ?", claseID, []models.EstadoReserva{
			models.EstadoReservaPendiente,
			models.EstadoReservaConfirmada,
		}).
		Order("created_at ASC").
		Find(&reservations).Error
	return reservations, err
}

func (r *reservationRepository) Create(res *models.Reservation) error {
	return r.db.Create(res).Error
}

func (r *reservationRepository) Cancel(id uuid.UUID) error {
	return r.db.Model(&models.Reservation{}).
		Where("id = ?", id).
		Update("estado", models.EstadoReservaCancelada).Error
}

func (r *reservationRepository) HasActiveReservation(alumnoID, claseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("alumno_id = ? AND clase_id = ? AND estado IN ?", alumnoID, claseID,
			[]models.EstadoReserva{models.EstadoReservaPendiente, models.EstadoReservaConfirmada}).
		Count(&count).Error
	return count > 0, err
}

func (r *reservationRepository) CountActiveByClaseID(claseID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("clase_id = ? AND estado IN ?", claseID,
			[]models.EstadoReserva{models.EstadoReservaPendiente, models.EstadoReservaConfirmada}).
		Count(&count).Error
	return count, err
}
