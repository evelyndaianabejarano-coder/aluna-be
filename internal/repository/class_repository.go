package repository

import (
	"time"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassFilters struct {
	CategoriaID *uuid.UUID
	ProfesorID  *uuid.UUID
	Modalidad   *models.Modalidad
	Fecha       *time.Time // filtra el día completo (00:00–23:59 UTC)
	Offset      int
	Limit       int
}

type ClassRepository interface {
	List(filters ClassFilters) ([]models.Class, int64, error)
	FindByID(id uuid.UUID) (*models.Class, error)
	Create(class *models.Class) error
	Update(class *models.Class) error
	Cancel(id uuid.UUID) error
	CountActiveReservations(classID uuid.UUID) (int64, error)
}

type classRepository struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) List(f ClassFilters) ([]models.Class, int64, error) {
	q := r.db.Model(&models.Class{})

	if f.CategoriaID != nil {
		q = q.Where("categoria_id = ?", *f.CategoriaID)
	}
	if f.ProfesorID != nil {
		q = q.Where("profesor_id = ?", *f.ProfesorID)
	}
	if f.Modalidad != nil {
		q = q.Where("modalidad = ?", *f.Modalidad)
	}
	if f.Fecha != nil {
		start := time.Date(f.Fecha.Year(), f.Fecha.Month(), f.Fecha.Day(), 0, 0, 0, 0, time.UTC)
		end := start.Add(24 * time.Hour)
		q = q.Where("fecha_hora >= ? AND fecha_hora < ?", start, end)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var classes []models.Class
	err := q.
		Preload("Categoria").
		Preload("Profesor").
		Order("fecha_hora ASC").
		Offset(f.Offset).
		Limit(limit).
		Find(&classes).Error

	return classes, total, err
}

func (r *classRepository) FindByID(id uuid.UUID) (*models.Class, error) {
	var class models.Class
	err := r.db.
		Preload("Categoria").
		Preload("Profesor").
		First(&class, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *classRepository) Create(class *models.Class) error {
	return r.db.Create(class).Error
}

func (r *classRepository) Update(class *models.Class) error {
	return r.db.Save(class).Error
}

func (r *classRepository) Cancel(id uuid.UUID) error {
	return r.db.Model(&models.Class{}).
		Where("id = ?", id).
		Update("estado", models.EstadoCancelada).Error
}

func (r *classRepository) CountActiveReservations(classID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("reservations").
		Where("clase_id = ? AND estado IN ?", classID,
			[]string{"pendiente", "confirmada"}).
		Count(&count).Error
	return count, err
}
