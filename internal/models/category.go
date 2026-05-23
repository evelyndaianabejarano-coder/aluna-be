package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"  json:"id"`
	Nombre      string    `gorm:"uniqueIndex;not null"  json:"nombre"`
	Objetivo    string    `gorm:"not null"              json:"objetivo"`
	Descripcion string    `gorm:"not null"              json:"descripcion"`
	CreatedAt   time.Time `                             json:"created_at"`
	UpdatedAt   time.Time `                             json:"updated_at"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
