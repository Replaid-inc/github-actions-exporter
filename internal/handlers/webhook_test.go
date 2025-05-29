package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
		WebhookHandler(c, processor, logger, "")
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
		WebhookHandler(c, processor, logger, "")
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
		WebhookHandler(c, processor, logger, "")
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
		WebhookHandler(c, processor, logger, "")
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

// generateSignature creates a GitHub-style HMAC signature for testing
func generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookHandler_ValidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	secret := "test-secret"

	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger, secret)
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

	signature := generateSignature([]byte(payload), secret)

	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"processed"`)
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	secret := "test-secret"

	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger, secret)
	})

	payload := `{"action":"completed","workflow_run":{"id":789}}`
	wrongSignature := "sha256=wrong_signature"

	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	req.Header.Set("X-Hub-Signature-256", wrongSignature)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid signature"`)
}

func TestWebhookHandler_MissingSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)
	secret := "test-secret"

	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger, secret)
	})

	payload := `{"action":"completed","workflow_run":{"id":789}}`

	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "workflow_run")
	// No signature header set
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"Invalid signature"`)
}

func TestWebhookHandler_NoSecretConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()
	registry := prometheus.NewRegistry()
	processor := metrics.NewMetricsProcessor(logger, registry)

	router.POST("/webhook", func(c *gin.Context) {
		WebhookHandler(c, processor, logger, "") // Empty secret
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
	// No signature header needed when no secret is configured
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"processed"`)
}

func TestVerifyGitHubSignature(t *testing.T) {
	payload := []byte(`{"test":"payload"}`)
	secret := "test-secret"

	// Test with valid signature
	validSignature := generateSignature(payload, secret)
	assert.True(t, verifyGitHubSignature(payload, validSignature, secret))

	// Test with invalid signature
	invalidSignature := "sha256=invalid"
	assert.False(t, verifyGitHubSignature(payload, invalidSignature, secret))

	// Test with missing sha256 prefix
	assert.False(t, verifyGitHubSignature(payload, "invalid", secret))

	// Test with empty signature
	assert.False(t, verifyGitHubSignature(payload, "", secret))

	// Test with empty secret (should allow all)
	assert.True(t, verifyGitHubSignature(payload, validSignature, ""))
	assert.True(t, verifyGitHubSignature(payload, "", ""))
}
