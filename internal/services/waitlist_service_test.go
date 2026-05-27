package services_test

import (
	"testing"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	repomocks "github.com/evelyndaianabejarano-coder/aluna-be/internal/repository/mocks"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newWaitlistSvc(
	waitRepo *repomocks.WaitlistRepository,
	reservRepo *repomocks.ReservationRepository,
	classRepo *repomocks.ClassRepository,
) services.WaitlistService {
	return services.NewWaitlistService(waitRepo, reservRepo, classRepo)
}

// --- Join ---

func TestWaitlist_Join_ClaseConCuposDisponibles(t *testing.T) {
	waitRepo := &repomocks.WaitlistRepository{}
	reservRepo := &repomocks.ReservationRepository{}
	classRepo := &repomocks.ClassRepository{}

	claseID := uuid.New()
	class := activeClass(10)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(5), nil) // 5 < 10 → hay cupos

	svc := newWaitlistSvc(waitRepo, reservRepo, classRepo)
	err := svc.Join(uuid.New(), claseID)

	assert.ErrorIs(t, err, apperrors.ErrClassNotFull)
	waitRepo.AssertNotCalled(t, "Add")
}

func TestWaitlist_Join_DuplicadoEnLista(t *testing.T) {
	waitRepo := &repomocks.WaitlistRepository{}
	reservRepo := &repomocks.ReservationRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	class := activeClass(5)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(5), nil) // llena
	waitRepo.On("HasEntry", alumnoID, claseID).Return(true, nil)

	svc := newWaitlistSvc(waitRepo, reservRepo, classRepo)
	err := svc.Join(alumnoID, claseID)

	assert.ErrorIs(t, err, apperrors.ErrWaitlistAlreadyJoined)
	waitRepo.AssertNotCalled(t, "Add")
}

func TestWaitlist_Join_Exitoso(t *testing.T) {
	waitRepo := &repomocks.WaitlistRepository{}
	reservRepo := &repomocks.ReservationRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	class := activeClass(5)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(5), nil) // llena
	waitRepo.On("HasEntry", alumnoID, claseID).Return(false, nil)
	waitRepo.On("Add", mock.AnythingOfType("*models.Waitlist")).Return(nil)

	svc := newWaitlistSvc(waitRepo, reservRepo, classRepo)
	err := svc.Join(alumnoID, claseID)

	assert.NoError(t, err)
	waitRepo.AssertExpectations(t)

	added := waitRepo.Calls[1].Arguments.Get(0).(*models.Waitlist)
	assert.Equal(t, alumnoID, added.AlumnoID)
	assert.Equal(t, claseID, added.ClaseID)
}

func TestWaitlist_Join_ClaseCancelada(t *testing.T) {
	waitRepo := &repomocks.WaitlistRepository{}
	reservRepo := &repomocks.ReservationRepository{}
	classRepo := &repomocks.ClassRepository{}

	claseID := uuid.New()
	class := activeClass(10)
	class.ID = claseID
	class.Estado = models.EstadoCancelada

	classRepo.On("FindByID", claseID).Return(class, nil)

	svc := newWaitlistSvc(waitRepo, reservRepo, classRepo)
	err := svc.Join(uuid.New(), claseID)

	assert.ErrorIs(t, err, apperrors.ErrClassNotFound)
	waitRepo.AssertNotCalled(t, "Add")
}
