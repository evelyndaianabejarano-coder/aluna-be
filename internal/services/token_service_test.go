package services_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTokenService(t *testing.T) (services.TokenService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := &config.Config{
		JWTSecret:                  "test-secret-32-chars-long-enough",
		JWTExpirationMinutes:       "15",
		RefreshTokenExpirationDays: "7",
	}
	svc, err := services.NewTokenService(cfg, rdb)
	require.NoError(t, err)
	return svc, mr
}

func TestGenerateAccessToken_ClaimsCorrectos(t *testing.T) {
	svc, _ := newTokenService(t)
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(userID, models.RoleAlumno)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, models.RoleAlumno, claims.Role)
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), claims.ExpiresAt.Time, 5*time.Second)
}

func TestGenerateRefreshToken_GuardadoEnRedis(t *testing.T) {
	svc, mr := newTokenService(t)
	userID := uuid.New()

	token, err := svc.GenerateRefreshToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)

	// Verifica que el JTI existe en Redis
	key := "refresh:" + claims.JTI
	assert.True(t, mr.Exists(key), "el refresh token debe estar en Redis")

	// Verifica que tiene TTL (no persiste para siempre)
	ttl := mr.TTL(key)
	assert.Greater(t, ttl, time.Duration(0))
}

func TestValidateToken_TokenExpirado(t *testing.T) {
	svc, mr := newTokenService(t)
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(userID, models.RoleAlumno)
	require.NoError(t, err)

	// Avanza el tiempo en miniredis y el reloj del test
	mr.FastForward(20 * time.Minute)

	// El token tiene exp de 15 minutos, jwt lo valida con el reloj real
	// Para forzar expiración necesitamos un token con TTL mínimo — usamos config ad-hoc
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := &config.Config{
		JWTSecret:                  "test-secret-32-chars-long-enough",
		JWTExpirationMinutes:       "0", // expira inmediatamente
		RefreshTokenExpirationDays: "7",
	}
	shortSvc, err := services.NewTokenService(cfg, rdb)
	require.NoError(t, err)

	expiredToken, err := shortSvc.GenerateAccessToken(userID, models.RoleAlumno)
	require.NoError(t, err)

	// Espera 1 segundo para que expire (exp=now+0min ≈ ya expirado)
	time.Sleep(100 * time.Millisecond)

	_, err = svc.ValidateToken(expiredToken)
	assert.Error(t, err, "un token expirado debe devolver error")

	_ = token // token con 15min usado solo para verificar generación arriba
}

func TestValidateToken_FirmaIncorrecta(t *testing.T) {
	svc, _ := newTokenService(t)

	// Token firmado con otro secret
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"}) // no se usa en este test
	otherCfg := &config.Config{
		JWTSecret:                  "otro-secret-completamente-distinto",
		JWTExpirationMinutes:       "15",
		RefreshTokenExpirationDays: "7",
	}
	otherSvc, err := services.NewTokenService(otherCfg, rdb)
	require.NoError(t, err)

	token, err := otherSvc.GenerateAccessToken(uuid.New(), models.RoleAlumno)
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err, "un token con firma incorrecta debe rechazarse")
}

func TestValidateRefreshToken_Revocado(t *testing.T) {
	svc, _ := newTokenService(t)
	userID := uuid.New()

	token, err := svc.GenerateRefreshToken(userID)
	require.NoError(t, err)

	err = svc.RevokeRefreshToken(token)
	require.NoError(t, err)

	_, err = svc.ValidateRefreshToken(token)
	assert.Error(t, err, "un refresh token revocado debe rechazarse")
}

func TestRevokeRefreshToken_EliminaDeRedis(t *testing.T) {
	svc, mr := newTokenService(t)
	userID := uuid.New()

	token, err := svc.GenerateRefreshToken(userID)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)

	key := "refresh:" + claims.JTI
	assert.True(t, mr.Exists(key))

	err = svc.RevokeRefreshToken(token)
	require.NoError(t, err)

	assert.False(t, mr.Exists(key), "el refresh token debe eliminarse de Redis al revocar")
}
