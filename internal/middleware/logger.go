package middleware

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := newRequestID()
		c.Set("request_id", requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		event := log.Info().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Int("status", c.Writer.Status()).
			Dur("latency", latency)

		if userID, exists := c.Get("user_id"); exists {
			event = event.Interface("user_id", userID)
		}

		event.Msg("request")
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
