package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsExposer exposes Prometheus metrics via HTTP endpoint
type MetricsExposer struct {
	logger         *zap.Logger
	registry       *prometheus.Registry
	router         *gin.Engine
	metricsHandler http.Handler
}

// NewMetricsExposer creates a new metrics exposer
func NewMetricsExposer(logger *zap.Logger, registry *prometheus.Registry) *MetricsExposer {
	return &MetricsExposer{
		logger:         logger,
		registry:       registry,
		metricsHandler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}
}

// RegisterRoutes registers the metrics endpoint with the provided router
func (e *MetricsExposer) WithMetricsEndpoint(router *gin.Engine) {
	router.GET("/metrics", func(c *gin.Context) {
		// Set the Prometheus content type header
		c.Header("Content-Type", "text/plain; version=0.0.4")
		e.metricsHandler.ServeHTTP(c.Writer, c.Request)
	})

	e.router = router
	e.logger.Info("Metrics endpoint registered at /metrics")
}
