package main

import (
	"time"

	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/database"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/handlers"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/logger"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error cargando configuración")
	}

	logger.Setup(cfg)

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error conectando a PostgreSQL")
	}
	log.Info().Msg("PostgreSQL conectado")

	if err := database.RunMigrations(cfg.DBDSN); err != nil {
		log.Fatal().Err(err).Msg("error ejecutando migraciones")
	}

	rdb, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error conectando a Redis")
	}
	log.Info().Msg("Redis conectado")

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)

	tokenSvc, err := services.NewTokenService(cfg, rdb)
	if err != nil {
		log.Fatal().Err(err).Msg("error inicializando token service")
	}

	authSvc := services.NewAuthService(userRepo, tokenSvc)
	authHandler := handlers.NewAuthHandler(authSvc)

	resetSvc := services.NewResetService(userRepo, resetTokenRepo)
	resetHandler := handlers.NewResetHandler(resetSvc)

	userHandler := handlers.NewUserHandler(userRepo)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	classRepo := repository.NewClassRepository(db)
	classSvc := services.NewClassService(classRepo)
	classHandler := handlers.NewClassHandler(classSvc)

	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(middleware.HTTPMetrics())

	r.GET("/health", handlers.Liveness)
	r.GET("/health/ready", handlers.Readiness(db, rdb))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authMiddleware := middleware.Auth(tokenSvc)
	loginRateLimit := middleware.RateLimit(rdb, 5, time.Minute)
	forgotRateLimit := middleware.RateLimit(rdb, 3, time.Minute)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", loginRateLimit, authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", forgotRateLimit, resetHandler.ForgotPassword)
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

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal().Err(err).Msg("error iniciando servidor")
	}
}
