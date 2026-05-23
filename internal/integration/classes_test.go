package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createCategoria es un helper que crea una categoría y devuelve su ID.
func createCategoria(t *testing.T, router http.Handler, adminToken string) string {
	t.Helper()
	w := postWithToken(router, "/api/v1/categories",
		`{"nombre":"Hatha","objetivo":"fuerza","descripcion":"Hatha Yoga"}`, adminToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp["id"].(string)
}

// createClaseBody construye el JSON para crear una clase.
func createClaseBody(categoriaID string) string {
	fecha := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	return fmt.Sprintf(`{
		"titulo": "Hatha matutino",
		"descripcion": "Clase energizante",
		"categoria_id": %q,
		"fecha_hora": %q,
		"duracion": 60,
		"cupo_maximo": 10,
		"modalidad": "presencial"
	}`, categoriaID, fecha)
}

// --- Clases ---

func TestListClases_Publico(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(0), resp["total"])
}

func TestCrearClase_ProfesorExitoso(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.c@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.c@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)
	body := createClaseBody(catID)

	w := postWithToken(router, "/api/v1/classes", body, profToken)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Hatha matutino", resp["titulo"])
	assert.Equal(t, "presencial", resp["modalidad"])
	assert.NotEmpty(t, resp["id"])
}

func TestCrearClase_AlumnoRecibe403(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.ca@mail.com", "pass123", "Admin", models.RoleAdmin)
	catID := createCategoria(t, router, adminToken)

	reg := registerUser(t, router, "alumno.cl@mail.com", "password123", "Alumno")
	alumnoToken := reg["accessToken"].(string)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), alumnoToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetClase_CuposDisponibles(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.g@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.g@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)
	cupoMaximo := created["cupo_maximo"].(float64)

	// GET detalle — en Fase 3 no hay reservas, cupos_disponibles == cupo_maximo
	w2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes/"+classID, nil)
	router.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)
	var detail map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &detail))
	assert.Equal(t, cupoMaximo, detail["cupos_disponibles"])
	assert.NotEmpty(t, detail["categoria_nombre"])
	assert.NotEmpty(t, detail["profesor_nombre"])
}

func TestListClases_FiltroModalidad(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.f@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.f@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)
	fecha := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)

	// Crear clase presencial
	postWithToken(router, "/api/v1/classes", fmt.Sprintf(`{
		"titulo":"Presencial","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":5,"modalidad":"presencial"
	}`, catID, fecha), profToken)

	// Crear clase online
	postWithToken(router, "/api/v1/classes", fmt.Sprintf(`{
		"titulo":"Online","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":5,"modalidad":"online"
	}`, catID, fecha), profToken)

	// Filtrar solo online
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes?modalidad=online", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Online", data[0].(map[string]any)["titulo"])
}

func TestListClases_FiltroFecha(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.fd@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.fd@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	hoy := time.Now().UTC()
	manana := hoy.Add(24 * time.Hour)
	pasadoManana := hoy.Add(48 * time.Hour)

	postWithToken(router, "/api/v1/classes", fmt.Sprintf(`{
		"titulo":"Manana","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":5,"modalidad":"presencial"
	}`, catID, manana.Format(time.RFC3339)), profToken)

	postWithToken(router, "/api/v1/classes", fmt.Sprintf(`{
		"titulo":"Pasado","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":5,"modalidad":"presencial"
	}`, catID, pasadoManana.Format(time.RFC3339)), profToken)

	// Filtrar por fecha de mañana
	fechaFiltro := manana.Format("2006-01-02")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes?fecha="+fechaFiltro, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Manana", data[0].(map[string]any)["titulo"])
}

func TestCancelarClase_Dueno(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.can@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.can@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	w2 := deleteWithToken(router, "/api/v1/classes/"+classID, profToken)
	assert.Equal(t, http.StatusNoContent, w2.Code)
}

func TestCancelarClase_AlumnoRecibe403(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.can2@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.can2@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	reg := registerUser(t, router, "alumno.can@mail.com", "password123", "Alumno")
	alumnoToken := reg["accessToken"].(string)

	w2 := deleteWithToken(router, "/api/v1/classes/"+classID, alumnoToken)
	assert.Equal(t, http.StatusForbidden, w2.Code)
	assertErrorCode(t, w2, "AUTH_UNAUTHORIZED")
}

func TestCancelarClase_OtroProfesorRecibe403(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.can3@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, prof1Token := testutil.CreateUserWithRole(t, db, "prof1@mail.com", "pass123", "Profe1", models.RoleProfesor)
	_, prof2Token := testutil.CreateUserWithRole(t, db, "prof2@mail.com", "pass123", "Profe2", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), prof1Token)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	// Otro profesor intenta cancelar
	w2 := deleteWithToken(router, "/api/v1/classes/"+classID, prof2Token)
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

func TestActualizarClase_Dueno(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.upd@mail.com", "password123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.upd@mail.com", "password123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	w2 := patchWithToken(router, "/api/v1/classes/"+classID, `{"titulo":"Hatha actualizado","duracion":90}`, profToken)
	assert.Equal(t, http.StatusOK, w2.Code)
	var updated map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &updated))
	assert.Equal(t, "Hatha actualizado", updated["titulo"])
	assert.Equal(t, float64(90), updated["duracion"])
}

func TestActualizarClase_AlumnoRecibe403(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.upd2@mail.com", "password123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.upd2@mail.com", "password123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	reg := registerUser(t, router, "alumno.upd@mail.com", "password123", "Alumno")
	alumnoToken := reg["accessToken"].(string)

	w2 := patchWithToken(router, "/api/v1/classes/"+classID, `{"titulo":"intento"}`, alumnoToken)
	assert.Equal(t, http.StatusForbidden, w2.Code)
	assertErrorCode(t, w2, "AUTH_UNAUTHORIZED")
}

func TestActualizarClase_AdminPuedeEditarClaseAjena(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.upd3@mail.com", "password123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.upd3@mail.com", "password123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)

	w := postWithToken(router, "/api/v1/classes", createClaseBody(catID), profToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	classID := created["id"].(string)

	w2 := patchWithToken(router, "/api/v1/classes/"+classID, `{"titulo":"Editada por admin"}`, adminToken)
	assert.Equal(t, http.StatusOK, w2.Code)
	var updated map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &updated))
	assert.Equal(t, "Editada por admin", updated["titulo"])
}

func TestGetClase_NoExiste_404(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes/00000000-0000-0000-0000-000000000000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorCode(t, w, "CLASS_NOT_FOUND")
}

func TestCrearClase_FechaPasada_422(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.fp@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.fp@mail.com", "pass123", "Profe", models.RoleProfesor)

	catID := createCategoria(t, router, adminToken)
	fechaPasada := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)

	body := fmt.Sprintf(`{
		"titulo":"Clase pasada","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":5,"modalidad":"presencial"
	}`, catID, fechaPasada)

	w := postWithToken(router, "/api/v1/classes", body, profToken)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assertErrorCode(t, w, "CLASS_INVALID_DATE")
}
