package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAlumno   Role = "alumno"
	RoleProfesor Role = "profesor"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"      json:"id"`
	Email        string    `gorm:"uniqueIndex;not null"       json:"email"`
	PasswordHash string    `gorm:"not null"                   json:"-"`
	Role         Role      `gorm:"not null;default:'alumno'"  json:"role"`
	Nombre       string    `gorm:"not null"                   json:"nombre"`
	CreatedAt    time.Time `                                  json:"created_at"`
	UpdatedAt    time.Time `                                  json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
