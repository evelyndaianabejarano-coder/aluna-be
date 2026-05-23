package repository

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]models.Category, error)
	FindByID(id uuid.UUID) (*models.Category, error)
	Create(category *models.Category) error
	Update(category *models.Category) error
	Delete(id uuid.UUID) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Order("nombre ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Category{}, "id = ?", id).Error
}
