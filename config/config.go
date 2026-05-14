package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	Env      string
	DBDSN    string
	RedisURL string

	JWTSecret                  string
	JWTExpirationMinutes       string
	RefreshTokenExpirationDays string

	LogLevel string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:                    getEnv("APP_PORT", "8080"),
		Env:                        getEnv("ENV", "development"),
		DBDSN:                      os.Getenv("DB_DSN"),
		RedisURL:                   os.Getenv("REDIS_URL"),
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		JWTExpirationMinutes:       getEnv("JWT_EXPIRATION_MINUTES", "15"),
		RefreshTokenExpirationDays: getEnv("REFRESH_TOKEN_EXPIRATION_DAYS", "7"),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"DB_DSN":     c.DBDSN,
		"REDIS_URL":  c.RedisURL,
		"JWT_SECRET": c.JWTSecret,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("variable de entorno requerida no definida: %s", key)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
