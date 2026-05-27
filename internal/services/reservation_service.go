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

type ReservationDTO struct {
	ID        string     `json:"id"`
	AlumnoID  string     `json:"alumno_id"`
	ClaseID   string     `json:"clase_id"`
	Estado    string     `json:"estado"`
	Asistio   *bool      `json:"asistio"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Clase     *ClassDTO  `json:"clase,omitempty"`
}

type AttendanceRecord struct {
	AlumnoID uuid.UUID
	Presente bool
}

type ReservationService interface {
	Reserve(alumnoID, claseID uuid.UUID) (*ReservationDTO, error)
	Cancel(id, userID uuid.UUID, role models.Role) error
	GetMy(alumnoID uuid.UUID, page, limit int) ([]ReservationDTO, int64, error)
	Get(id, userID uuid.UUID, role models.Role) (*ReservationDTO, error)
	GetStudents(claseID uuid.UUID) ([]ReservationDTO, error)
	MarkAttendance(claseID uuid.UUID, attendances []AttendanceRecord, userID uuid.UUID, role models.Role) error
}

type reservationService struct {
	db            *gorm.DB
	reservRepo    repository.ReservationRepository
	waitRepo      repository.WaitlistRepository
	classRepo     repository.ClassRepository
	newReservRepo func(*gorm.DB) repository.ReservationRepository
	newWaitRepo   func(*gorm.DB) repository.WaitlistRepository
	newClassRepo  func(*gorm.DB) repository.ClassRepository
	txFunc        func(func(*gorm.DB) error) error
}

func NewReservationService(
	db *gorm.DB,
	reservRepo repository.ReservationRepository,
	waitRepo repository.WaitlistRepository,
	classRepo repository.ClassRepository,
) ReservationService {
	return &reservationService{
		db:            db,
		reservRepo:    reservRepo,
		waitRepo:      waitRepo,
		classRepo:     classRepo,
		newReservRepo: repository.NewReservationRepository,
		newWaitRepo:   repository.NewWaitlistRepository,
		newClassRepo:  repository.NewClassRepository,
		txFunc:        func(fc func(*gorm.DB) error) error { return db.Transaction(fc) },
	}
}

// NewReservationServiceForTest permite inyectar repos mockeados y un txFunc sin DB real.
// Los factories reciben un *gorm.DB (nil en tests) y devuelven el mock directamente.
func NewReservationServiceForTest(
	reservRepo repository.ReservationRepository,
	waitRepo repository.WaitlistRepository,
	classRepo repository.ClassRepository,
	txFunc func(func(*gorm.DB) error) error,
) ReservationService {
	return &reservationService{
		reservRepo:    reservRepo,
		waitRepo:      waitRepo,
		classRepo:     classRepo,
		newReservRepo: func(_ *gorm.DB) repository.ReservationRepository { return reservRepo },
		newWaitRepo:   func(_ *gorm.DB) repository.WaitlistRepository { return waitRepo },
		newClassRepo:  func(_ *gorm.DB) repository.ClassRepository { return classRepo },
		txFunc:        txFunc,
	}
}

func (s *reservationService) Reserve(alumnoID, claseID uuid.UUID) (*ReservationDTO, error) {
	var dto *ReservationDTO

	err := s.txFunc(func(tx *gorm.DB) error {
		txReservRepo := s.newReservRepo(tx)
		txClassRepo := s.newClassRepo(tx)

		class, err := txClassRepo.FindByID(claseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrClassNotFound
			}
			return apperrors.ErrInternal
		}
		if class.Estado != models.EstadoActiva {
			return apperrors.ErrClassNotFound
		}

		count, err := txReservRepo.CountActiveByClaseID(claseID)
		if err != nil {
			return apperrors.ErrInternal
		}
		if int(count) >= class.CupoMaximo {
			return apperrors.ErrReservationClassFull
		}

		has, err := txReservRepo.HasActiveReservation(alumnoID, claseID)
		if err != nil {
			return apperrors.ErrInternal
		}
		if has {
			return apperrors.ErrReservationConflict
		}

		res := &models.Reservation{
			AlumnoID: alumnoID,
			ClaseID:  claseID,
			Estado:   models.EstadoReservaPendiente,
		}
		if err := txReservRepo.Create(res); err != nil {
			return apperrors.ErrInternal
		}

		created, err := txReservRepo.FindByID(res.ID)
		if err != nil {
			return apperrors.ErrInternal
		}
		result := toReservationDTO(created)
		dto = &result
		return nil
	})

	if err != nil {
		return nil, err
	}
	return dto, nil
}

func (s *reservationService) Cancel(id, userID uuid.UUID, role models.Role) error {
	return s.txFunc(func(tx *gorm.DB) error {
		txReservRepo := s.newReservRepo(tx)
		txWaitRepo := s.newWaitRepo(tx)

		res, err := txReservRepo.FindByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrReservationNotFound
			}
			return apperrors.ErrInternal
		}

		if !isOwnerOrAdmin(res.AlumnoID, userID, role) {
			return apperrors.ErrUnauthorized
		}
		if res.Estado == models.EstadoReservaCancelada {
			return apperrors.ErrReservationAlreadyCancelled
		}

		if err := txReservRepo.Cancel(id); err != nil {
			return apperrors.ErrInternal
		}

		// Promover al primero de la lista de espera, si existe
		first, err := txWaitRepo.GetFirst(res.ClaseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return apperrors.ErrInternal
		}

		promoted := &models.Reservation{
			AlumnoID: first.AlumnoID,
			ClaseID:  res.ClaseID,
			Estado:   models.EstadoReservaPendiente,
		}
		if err := txReservRepo.Create(promoted); err != nil {
			return apperrors.ErrInternal
		}
		if err := txWaitRepo.Remove(first.AlumnoID, res.ClaseID); err != nil {
			return apperrors.ErrInternal
		}

		return nil
	})
}

func (s *reservationService) GetMy(alumnoID uuid.UUID, page, limit int) ([]ReservationDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	reservations, total, err := s.reservRepo.FindByAlumnoID(alumnoID, offset, limit)
	if err != nil {
		return nil, 0, apperrors.ErrInternal
	}

	dtos := make([]ReservationDTO, len(reservations))
	for i, r := range reservations {
		dtos[i] = toReservationDTO(&r)
	}
	return dtos, total, nil
}

func (s *reservationService) Get(id, userID uuid.UUID, role models.Role) (*ReservationDTO, error) {
	res, err := s.reservRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrReservationNotFound
		}
		return nil, apperrors.ErrInternal
	}

	if !isOwnerOrAdmin(res.AlumnoID, userID, role) {
		return nil, apperrors.ErrUnauthorized
	}

	dto := toReservationDTO(res)
	return &dto, nil
}

func (s *reservationService) GetStudents(claseID uuid.UUID) ([]ReservationDTO, error) {
	reservations, err := s.reservRepo.FindByClaseID(claseID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	dtos := make([]ReservationDTO, len(reservations))
	for i, r := range reservations {
		dtos[i] = toReservationDTO(&r)
	}
	return dtos, nil
}

func (s *reservationService) MarkAttendance(claseID uuid.UUID, attendances []AttendanceRecord, userID uuid.UUID, role models.Role) error {
	class, err := s.classRepo.FindByID(claseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrClassNotFound
		}
		return apperrors.ErrInternal
	}

	if !isOwnerOrAdmin(class.ProfesorID, userID, role) {
		return apperrors.ErrUnauthorized
	}

	activeStates := []models.EstadoReserva{
		models.EstadoReservaPendiente,
		models.EstadoReservaConfirmada,
	}

	for _, a := range attendances {
		err := s.db.Model(&models.Reservation{}).
			Where("clase_id = ? AND alumno_id = ? AND estado IN ?", claseID, a.AlumnoID, activeStates).
			Update("asistio", a.Presente).Error
		if err != nil {
			return apperrors.ErrInternal
		}
	}
	return nil
}

func toReservationDTO(r *models.Reservation) ReservationDTO {
	dto := ReservationDTO{
		ID:        r.ID.String(),
		AlumnoID:  r.AlumnoID.String(),
		ClaseID:   r.ClaseID.String(),
		Estado:    string(r.Estado),
		Asistio:   r.Asistio,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Clase != nil {
		classDTO := toClassDTO(r.Clase, 0)
		dto.Clase = &classDTO
	}
	return dto
}
