package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Readiness(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}
		healthy := true

		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			checks["postgres"] = "unavailable"
			healthy = false
		} else {
			checks["postgres"] = "ok"
		}

		if err := rdb.Ping(context.Background()).Err(); err != nil {
			checks["redis"] = "unavailable"
			healthy = false
		} else {
			checks["redis"] = "ok"
		}

		if !healthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "checks": checks})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "checks": checks})
	}
}
