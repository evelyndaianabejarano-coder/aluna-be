package logger

import (
	"os"
	"time"

	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Setup(cfg *config.Config) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "aluna").
		Str("env", cfg.Env).
		Logger()
}
