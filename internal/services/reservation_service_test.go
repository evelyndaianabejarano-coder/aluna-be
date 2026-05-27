package services_test

import (
	"testing"
	"time"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	repomocks "github.com/evelyndaianabejarano-coder/aluna-be/internal/repository/mocks"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// noopTx ejecuta el callback directamente sin base de datos real.
func noopTx(fc func(*gorm.DB) error) error {
	return fc(nil)
}

func newReservSvc(
	reservRepo *repomocks.ReservationRepository,
	waitRepo *repomocks.WaitlistRepository,
	classRepo *repomocks.ClassRepository,
) services.ReservationService {
	return services.NewReservationServiceForTest(reservRepo, waitRepo, classRepo, noopTx)
}

func baseReservation(alumnoID, claseID uuid.UUID) *models.Reservation {
	return &models.Reservation{
		ID:       uuid.New(),
		AlumnoID: alumnoID,
		ClaseID:  claseID,
		Estado:   models.EstadoReservaPendiente,
		Clase: &models.Class{
			Titulo:     "Hatha",
			CupoMaximo: 10,
			Estado:     models.EstadoActiva,
			FechaHora:  time.Now().Add(24 * time.Hour),
		},
	}
}

func activeClass(cupo int) *models.Class {
	return &models.Class{
		ID:         uuid.New(),
		Titulo:     "Hatha",
		CupoMaximo: cupo,
		Estado:     models.EstadoActiva,
		FechaHora:  time.Now().Add(24 * time.Hour),
	}
}

// --- Reserve ---

func TestReserva_ClaseLlena(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	claseID := uuid.New()
	class := activeClass(5)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(5), nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	_, err := svc.Reserve(uuid.New(), claseID)

	assert.ErrorIs(t, err, apperrors.ErrReservationClassFull)
	reservRepo.AssertNotCalled(t, "Create")
}

func TestReserva_DuplicadoActivo(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	class := activeClass(10)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(3), nil)
	reservRepo.On("HasActiveReservation", alumnoID, claseID).Return(true, nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	_, err := svc.Reserve(alumnoID, claseID)

	assert.ErrorIs(t, err, apperrors.ErrReservationConflict)
	reservRepo.AssertNotCalled(t, "Create")
}

func TestReserva_Exitosa_CreaPendiente(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	class := activeClass(10)
	class.ID = claseID

	classRepo.On("FindByID", claseID).Return(class, nil)
	reservRepo.On("CountActiveByClaseID", claseID).Return(int64(2), nil)
	reservRepo.On("HasActiveReservation", alumnoID, claseID).Return(false, nil)
	reservRepo.On("Create", mock.AnythingOfType("*models.Reservation")).Return(nil)
	reservRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(baseReservation(alumnoID, claseID), nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	dto, err := svc.Reserve(alumnoID, claseID)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, string(models.EstadoReservaPendiente), dto.Estado)

	created := reservRepo.Calls[2].Arguments.Get(0).(*models.Reservation)
	assert.Equal(t, models.EstadoReservaPendiente, created.Estado)
	assert.Equal(t, alumnoID, created.AlumnoID)
}

// --- Cancel ---

func TestReserva_Cancel_AlumnoAjenoNoPuede(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	otroID := uuid.New()
	reservaID := uuid.New()

	res := baseReservation(alumnoID, uuid.New())
	res.ID = reservaID

	reservRepo.On("FindByID", reservaID).Return(res, nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	err := svc.Cancel(reservaID, otroID, models.RoleAlumno)

	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	reservRepo.AssertNotCalled(t, "Cancel")
}

func TestReserva_Cancel_YaCancelada(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	reservaID := uuid.New()

	res := baseReservation(alumnoID, uuid.New())
	res.ID = reservaID
	res.Estado = models.EstadoReservaCancelada

	reservRepo.On("FindByID", reservaID).Return(res, nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	err := svc.Cancel(reservaID, alumnoID, models.RoleAlumno)

	assert.ErrorIs(t, err, apperrors.ErrReservationAlreadyCancelled)
	reservRepo.AssertNotCalled(t, "Cancel")
}

func TestReserva_Cancel_DueñoSinWaitlist(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	reservaID := uuid.New()

	res := baseReservation(alumnoID, claseID)
	res.ID = reservaID

	reservRepo.On("FindByID", reservaID).Return(res, nil)
	reservRepo.On("Cancel", reservaID).Return(nil)
	waitRepo.On("GetFirst", claseID).Return(nil, gorm.ErrRecordNotFound)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	err := svc.Cancel(reservaID, alumnoID, models.RoleAlumno)

	assert.NoError(t, err)
	reservRepo.AssertExpectations(t)
	waitRepo.AssertNotCalled(t, "Remove")
}

func TestReserva_Cancel_PromocionAutomaticaWaitlist(t *testing.T) {
	reservRepo := &repomocks.ReservationRepository{}
	waitRepo := &repomocks.WaitlistRepository{}
	classRepo := &repomocks.ClassRepository{}

	alumnoID := uuid.New()
	claseID := uuid.New()
	reservaID := uuid.New()
	esperandoID := uuid.New()

	res := baseReservation(alumnoID, claseID)
	res.ID = reservaID

	waitEntry := &models.Waitlist{ID: uuid.New(), AlumnoID: esperandoID, ClaseID: claseID}

	reservRepo.On("FindByID", reservaID).Return(res, nil)
	reservRepo.On("Cancel", reservaID).Return(nil)
	waitRepo.On("GetFirst", claseID).Return(waitEntry, nil)
	reservRepo.On("Create", mock.AnythingOfType("*models.Reservation")).Return(nil)
	waitRepo.On("Remove", esperandoID, claseID).Return(nil)

	svc := newReservSvc(reservRepo, waitRepo, classRepo)
	err := svc.Cancel(reservaID, alumnoID, models.RoleAlumno)

	assert.NoError(t, err)
	reservRepo.AssertExpectations(t)
	waitRepo.AssertExpectations(t)

	// La nueva reserva es para el alumno promovido en estado pendiente
	promoted := reservRepo.Calls[2].Arguments.Get(0).(*models.Reservation)
	assert.Equal(t, esperandoID, promoted.AlumnoID)
	assert.Equal(t, models.EstadoReservaPendiente, promoted.Estado)
}
