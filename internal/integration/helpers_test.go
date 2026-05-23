package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- HTTP helpers ---

func post(router http.Handler, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func postWithToken(router http.Handler, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

func getWithToken(router http.Handler, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

func patchWithToken(router http.Handler, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

func deleteWithToken(router http.Handler, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

// --- Assertion helpers ---

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, code string) {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj, ok := resp["error"].(map[string]any)
	require.True(t, ok, "respuesta debe tener campo 'error'")
	assert.Equal(t, code, errObj["code"])
}

// --- Auth helpers reutilizables ---

func registerUser(t *testing.T, router http.Handler, email, password, nombre string) map[string]any {
	t.Helper()
	w := post(router, "/api/v1/auth/register", `{"email":"`+email+`","password":"`+password+`","nombre":"`+nombre+`"}`)
	require.Equal(t, http.StatusCreated, w.Code, "register falló: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func loginUser(t *testing.T, router http.Handler, email, password string) map[string]any {
	t.Helper()
	w := post(router, "/api/v1/auth/login", `{"email":"`+email+`","password":"`+password+`"}`)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}
