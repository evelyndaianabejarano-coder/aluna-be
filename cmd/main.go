package main

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/database"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/handlers"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/logger"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/gin-gonic/gin"
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

	rdb, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error conectando a Redis")
	}
	log.Info().Msg("Redis conectado")

	r := gin.New()
	r.Use(middleware.RequestLogger())

	r.GET("/health", handlers.Liveness)
	r.GET("/health/ready", handlers.Readiness(db, rdb))

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal().Err(err).Msg("error iniciando servidor")
	}
}
