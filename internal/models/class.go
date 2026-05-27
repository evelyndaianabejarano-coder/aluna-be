package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Modalidad string
type EstadoClase string

const (
	ModalidadPresencial Modalidad   = "presencial"
	ModalidadOnline     Modalidad   = "online"
	EstadoActiva        EstadoClase = "activa"
	EstadoCancelada     EstadoClase = "cancelada"
)

type Class struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"               json:"id"`
	Titulo      string      `gorm:"not null"                           json:"titulo"`
	Descripcion string      `gorm:"not null"                           json:"descripcion"`
	ProfesorID  uuid.UUID   `gorm:"type:uuid;not null;index"           json:"profesor_id"`
	CategoriaID uuid.UUID   `gorm:"type:uuid;not null;index"           json:"categoria_id"`
	FechaHora   time.Time   `gorm:"not null;index"                     json:"fecha_hora"`
	Duracion    int         `gorm:"not null"                           json:"duracion"`
	CupoMaximo  int         `gorm:"not null"                           json:"cupo_maximo"`
	Modalidad   Modalidad   `gorm:"not null;default:'presencial';index" json:"modalidad"`
	Estado      EstadoClase `gorm:"not null;default:'activa';index"    json:"estado"`
	CreatedAt   time.Time   `                                          json:"created_at"`
	UpdatedAt   time.Time   `                                          json:"updated_at"`

	// Relaciones — se cargan con Preload
	Profesor  *User     `gorm:"foreignKey:ProfesorID"  json:"profesor,omitempty"`
	Categoria *Category `gorm:"foreignKey:CategoriaID" json:"categoria,omitempty"`
}

func (c *Class) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
