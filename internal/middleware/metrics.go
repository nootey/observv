package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/observability"
	"strconv"
	"time"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		observability.HTTPInFlight.Inc()
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		observability.HTTPRequestsTotal.WithLabelValues(method, route, status).Inc()
		observability.HTTPRequestDuration.WithLabelValues(method, route).Observe(duration)
		observability.HTTPInFlight.Dec()
	}
}
