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

func newClassService(repo *repomocks.ClassRepository) services.ClassService {
	return services.NewClassService(repo)
}

func baseClass(profesorID, categoriaID uuid.UUID) *models.Class {
	return &models.Class{
		ID:          uuid.New(),
		Titulo:      "Hatha matutino",
		Descripcion: "Clase relajante",
		ProfesorID:  profesorID,
		CategoriaID: categoriaID,
		FechaHora:   time.Now().Add(24 * time.Hour),
		Duracion:    60,
		CupoMaximo:  10,
		Modalidad:   models.ModalidadPresencial,
		Estado:      models.EstadoActiva,
		Profesor:    &models.User{Nombre: "Profe Ana"},
		Categoria:   &models.Category{Nombre: "Hatha"},
	}
}

// --- Create ---

func TestCreate_FechaPasada(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	svc := newClassService(repo)

	req := services.CreateClassRequest{
		Titulo:      "Clase",
		Descripcion: "Desc",
		CategoriaID: uuid.New().String(),
		FechaHora:   time.Now().Add(-1 * time.Hour),
		Duracion:    60,
		CupoMaximo:  10,
		Modalidad:   models.ModalidadPresencial,
	}

	_, err := svc.Create(req, uuid.New(), models.RoleProfesor)
	assert.ErrorIs(t, err, apperrors.ErrClassInvalidDate)
	repo.AssertNotCalled(t, "Create")
}

func TestCreate_ProfesorUsaSuPropioID(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()

	repo.On("Create", mock.AnythingOfType("*models.Class")).Return(nil)
	repo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(baseClass(profesorID, categoriaID), nil)

	svc := newClassService(repo)

	otroID := uuid.New().String()
	req := services.CreateClassRequest{
		Titulo:      "Hatha",
		Descripcion: "Desc",
		CategoriaID: categoriaID.String(),
		FechaHora:   time.Now().Add(24 * time.Hour),
		Duracion:    60,
		CupoMaximo:  10,
		Modalidad:   models.ModalidadPresencial,
		ProfesorID:  otroID, // debe ser ignorado
	}

	_, err := svc.Create(req, profesorID, models.RoleProfesor)
	assert.NoError(t, err)

	// Verifica que el class creado tiene el profesorID del token, no el del body
	createdClass := repo.Calls[0].Arguments.Get(0).(*models.Class)
	assert.Equal(t, profesorID, createdClass.ProfesorID)

	repo.AssertExpectations(t)
}

func TestCreate_AdminPuedeSetearProfesorID(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	adminID := uuid.New()
	profesorEspecifico := uuid.New()
	categoriaID := uuid.New()

	repo.On("Create", mock.AnythingOfType("*models.Class")).Return(nil)
	repo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(baseClass(profesorEspecifico, categoriaID), nil)

	svc := newClassService(repo)

	req := services.CreateClassRequest{
		Titulo:      "Hatha",
		Descripcion: "Desc",
		CategoriaID: categoriaID.String(),
		FechaHora:   time.Now().Add(24 * time.Hour),
		Duracion:    60,
		CupoMaximo:  10,
		Modalidad:   models.ModalidadPresencial,
		ProfesorID:  profesorEspecifico.String(),
	}

	_, err := svc.Create(req, adminID, models.RoleAdmin)
	assert.NoError(t, err)

	createdClass := repo.Calls[0].Arguments.Get(0).(*models.Class)
	assert.Equal(t, profesorEspecifico, createdClass.ProfesorID)

	repo.AssertExpectations(t)
}

// --- Get ---

func TestGet_CuposDisponiblesCalculados(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID
	class.CupoMaximo = 10

	repo.On("FindByID", classID).Return(class, nil)
	repo.On("CountActiveReservations", classID).Return(int64(3), nil)

	svc := newClassService(repo)
	dto, err := svc.Get(classID)

	assert.NoError(t, err)
	assert.Equal(t, 7, dto.CuposDisponibles)
	repo.AssertExpectations(t)
}

func TestGet_CuposNoNegativosSiHayOverbooking(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID
	class.CupoMaximo = 5

	repo.On("FindByID", classID).Return(class, nil)
	repo.On("CountActiveReservations", classID).Return(int64(8), nil)

	svc := newClassService(repo)
	dto, err := svc.Get(classID)

	assert.NoError(t, err)
	assert.Equal(t, 0, dto.CuposDisponibles)
	repo.AssertExpectations(t)
}

func TestGet_ClaseNoEncontrada(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	classID := uuid.New()

	repo.On("FindByID", classID).Return(nil, gorm.ErrRecordNotFound)

	svc := newClassService(repo)
	_, err := svc.Get(classID)

	assert.ErrorIs(t, err, apperrors.ErrClassNotFound)
}

// --- Update ---

func TestUpdate_DueñoPuedeEditar(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil).Once()
	repo.On("Update", mock.AnythingOfType("*models.Class")).Return(nil)
	repo.On("FindByID", classID).Return(class, nil).Once()
	repo.On("CountActiveReservations", classID).Return(int64(0), nil)

	svc := newClassService(repo)
	req := services.UpdateClassRequest{Titulo: "Nuevo título"}
	_, err := svc.Update(classID, req, profesorID, models.RoleProfesor)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdate_AlumnoNoPuedeEditar(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()
	alumnoID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil)

	svc := newClassService(repo)
	req := services.UpdateClassRequest{Titulo: "Intento fallido"}
	_, err := svc.Update(classID, req, alumnoID, models.RoleAlumno)

	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdate_AdminPuedeEditarClaseAjena(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()
	adminID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil).Once()
	repo.On("Update", mock.AnythingOfType("*models.Class")).Return(nil)
	repo.On("FindByID", classID).Return(class, nil).Once()
	repo.On("CountActiveReservations", classID).Return(int64(0), nil)

	svc := newClassService(repo)
	req := services.UpdateClassRequest{Titulo: "Editada por admin"}
	_, err := svc.Update(classID, req, adminID, models.RoleAdmin)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdate_FechaPasadaRechazada(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil)

	svc := newClassService(repo)
	pasada := time.Now().Add(-1 * time.Hour)
	req := services.UpdateClassRequest{FechaHora: &pasada}
	_, err := svc.Update(classID, req, profesorID, models.RoleProfesor)

	assert.ErrorIs(t, err, apperrors.ErrClassInvalidDate)
	repo.AssertNotCalled(t, "Update")
}

// --- Cancel ---

func TestCancel_DueñoPuedeCancelar(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil)
	repo.On("Cancel", classID).Return(nil)

	svc := newClassService(repo)
	err := svc.Cancel(classID, profesorID, models.RoleProfesor)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCancel_AlumnoNoPuedeCancelar(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()
	alumnoID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID

	repo.On("FindByID", classID).Return(class, nil)

	svc := newClassService(repo)
	err := svc.Cancel(classID, alumnoID, models.RoleAlumno)

	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	repo.AssertNotCalled(t, "Cancel")
}

func TestCancel_YaCancelada(t *testing.T) {
	repo := &repomocks.ClassRepository{}
	profesorID := uuid.New()
	categoriaID := uuid.New()
	classID := uuid.New()

	class := baseClass(profesorID, categoriaID)
	class.ID = classID
	class.Estado = models.EstadoCancelada

	repo.On("FindByID", classID).Return(class, nil)

	svc := newClassService(repo)
	err := svc.Cancel(classID, profesorID, models.RoleProfesor)

	assert.ErrorIs(t, err, apperrors.ErrClassAlreadyCancelled)
	repo.AssertNotCalled(t, "Cancel")
}
