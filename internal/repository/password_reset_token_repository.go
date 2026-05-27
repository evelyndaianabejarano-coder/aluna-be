package repository

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"gorm.io/gorm"
)

type PasswordResetTokenRepository interface {
	Create(token *models.PasswordResetToken) error
	FindByToken(tokenHash string) (*models.PasswordResetToken, error)
	MarkUsed(token *models.PasswordResetToken) error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetTokenRepository) FindByToken(tokenHash string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	if err := r.db.Where("token = ?", tokenHash).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *passwordResetTokenRepository) MarkUsed(token *models.PasswordResetToken) error {
	return r.db.Save(token).Error
}
