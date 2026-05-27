package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requests HTTP por método, path y status code.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latencia de requests HTTP por método y path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		elapsed := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(elapsed)
	}
}
