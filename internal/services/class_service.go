package services

import (
	"errors"
	"time"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassDTO struct {
	ID               string    `json:"id"`
	Titulo           string    `json:"titulo"`
	Descripcion      string    `json:"descripcion"`
	ProfesorID       string    `json:"profesor_id"`
	ProfesorNombre   string    `json:"profesor_nombre"`
	CategoriaID      string    `json:"categoria_id"`
	CategoriaNombre  string    `json:"categoria_nombre"`
	FechaHora        time.Time `json:"fecha_hora"`
	Duracion         int       `json:"duracion"`
	CupoMaximo       int       `json:"cupo_maximo"`
	CuposDisponibles int       `json:"cupos_disponibles"`
	Modalidad        string    `json:"modalidad"`
	Estado           string    `json:"estado"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateClassRequest struct {
	Titulo      string          `json:"titulo"       binding:"required"`
	Descripcion string          `json:"descripcion"  binding:"required"`
	CategoriaID string          `json:"categoria_id" binding:"required,uuid"`
	FechaHora   time.Time       `json:"fecha_hora"   binding:"required"`
	Duracion    int             `json:"duracion"     binding:"required,min=1"`
	CupoMaximo  int             `json:"cupo_maximo"  binding:"required,min=1"`
	Modalidad   models.Modalidad `json:"modalidad"   binding:"required,oneof=presencial online"`
	ProfesorID  string          `json:"profesor_id"` // solo admin puede setearlo
}

type UpdateClassRequest struct {
	Titulo      string           `json:"titulo"`
	Descripcion string           `json:"descripcion"`
	CategoriaID string           `json:"categoria_id" binding:"omitempty,uuid"`
	FechaHora   *time.Time       `json:"fecha_hora"`
	Duracion    int              `json:"duracion"     binding:"omitempty,min=1"`
	CupoMaximo  int              `json:"cupo_maximo"  binding:"omitempty,min=1"`
	Modalidad   models.Modalidad `json:"modalidad"    binding:"omitempty,oneof=presencial online"`
}

type ClassService interface {
	List(filters repository.ClassFilters) ([]ClassDTO, int64, error)
	Get(id uuid.UUID) (*ClassDTO, error)
	Create(req CreateClassRequest, userID uuid.UUID, role models.Role) (*ClassDTO, error)
	Update(id uuid.UUID, req UpdateClassRequest, userID uuid.UUID, role models.Role) (*ClassDTO, error)
	Cancel(id uuid.UUID, userID uuid.UUID, role models.Role) error
}

type classService struct {
	classRepo repository.ClassRepository
}

func NewClassService(classRepo repository.ClassRepository) ClassService {
	return &classService{classRepo: classRepo}
}

func (s *classService) List(filters repository.ClassFilters) ([]ClassDTO, int64, error) {
	classes, total, err := s.classRepo.List(filters)
	if err != nil {
		return nil, 0, apperrors.ErrInternal
	}
	dtos := make([]ClassDTO, len(classes))
	for i, c := range classes {
		// CountActiveReservations siempre devuelve 0 en Fase 3.
		// En Fase 4 se conecta a reservas; por ahora cupos = cupo_maximo.
		dtos[i] = toClassDTO(&c, c.CupoMaximo)
	}
	return dtos, total, nil
}

func (s *classService) Get(id uuid.UUID) (*ClassDTO, error) {
	class, err := s.classRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClassNotFound
		}
		return nil, apperrors.ErrInternal
	}

	reservas, err := s.classRepo.CountActiveReservations(id)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	cuposDisponibles := max(class.CupoMaximo-int(reservas), 0)

	dto := toClassDTO(class, cuposDisponibles)
	return &dto, nil
}

func (s *classService) Create(req CreateClassRequest, userID uuid.UUID, role models.Role) (*ClassDTO, error) {
	if req.FechaHora.Before(time.Now()) {
		return nil, apperrors.ErrClassInvalidDate
	}

	categoriaID, err := uuid.Parse(req.CategoriaID)
	if err != nil {
		return nil, apperrors.ErrValidation
	}

	// Profesores solo pueden crear clases propias
	profesorID := userID
	if role == models.RoleAdmin && req.ProfesorID != "" {
		if parsed, err := uuid.Parse(req.ProfesorID); err == nil {
			profesorID = parsed
		}
	}

	class := &models.Class{
		Titulo:      req.Titulo,
		Descripcion: req.Descripcion,
		ProfesorID:  profesorID,
		CategoriaID: categoriaID,
		FechaHora:   req.FechaHora,
		Duracion:    req.Duracion,
		CupoMaximo:  req.CupoMaximo,
		Modalidad:   req.Modalidad,
		Estado:      models.EstadoActiva,
	}
	if err := s.classRepo.Create(class); err != nil {
		return nil, apperrors.ErrInternal
	}

	created, err := s.classRepo.FindByID(class.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	dto := toClassDTO(created, created.CupoMaximo)
	return &dto, nil
}

func (s *classService) Update(id uuid.UUID, req UpdateClassRequest, userID uuid.UUID, role models.Role) (*ClassDTO, error) {
	class, err := s.classRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClassNotFound
		}
		return nil, apperrors.ErrInternal
	}

	if !isOwnerOrAdmin(class.ProfesorID, userID, role) {
		return nil, apperrors.ErrUnauthorized
	}

	if req.Titulo != "" {
		class.Titulo = req.Titulo
	}
	if req.Descripcion != "" {
		class.Descripcion = req.Descripcion
	}
	if req.CategoriaID != "" {
		if parsed, err := uuid.Parse(req.CategoriaID); err == nil {
			class.CategoriaID = parsed
		}
	}
	if req.FechaHora != nil {
		if req.FechaHora.Before(time.Now()) {
			return nil, apperrors.ErrClassInvalidDate
		}
		class.FechaHora = *req.FechaHora
	}
	if req.Duracion > 0 {
		class.Duracion = req.Duracion
	}
	if req.CupoMaximo > 0 {
		class.CupoMaximo = req.CupoMaximo
	}
	if req.Modalidad != "" {
		class.Modalidad = req.Modalidad
	}

	if err := s.classRepo.Update(class); err != nil {
		return nil, apperrors.ErrInternal
	}

	updated, err := s.classRepo.FindByID(id)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	reservas, err := s.classRepo.CountActiveReservations(id)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	cuposDisponibles := max(updated.CupoMaximo-int(reservas), 0)
	dto := toClassDTO(updated, cuposDisponibles)
	return &dto, nil
}

func (s *classService) Cancel(id uuid.UUID, userID uuid.UUID, role models.Role) error {
	class, err := s.classRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrClassNotFound
		}
		return apperrors.ErrInternal
	}

	if class.Estado == models.EstadoCancelada {
		return apperrors.ErrClassAlreadyCancelled
	}

	if !isOwnerOrAdmin(class.ProfesorID, userID, role) {
		return apperrors.ErrUnauthorized
	}

	if err := s.classRepo.Cancel(id); err != nil {
		return apperrors.ErrInternal
	}
	return nil
}

func isOwnerOrAdmin(ownerID, userID uuid.UUID, role models.Role) bool {
	return role == models.RoleAdmin || ownerID == userID
}

func toClassDTO(c *models.Class, cuposDisponibles int) ClassDTO {
	dto := ClassDTO{
		ID:               c.ID.String(),
		Titulo:           c.Titulo,
		Descripcion:      c.Descripcion,
		ProfesorID:       c.ProfesorID.String(),
		CategoriaID:      c.CategoriaID.String(),
		FechaHora:        c.FechaHora,
		Duracion:         c.Duracion,
		CupoMaximo:       c.CupoMaximo,
		CuposDisponibles: cuposDisponibles,
		Modalidad:        string(c.Modalidad),
		Estado:           string(c.Estado),
		CreatedAt:        c.CreatedAt,
	}
	if c.Profesor != nil {
		dto.ProfesorNombre = c.Profesor.Nombre
	}
	if c.Categoria != nil {
		dto.CategoriaNombre = c.Categoria.Nombre
	}
	return dto
}
