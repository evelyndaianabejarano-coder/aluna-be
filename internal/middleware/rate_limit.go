package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit returns a sliding-window rate limiter keyed by client IP.
// limit: max requests allowed within the window duration.
func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		// Use the route path as part of the key so limits are per-endpoint
		key := fmt.Sprintf("rate_limit:%s:%s", c.FullPath(), ip)

		now := time.Now()
		windowStart := now.Add(-window)

		ctx := context.Background()

		pipe := rdb.Pipeline()
		// Remove entries outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixMicro(), 10))
		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMicro()), Member: now.UnixMicro()})
		// Count requests in the window
		countCmd := pipe.ZCard(ctx, key)
		// Reset TTL
		pipe.Expire(ctx, key, window)

		if _, err := pipe.Exec(ctx); err != nil {
			// Si Redis falla, dejamos pasar para no bloquear el servicio
			c.Next()
			return
		}

		count := countCmd.Val()
		if count > int64(limit) {
			retryAfter := int(window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			apperrors.Respond(c, apperrors.ErrRateLimitExceeded)
			c.Abort()
			return
		}

		c.Next()
	}
}
