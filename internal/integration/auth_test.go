package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Exitoso(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := post(router, "/api/v1/auth/register", `{"email":"ana@mail.com","password":"password123","nombre":"Ana"}`)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ana@mail.com", resp["user"].(map[string]any)["email"])
	assert.NotEmpty(t, resp["accessToken"])
	assert.NotEmpty(t, resp["refreshToken"])
}

func TestRegister_EmailDuplicado(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	post(router, "/api/v1/auth/register", `{"email":"dup@mail.com","password":"password123","nombre":"Dup"}`)
	w := post(router, "/api/v1/auth/register", `{"email":"dup@mail.com","password":"password123","nombre":"Dup"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
	assertErrorCode(t, w, "AUTH_EMAIL_ALREADY_EXISTS")
}

func TestRegister_BodyInvalido(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := post(router, "/api/v1/auth/register", `{"email":"no-es-email","password":"pass","nombre":""}`)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestLogin_Exitoso(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	registerUser(t, router, "login@mail.com", "password123", "Login")
	w := post(router, "/api/v1/auth/login", `{"email":"login@mail.com","password":"password123"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["accessToken"])
}

func TestLogin_CredencialesInvalidas(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	registerUser(t, router, "cred@mail.com", "correcta123", "Cred")
	w := post(router, "/api/v1/auth/login", `{"email":"cred@mail.com","password":"incorrecta"}`)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrorCode(t, w, "AUTH_INVALID_CREDENTIALS")
}

func TestRefresh_Exitoso(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "refresh@mail.com", "password123", "Ref")
	refreshToken := reg["refreshToken"].(string)

	w := post(router, "/api/v1/auth/refresh", `{"refreshToken":"`+refreshToken+`"}`)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["accessToken"])
	assert.NotEmpty(t, resp["refreshToken"])
	assert.NotEqual(t, refreshToken, resp["refreshToken"], "el refresh token debe rotarse")
}

func TestRefresh_TokenInvalido(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	w := post(router, "/api/v1/auth/refresh", `{"refreshToken":"token-falso"}`)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrorCode(t, w, "AUTH_INVALID_REFRESH_TOKEN")
}

func TestLogout_Exitoso(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	reg := registerUser(t, router, "logout@mail.com", "password123", "Logout")
	refreshToken := reg["refreshToken"].(string)

	w := post(router, "/api/v1/auth/logout", `{"refreshToken":"`+refreshToken+`"}`)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// El token ya no debe ser válido
	w2 := post(router, "/api/v1/auth/refresh", `{"refreshToken":"`+refreshToken+`"}`)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestForgotPassword_SiempreDev204(t *testing.T) {
	router, cleanup := testutil.SetupApp(t)
	defer cleanup()

	// Email existente — 204
	registerUser(t, router, "fp@mail.com", "password123", "FP")
	w1 := post(router, "/api/v1/auth/forgot-password", `{"email":"fp@mail.com"}`)
	assert.Equal(t, http.StatusNoContent, w1.Code)

	// Email inexistente — también 204, no revela si existe
	w2 := post(router, "/api/v1/auth/forgot-password", `{"email":"noexiste@mail.com"}`)
	assert.Equal(t, http.StatusNoContent, w2.Code)
}

