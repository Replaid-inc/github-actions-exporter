package server

import (
	"context"
	"gh-actions-exporter/internal/handlers"
	"gh-actions-exporter/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServer(port string) {

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create a logger with debug level enabled
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	logger.Debug("Debug logging enabled")

	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	exposer := metrics.NewMetricsExposer(logger, registry)

	r := gin.New()
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)
	r.SetTrustedProxies(nil)

	// Use middleware to inject dependencies
	r.POST("/webhook", func(c *gin.Context) {
		handlers.WebhookHandler(c, processor, logger)
	})

	r.GET("/health", handleHealth)

	exposer.WithMetricsEndpoint(r)

	// Start HTTP server
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Shutdown gracefully with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
