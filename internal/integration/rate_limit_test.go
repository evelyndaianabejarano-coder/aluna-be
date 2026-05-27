package integration_test

import (
	"net/http"
	"testing"

	"github.com/evelyndaianabejarano-coder/aluna-be/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_Login_429(t *testing.T) {
	// Limit: 2 intentos por minuto para forzar 429 rápido
	router, cleanup := testutil.RateSetupWithLimits(t, 2, 3)
	defer cleanup()

	body := `{"email":"x@mail.com","password":"pass"}`

	// Primeros 2 — pasan (401 por credenciales inválidas, no 429)
	for i := 0; i < 2; i++ {
		w := post(router, "/api/v1/auth/login", body)
		assert.NotEqual(t, http.StatusTooManyRequests, w.Code, "request %d no debería ser 429", i+1)
	}

	// Tercer intento — debe ser 429
	w := post(router, "/api/v1/auth/login", body)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assertErrorCode(t, w, "RATE_LIMIT_EXCEEDED")
	assert.NotEmpty(t, w.Header().Get("Retry-After"), "debe incluir el header Retry-After")
}

func TestRateLimit_ForgotPassword_429(t *testing.T) {
	router, cleanup := testutil.RateSetupWithLimits(t, 5, 2)
	defer cleanup()

	// Primero registramos un usuario
	post(router, "/api/v1/auth/register", `{"email":"fp2@mail.com","password":"password123","nombre":"FP2"}`)

	body := `{"email":"fp2@mail.com"}`

	// Primeros 2 — pasan (204)
	for i := 0; i < 2; i++ {
		w := post(router, "/api/v1/auth/forgot-password", body)
		assert.Equal(t, http.StatusNoContent, w.Code, "request %d debería ser 204", i+1)
	}

	// Tercer intento — debe ser 429
	w := post(router, "/api/v1/auth/forgot-password", body)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assertErrorCode(t, w, "RATE_LIMIT_EXCEEDED")
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}
