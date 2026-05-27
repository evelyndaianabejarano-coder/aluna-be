package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"not null;index"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
}

func (p *PasswordResetToken) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
