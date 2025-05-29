package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gh-actions-exporter/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWebhookHandler_UnsupportedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger)
	})

	payload := `{"action":"opened","issue":{"number":1}}`
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "issues")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ignored"`)
	assert.Contains(t, w.Body.String(), `"event":"issues"`)
}

func TestWebhookHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger)
	})

	payload := `{
		"action":"completed",
		"workflow_run":{
			"id":789,
			"name":"CI",
			"status":"completed",
			"conclusion":"success",
			"run_started_at":"2023-01-01T12:00:00Z",
			"updated_at":"2023-01-01T12:10:00Z",
			"head_branch":"main",
			"event":"push",
			"head_sha":"acb5820ced9479c074f688cc328bf03f341a511d"
		},
		"repository":{
			"full_name":"owner/repo"
		}
	}`
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"processed"`)
}

func TestWebhookHandler_InvalidWorkflowRunPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger)
	})

	payload := `{invalid json}`
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Failed to parse workflow_run event"`)
}

func TestWebhookHandler_TagBasedWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger)
	})

	payload := `{
		"action":"completed",
		"workflow_run":{
			"id":790,
			"name":"Release",
			"status":"completed",
			"conclusion":"success",
			"run_started_at":"2023-01-01T12:00:00Z",
			"updated_at":"2023-01-01T12:10:00Z",
			"head_branch":"v1.2.3",
			"event":"push",
			"head_sha":"bcb5820ced9479c074f688cc328bf03f341a511e"
		},
		"repository":{
			"full_name":"owner/repo"
		}
	}`
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"processed"`)
}
