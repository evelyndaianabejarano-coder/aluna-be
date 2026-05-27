package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Categorías ---

func TestListCategorias_Publico(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
}

func TestCrearCategoria_Admin(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin@mail.com", "pass123", "Admin", models.RoleAdmin)

	body := `{"nombre":"Hatha","objetivo":"fuerza","descripcion":"Clases de Hatha Yoga"}`
	w := postWithToken(router, "/api/v1/categories", body, adminToken)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Hatha", resp["nombre"])
	assert.NotEmpty(t, resp["id"])
}

func TestCrearCategoria_AlumnoRecibe403(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "alumno.cat@mail.com", "password123", "Alumno")
	token := reg["accessToken"].(string)

	body := `{"nombre":"X","objetivo":"x","descripcion":"x"}`
	w := postWithToken(router, "/api/v1/categories", body, token)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assertErrorCode(t, w, "AUTH_UNAUTHORIZED")
}

func TestCrearCategoria_ProfesorRecibe403(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, profToken := testutil.CreateUserWithRole(t, db, "prof.cat@mail.com", "pass123", "Profe", models.RoleProfesor)

	body := `{"nombre":"X","objetivo":"x","descripcion":"x"}`
	w := postWithToken(router, "/api/v1/categories", body, profToken)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCrearCategoria_SinToken_401(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	body := `{"nombre":"X","objetivo":"x","descripcion":"x"}`
	w := postWithToken(router, "/api/v1/categories", body, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestActualizarCategoria_Admin(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin2@mail.com", "pass123", "Admin", models.RoleAdmin)

	// Crear categoría primero
	w := postWithToken(router, "/api/v1/categories", `{"nombre":"Yin","objetivo":"relajacion","descripcion":"Yin Yoga"}`, adminToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	id := created["id"].(string)

	// Actualizar
	w2 := patchWithToken(router, "/api/v1/categories/"+id, `{"nombre":"Yin Updated","objetivo":"calma","descripcion":"Actualizada"}`, adminToken)
	assert.Equal(t, http.StatusOK, w2.Code)
	var updated map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &updated))
	assert.Equal(t, "Yin Updated", updated["nombre"])
}

func TestEliminarCategoria_Admin_204(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin3@mail.com", "pass123", "Admin", models.RoleAdmin)

	w := postWithToken(router, "/api/v1/categories", `{"nombre":"Para borrar","objetivo":"x","descripcion":"x"}`, adminToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	id := created["id"].(string)

	w2 := deleteWithToken(router, "/api/v1/categories/"+id, adminToken)
	assert.Equal(t, http.StatusNoContent, w2.Code)

	// Verificar que ya no existe — la lista debe estar vacía (solo había una)
	w3 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	router.ServeHTTP(w3, req)
	var list map[string]any
	require.NoError(t, json.Unmarshal(w3.Body.Bytes(), &list))
	data := list["data"].([]any)
	assert.Empty(t, data)
}

func TestCrearCategoria_BodyInvalido_422(t *testing.T) {
	router, db, cleanup := testutil.SetupAppWithDB(t)
	defer cleanup()

	_, adminToken := testutil.CreateUserWithRole(t, db, "admin4@mail.com", "pass123", "Admin", models.RoleAdmin)

	w := postWithToken(router, "/api/v1/categories", `{"nombre":""}`, adminToken)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
