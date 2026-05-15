package main

import (
	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/database"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/handlers"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/logger"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
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

	rdb, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error conectando a Redis")
	}
	log.Info().Msg("Redis conectado")

	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(middleware.HTTPMetrics())

	r.GET("/health", handlers.Liveness)
	r.GET("/health/ready", handlers.Readiness(db, rdb))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal().Err(err).Msg("error iniciando servidor")
	}
}
