package services

import (
	"errors"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WaitlistService interface {
	Join(alumnoID, claseID uuid.UUID) error
}

type waitlistService struct {
	waitRepo   repository.WaitlistRepository
	reservRepo repository.ReservationRepository
	classRepo  repository.ClassRepository
}

func NewWaitlistService(
	waitRepo repository.WaitlistRepository,
	reservRepo repository.ReservationRepository,
	classRepo repository.ClassRepository,
) WaitlistService {
	return &waitlistService{
		waitRepo:   waitRepo,
		reservRepo: reservRepo,
		classRepo:  classRepo,
	}
}

func (s *waitlistService) Join(alumnoID, claseID uuid.UUID) error {
	class, err := s.classRepo.FindByID(claseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrClassNotFound
		}
		return apperrors.ErrInternal
	}
	if class.Estado != models.EstadoActiva {
		return apperrors.ErrClassNotFound
	}

	count, err := s.reservRepo.CountActiveByClaseID(claseID)
	if err != nil {
		return apperrors.ErrInternal
	}
	if int(count) < class.CupoMaximo {
		return apperrors.ErrClassNotFull
	}

	has, err := s.waitRepo.HasEntry(alumnoID, claseID)
	if err != nil {
		return apperrors.ErrInternal
	}
	if has {
		return apperrors.ErrWaitlistAlreadyJoined
	}

	entry := &models.Waitlist{
		AlumnoID: alumnoID,
		ClaseID:  claseID,
	}
	if err := s.waitRepo.Add(entry); err != nil {
		return apperrors.ErrInternal
	}
	return nil
}
