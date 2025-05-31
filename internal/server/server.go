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

func StartServer(port string, webhookSecret string) {

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Get log level from environment variable, default to info
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Parse log level
	var zapLevel zap.AtomicLevel
	switch logLevel {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Create a logger with configurable level
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.Level = zapLevel
	logger, err := loggerConfig.Build()
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	logger.Info("Logger initialized", zap.String("level", logLevel))

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
		handlers.WebhookHandler(c, processor, logger, webhookSecret)
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
