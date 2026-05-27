package repository

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WaitlistRepository interface {
	Add(entry *models.Waitlist) error
	Remove(alumnoID, claseID uuid.UUID) error
	GetFirst(claseID uuid.UUID) (*models.Waitlist, error)
	HasEntry(alumnoID, claseID uuid.UUID) (bool, error)
	CountByClaseID(claseID uuid.UUID) (int64, error)
}

type waitlistRepository struct {
	db *gorm.DB
}

func NewWaitlistRepository(db *gorm.DB) WaitlistRepository {
	return &waitlistRepository{db: db}
}

func (r *waitlistRepository) Add(entry *models.Waitlist) error {
	return r.db.Create(entry).Error
}

func (r *waitlistRepository) Remove(alumnoID, claseID uuid.UUID) error {
	return r.db.
		Where("alumno_id = ? AND clase_id = ?", alumnoID, claseID).
		Delete(&models.Waitlist{}).Error
}

func (r *waitlistRepository) GetFirst(claseID uuid.UUID) (*models.Waitlist, error) {
	var entry models.Waitlist
	err := r.db.
		Where("clase_id = ?", claseID).
		Order("created_at ASC").
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *waitlistRepository) HasEntry(alumnoID, claseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Waitlist{}).
		Where("alumno_id = ? AND clase_id = ?", alumnoID, claseID).
		Count(&count).Error
	return count > 0, err
}

func (r *waitlistRepository) CountByClaseID(claseID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Waitlist{}).
		Where("clase_id = ?", claseID).
		Count(&count).Error
	return count, err
}
