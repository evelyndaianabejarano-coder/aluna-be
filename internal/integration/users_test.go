package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMe_SinToken_401(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := getWithToken(router, "/api/v1/users/me", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMe_ConToken_200(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "me@mail.com", "password123", "Me")
	accessToken := reg["accessToken"].(string)

	w := getWithToken(router, "/api/v1/users/me", accessToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "me@mail.com", resp["email"])
}

func TestListUsers_AlumnoRecibe403(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "alumno@mail.com", "password123", "Alumno")
	token := reg["accessToken"].(string)

	w := getWithToken(router, "/api/v1/users", token)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assertErrorCode(t, w, "AUTH_UNAUTHORIZED")
}

func TestGetUser_AlumnoRecibe403(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "a2@mail.com", "password123", "A2")
	token := reg["accessToken"].(string)
	userID := reg["user"].(map[string]any)["id"].(string)

	w := getWithToken(router, "/api/v1/users/"+userID, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteUser_AlumnoRecibe403(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "a3@mail.com", "password123", "A3")
	token := reg["accessToken"].(string)
	userID := reg["user"].(map[string]any)["id"].(string)

	w := deleteWithToken(router, "/api/v1/users/"+userID, token)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestFlujoCompleto_RegisterLoginProtected(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	// 1. Register
	reg := registerUser(t, router, "flujo@mail.com", "password123", "Flujo")
	assert.NotEmpty(t, reg["accessToken"])

	// 2. Login
	login := loginUser(t, router, "flujo@mail.com", "password123")
	accessToken := login["accessToken"].(string)

	// 3. Acceder a ruta protegida
	w := getWithToken(router, "/api/v1/users/me", accessToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var me map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &me))
	assert.Equal(t, "flujo@mail.com", me["email"])
}
