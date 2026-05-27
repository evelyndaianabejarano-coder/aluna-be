package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EstadoReserva string

const (
	EstadoReservaPendiente  EstadoReserva = "pendiente"
	EstadoReservaConfirmada EstadoReserva = "confirmada"
	EstadoReservaCancelada  EstadoReserva = "cancelada"
)

type Reservation struct {
	ID       uuid.UUID     `gorm:"type:uuid;primaryKey"          json:"id"`
	AlumnoID uuid.UUID     `gorm:"type:uuid;not null;index"      json:"alumno_id"`
	ClaseID  uuid.UUID     `gorm:"type:uuid;not null;index"      json:"clase_id"`
	Estado   EstadoReserva `gorm:"not null;default:'pendiente';index" json:"estado"`
	Asistio  *bool         `gorm:"default:null"                  json:"asistio"`
	CreatedAt time.Time    `                                     json:"created_at"`
	UpdatedAt time.Time    `                                     json:"updated_at"`

	// Relaciones — se cargan con Preload
	Alumno *User  `gorm:"foreignKey:AlumnoID" json:"alumno,omitempty"`
	Clase  *Class `gorm:"foreignKey:ClaseID"  json:"clase,omitempty"`
}

func (r *Reservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
