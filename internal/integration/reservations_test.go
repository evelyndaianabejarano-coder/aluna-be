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

// --- Helpers locales ---

func createClaseConCupo(t *testing.T, router http.Handler, token, catID string, cupo int) string {
	t.Helper()
	fecha := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	body := fmt.Sprintf(`{
		"titulo":"Clase test","descripcion":"desc","categoria_id":%q,
		"fecha_hora":%q,"duracion":60,"cupo_maximo":%d,"modalidad":"presencial"
	}`, catID, fecha, cupo)
	w := postWithToken(router, "/api/v1/classes", body, token)
	require.Equal(t, http.StatusCreated, w.Code, "crear clase: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp["id"].(string)
}

func reservarClase(t *testing.T, router http.Handler, token, claseID string) string {
	t.Helper()
	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations", body, token)
	require.Equal(t, http.StatusCreated, w.Code, "reservar clase: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp["id"].(string)
}

func getCuposDisponibles(t *testing.T, router http.Handler, claseID string) float64 {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes/"+claseID, nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp["cupos_disponibles"].(float64)
}

// --- Tests de Reservas ---

func TestReservar_AlumnoExitoso_201Pendiente(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.r1@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.r1@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	reg := registerUser(t, router, "alumno.r1@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)

	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations", body, alumnoToken)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "pendiente", resp["estado"])
	assert.NotEmpty(t, resp["id"])
}

func TestReservar_ReduceCuposDisponibles(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.r2@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.r2@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 5)

	cuposAntes := getCuposDisponibles(t, router, claseID)
	assert.Equal(t, float64(5), cuposAntes)

	reg := registerUser(t, router, "alumno.r2@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)
	reservarClase(t, router, alumnoToken, claseID)

	cuposDespues := getCuposDisponibles(t, router, claseID)
	assert.Equal(t, float64(4), cuposDespues)
}

func TestReservar_MismaClase_409Conflict(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.r3@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.r3@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	reg := registerUser(t, router, "alumno.r3@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)
	reservarClase(t, router, alumnoToken, claseID)

	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations", body, alumnoToken)

	assert.Equal(t, http.StatusConflict, w.Code)
	assertErrorCode(t, w, "RESERVATION_CONFLICT")
}

func TestReservar_ClaseLlena_409ClassFull(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.r4@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.r4@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 1)

	// Primer alumno llena la clase
	reg1 := registerUser(t, router, "alumno1.r4@mail.com", "pass123", "Alumno1")
	reservarClase(t, router, reg1["accessToken"].(string), claseID)

	// Segundo alumno intenta reservar → clase llena
	reg2 := registerUser(t, router, "alumno2.r4@mail.com", "pass123", "Alumno2")
	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations", body, reg2["accessToken"].(string))

	assert.Equal(t, http.StatusConflict, w.Code)
	assertErrorCode(t, w, "RESERVATION_CLASS_FULL")
}

func TestCancelarReserva_CupoLiberado(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.r5@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.r5@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 2)

	reg := registerUser(t, router, "alumno.r5@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)
	reservaID := reservarClase(t, router, alumnoToken, claseID)

	cuposAntes := getCuposDisponibles(t, router, claseID)
	assert.Equal(t, float64(1), cuposAntes)

	w := deleteWithToken(router, "/api/v1/reservations/"+reservaID, alumnoToken)
	assert.Equal(t, http.StatusNoContent, w.Code)

	cuposDespues := getCuposDisponibles(t, router, claseID)
	assert.Equal(t, float64(2), cuposDespues)
}

func TestWaitlist_UnirseExitoso(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.w1@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.w1@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 1)

	// Llenar la clase
	reg1 := registerUser(t, router, "alumno1.w1@mail.com", "pass123", "Alumno1")
	reservarClase(t, router, reg1["accessToken"].(string), claseID)

	// Segundo alumno se une a la lista de espera
	reg2 := registerUser(t, router, "alumno2.w1@mail.com", "pass123", "Alumno2")
	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations/waitlist", body, reg2["accessToken"].(string))

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestWaitlist_ClaseConCupos_400(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.w2@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.w2@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	reg := registerUser(t, router, "alumno.w2@mail.com", "pass123", "Alumno")
	body := fmt.Sprintf(`{"clase_id":%q}`, claseID)
	w := postWithToken(router, "/api/v1/reservations/waitlist", body, reg["accessToken"].(string))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorCode(t, w, "CLASS_NOT_FULL")
}

func TestCancelarReserva_ConWaitlist_PromocionAutomatica(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.w3@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.w3@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 1)

	// Alumno 1 reserva (llena la clase)
	reg1 := registerUser(t, router, "alumno1.w3@mail.com", "pass123", "Alumno1")
	token1 := reg1["accessToken"].(string)
	reservaID := reservarClase(t, router, token1, claseID)

	// Alumno 2 entra a la lista de espera
	reg2 := registerUser(t, router, "alumno2.w3@mail.com", "pass123", "Alumno2")
	token2 := reg2["accessToken"].(string)
	postWithToken(router, "/api/v1/reservations/waitlist", fmt.Sprintf(`{"clase_id":%q}`, claseID), token2)

	// Cupos disponibles = 0
	assert.Equal(t, float64(0), getCuposDisponibles(t, router, claseID))

	// Alumno 1 cancela → debe promover a alumno 2 automáticamente
	w := deleteWithToken(router, "/api/v1/reservations/"+reservaID, token1)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Cupos disponibles sigue en 0 porque alumno 2 fue promovido
	assert.Equal(t, float64(0), getCuposDisponibles(t, router, claseID))

	// Alumno 2 intenta unirse a la waitlist → debería fallar (ya tiene reserva)
	w2 := postWithToken(router, "/api/v1/reservations", fmt.Sprintf(`{"clase_id":%q}`, claseID), token2)
	assert.Equal(t, http.StatusConflict, w2.Code)
	assertErrorCode(t, w2, "RESERVATION_CONFLICT")
}

func TestGetReservasMias_Paginado(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.gm@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.gm@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)

	reg := registerUser(t, router, "alumno.gm@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)

	// Crear y reservar 2 clases
	for range 2 {
		claseID := createClaseConCupo(t, router, profToken, catID, 10)
		reservarClase(t, router, alumnoToken, claseID)
	}

	w := getWithToken(router, "/api/v1/reservations/me?page=1&limit=10", alumnoToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]any)
	assert.Len(t, data, 2)
	assert.Equal(t, float64(2), resp["total"])
}

func TestGetReserva_DuenoVsAjeno(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.gr@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.gr@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	reg1 := registerUser(t, router, "alumno1.gr@mail.com", "pass123", "Alumno1")
	token1 := reg1["accessToken"].(string)
	reservaID := reservarClase(t, router, token1, claseID)

	// Dueño puede ver su reserva
	w := getWithToken(router, "/api/v1/reservations/"+reservaID, token1)
	assert.Equal(t, http.StatusOK, w.Code)

	// Alumno ajeno no puede ver la reserva
	reg2 := registerUser(t, router, "alumno2.gr@mail.com", "pass123", "Alumno2")
	token2 := reg2["accessToken"].(string)
	w2 := getWithToken(router, "/api/v1/reservations/"+reservaID, token2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
	assertErrorCode(t, w2, "AUTH_UNAUTHORIZED")
}

func TestGetStudents_ProfesorPuedeVer(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.gs@mail.com", "pass123", "Admin", models.RoleAdmin)
	_, profToken := testutil.CreateUserWithRole(t, db, "prof.gs@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	// Dos alumnos reservan
	reg1 := registerUser(t, router, "alumno1.gs@mail.com", "pass123", "Alumno1")
	reg2 := registerUser(t, router, "alumno2.gs@mail.com", "pass123", "Alumno2")
	reservarClase(t, router, reg1["accessToken"].(string), claseID)
	reservarClase(t, router, reg2["accessToken"].(string), claseID)

	w := getWithToken(router, "/api/v1/classes/"+claseID+"/students", profToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]any)
	assert.Len(t, data, 2)

	// Alumno no tiene permiso
	reg3 := registerUser(t, router, "alumno3.gs@mail.com", "pass123", "Alumno3")
	w2 := getWithToken(router, "/api/v1/classes/"+claseID+"/students", reg3["accessToken"].(string))
	assert.Equal(t, http.StatusForbidden, w2.Code)
}

func TestMarkAttendance_ProfesorRegistraAsistencia(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin.ma@mail.com", "pass123", "Admin", models.RoleAdmin)
	profID, profToken := testutil.CreateUserWithRole(t, db, "prof.ma@mail.com", "pass123", "Profe", models.RoleProfesor)
	catID := createCategoria(t, router, adminToken)
	claseID := createClaseConCupo(t, router, profToken, catID, 10)

	reg := registerUser(t, router, "alumno.ma@mail.com", "pass123", "Alumno")
	alumnoToken := reg["accessToken"].(string)
	reservarClase(t, router, alumnoToken, claseID)

	// Obtener el alumno_id desde GET /me
	wMe := getWithToken(router, "/api/v1/users/me", alumnoToken)
	require.Equal(t, http.StatusOK, wMe.Code)
	var meResp map[string]any
	require.NoError(t, json.Unmarshal(wMe.Body.Bytes(), &meResp))
	alumnoID := meResp["id"].(string)
	_ = profID

	body := fmt.Sprintf(`{"alumnos":[{"id":%q,"presente":true}]}`, alumnoID)
	w := postWithToken(router, "/api/v1/classes/"+claseID+"/attendance", body, profToken)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verificar que la reserva tiene asistio=true
	wRes := getWithToken(router, "/api/v1/reservations/me", alumnoToken)
	require.Equal(t, http.StatusOK, wRes.Code)
	var resResp map[string]any
	require.NoError(t, json.Unmarshal(wRes.Body.Bytes(), &resResp))
	data := resResp["data"].([]any)
	require.Len(t, data, 1)
	reserva := data[0].(map[string]any)
	assert.Equal(t, true, reserva["asistio"])
}
