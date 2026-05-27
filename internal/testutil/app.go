package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/database"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/handlers"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

func migrationsPath() string {
	return "file://" + filepath.ToSlash(filepath.Join(projectRoot(), "migrations"))
}

type appCore struct {
	router   *gin.Engine
	db       *gorm.DB
	cleanup  func()
}

func buildApp(t *testing.T) appCore {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("aluna_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := database.ConnectPostgres(&config.Config{DBDSN: dsn})
	require.NoError(t, err)

	err = database.RunMigrationsFrom(dsn, migrationsPath())
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := &config.Config{
		JWTSecret:                  "test-secret-key-32-chars-long-ok",
		JWTExpirationMinutes:       "15",
		RefreshTokenExpirationDays: "7",
	}

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)

	tokenSvc, err := services.NewTokenService(cfg, rdb)
	require.NoError(t, err)

	authSvc := services.NewAuthService(userRepo, tokenSvc)
	authHandler := handlers.NewAuthHandler(authSvc)

	resetSvc := services.NewResetService(userRepo, resetTokenRepo)
	resetHandler := handlers.NewResetHandler(resetSvc)

	userHandler := handlers.NewUserHandler(userRepo)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	classRepo := repository.NewClassRepository(db)
	classSvc := services.NewClassService(classRepo)

	reservationRepo := repository.NewReservationRepository(db)
	waitlistRepo := repository.NewWaitlistRepository(db)
	reservationSvc := services.NewReservationService(db, reservationRepo, waitlistRepo, classRepo)
	waitlistSvc := services.NewWaitlistService(waitlistRepo, reservationRepo, classRepo)

	classHandler := handlers.NewClassHandler(classSvc, reservationSvc)
	reservationHandler := handlers.NewReservationHandler(reservationSvc, waitlistSvc)

	r := gin.New()
	authMiddleware := middleware.Auth(tokenSvc)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			loginRL := middleware.RateLimit(rdb, 5, time.Minute)
			forgotRL := middleware.RateLimit(rdb, 3, time.Minute)

			auth.POST("/register", authHandler.Register)
			auth.POST("/login", loginRL, authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", forgotRL, resetHandler.ForgotPassword)
			auth.POST("/reset-password", resetHandler.ResetPassword)
		}

		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.List)
			adminOnly := middleware.RequireRole("admin")
			categories.POST("", authMiddleware, adminOnly, categoryHandler.Create)
			categories.PATCH("/:id", authMiddleware, adminOnly, categoryHandler.Update)
			categories.DELETE("/:id", authMiddleware, adminOnly, categoryHandler.Delete)
		}

		classes := api.Group("/classes")
		{
			classes.GET("", classHandler.List)
			classes.GET("/:id", classHandler.Get)
			profesorAdmin := middleware.RequireRole("profesor", "admin")
			classes.POST("", authMiddleware, profesorAdmin, classHandler.Create)
			classes.PATCH("/:id", authMiddleware, classHandler.Update)
			classes.DELETE("/:id", authMiddleware, classHandler.Cancel)
			classes.GET("/:id/students", authMiddleware, profesorAdmin, classHandler.GetStudents)
			classes.POST("/:id/attendance", authMiddleware, profesorAdmin, classHandler.MarkAttendance)
		}

		alumnoOnly := middleware.RequireRole("alumno")
		reservations := api.Group("/reservations", authMiddleware)
		{
			reservations.POST("", alumnoOnly, reservationHandler.Reserve)
			reservations.GET("/me", alumnoOnly, reservationHandler.GetMy)
			reservations.GET("/:id", reservationHandler.Get)
			reservations.DELETE("/:id", reservationHandler.Cancel)
			reservations.POST("/waitlist", alumnoOnly, reservationHandler.JoinWaitlist)
		}

		users := api.Group("/users", authMiddleware)
		{
			users.GET("/me", userHandler.Me)
			users.PATCH("/me", userHandler.UpdateMe)

			adminOnly := middleware.RequireRole("admin")
			users.GET("", adminOnly, userHandler.ListUsers)
			users.GET("/:id", adminOnly, userHandler.GetUser)
			users.DELETE("/:id", adminOnly, userHandler.DeleteUser)
		}
	}

	return appCore{
		router:  r,
		db:      db,
		cleanup: func() { _ = pgContainer.Terminate(ctx) },
	}
}

// SetupApp devuelve el router y cleanup. Compatible con tests de Fases 2 y 3.
func SetupApp(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	core := buildApp(t)
	return core.router, core.cleanup
}

// SetupAppWithDB devuelve el router, la DB y cleanup.
// Usarlo cuando los tests necesitan insertar datos directamente (ej. crear usuarios con roles).
func SetupAppWithDB(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	t.Helper()
	core := buildApp(t)
	return core.router, core.db, core.cleanup
}

// CreateUserWithRole inserta un usuario con el rol dado directamente en DB
// y devuelve su accessToken (válido para autenticar requests en tests).
func CreateUserWithRole(t *testing.T, db *gorm.DB, email, password, nombre string, role models.Role) (userID uuid.UUID, accessToken string) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Nombre:       nombre,
		Role:         role,
	}
	require.NoError(t, db.Create(user).Error)

	// Generamos el token con el mismo cfg que buildApp usa
	cfg := &config.Config{
		JWTSecret:                  "test-secret-key-32-chars-long-ok",
		JWTExpirationMinutes:       "15",
		RefreshTokenExpirationDays: "7",
	}
	// miniredis no es necesario para generar access token
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	tokenSvc, err := services.NewTokenService(cfg, rdb)
	require.NoError(t, err)

	token, err := tokenSvc.GenerateAccessToken(user.ID, role)
	require.NoError(t, err)

	return user.ID, token
}

// RateSetupWithLimits devuelve un router con rate limits configurables para testear el 429.
func RateSetupWithLimits(t *testing.T, loginLimit, forgotLimit int) (*gin.Engine, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("aluna_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := database.ConnectPostgres(&config.Config{DBDSN: dsn})
	require.NoError(t, err)
	require.NoError(t, database.RunMigrationsFrom(dsn, migrationsPath()))

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := &config.Config{
		JWTSecret:                  "test-secret-key-32-chars-long-ok",
		JWTExpirationMinutes:       "15",
		RefreshTokenExpirationDays: "7",
	}

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	tokenSvc, err := services.NewTokenService(cfg, rdb)
	require.NoError(t, err)

	authSvc := services.NewAuthService(userRepo, tokenSvc)
	authHandler := handlers.NewAuthHandler(authSvc)
	resetSvc := services.NewResetService(userRepo, resetTokenRepo)
	resetHandler := handlers.NewResetHandler(resetSvc)

	r := gin.New()
	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/login", middleware.RateLimit(rdb, loginLimit, time.Minute), authHandler.Login)
	auth.POST("/forgot-password", middleware.RateLimit(rdb, forgotLimit, time.Minute), resetHandler.ForgotPassword)
	auth.POST("/register", authHandler.Register)

	cleanup := func() { _ = pgContainer.Terminate(ctx) }
	return r, cleanup
}

func JSON(body string) string {
	return fmt.Sprintf("%s", body)
}
