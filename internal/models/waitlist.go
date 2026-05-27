package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Waitlist struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"     json:"id"`
	AlumnoID uuid.UUID `gorm:"type:uuid;not null;index" json:"alumno_id"`
	ClaseID  uuid.UUID `gorm:"type:uuid;not null;index" json:"clase_id"`
	CreatedAt time.Time `                               json:"created_at"`

	// Relaciones — se cargan con Preload
	Alumno *User  `gorm:"foreignKey:AlumnoID" json:"alumno,omitempty"`
	Clase  *Class `gorm:"foreignKey:ClaseID"  json:"clase,omitempty"`
}

func (w *Waitlist) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
