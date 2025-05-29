package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"

	"gh-actions-exporter/internal/metrics"
)

// GitHubWorkflowRunEvent represents the workflow_run event payload structure
type GitHubWorkflowRunEvent struct {
	Action      string `json:"action"`
	WorkflowRun struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		StartedAt  string `json:"run_started_at"`
		UpdatedAt  string `json:"updated_at"`
		HeadBranch string `json:"head_branch"` // Branch name
		Event      string `json:"event"`       // Trigger event type (push, pull_request, etc.)
		HeadCommit struct {
			ID string `json:"id"`
		} `json:"head_commit"`
		HeadSHA string `json:"head_sha"` // The SHA of the head commit
	} `json:"workflow_run"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// verifyGitHubSignature verifies the GitHub webhook signature
func verifyGitHubSignature(payload []byte, signature string, secret string) bool {
	if secret == "" {
		// If no secret is configured, skip verification
		return true
	}

	if signature == "" {
		return false
	}

	// Remove the "sha256=" prefix from the signature
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	signature = signature[7:]

	// Create HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// WebhookHandler processes GitHub webhook events
func WebhookHandler(c *gin.Context, processor *metrics.MetricsProcessor, logger *zap.Logger, webhookSecret string) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(400, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify GitHub signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if !verifyGitHubSignature(body, signature, webhookSecret) {
		logger.Error("Invalid webhook signature")
		c.JSON(401, gin.H{"error": "Invalid signature"})
		return
	}

	// Get the GitHub event type from the header
	eventType := c.GetHeader("X-GitHub-Event")
	logger.Debug("Received webhook", zap.String("event", eventType))

	// Process based on event type
	switch eventType {
	case "workflow_run":
		processWorkflowRunEvent(c, body, processor, logger)
	default:
		logger.Debug("Ignoring unsupported event type", zap.String("event", eventType))
		c.JSON(200, gin.H{"status": "ignored", "event": eventType})
		return
	}

	c.JSON(200, gin.H{"status": "processed"})
}

// processWorkflowRunEvent handles workflow_run events
func processWorkflowRunEvent(c *gin.Context, body []byte, processor *metrics.MetricsProcessor, logger *zap.Logger) {
	var event GitHubWorkflowRunEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("Failed to parse workflow_run event", zap.Error(err))
		c.JSON(400, gin.H{"error": "Failed to parse workflow_run event"})
		return
	}

	// Parse time fields
	startedAt, _ := time.Parse(time.RFC3339, event.WorkflowRun.StartedAt)
	updatedAt, _ := time.Parse(time.RFC3339, event.WorkflowRun.UpdatedAt)

	// Determine ref type (branch or tag)
	refType := "branch"
	refName := event.WorkflowRun.HeadBranch

	// For tag-based triggers, we need to detect this based on several factors:
	// 1. The event type is typically "push"
	// 2. The branch name might match common tag patterns (v1.0.0, tags/v1, etc.)

	// Common tag naming patterns
	isLikelyTag := false

	if len(refName) > 0 {
		// Check for semantic version patterns
		if refName[0] == 'v' && len(refName) > 1 {
			// v1.0.0 pattern
			isLikelyTag = true
		} else if strings.HasPrefix(refName, "tags/") {
			// tags/v1.0.0 pattern
			isLikelyTag = true
		} else if strings.HasPrefix(refName, "refs/tags/") {
			// refs/tags/v1.0.0 pattern
			isLikelyTag = true
		}
	}

	if event.WorkflowRun.Event == "push" && isLikelyTag {
		// This is likely a tag-based workflow
		refType = "tag"
	}

	// Create workflow run object
	run := metrics.WorkflowRun{
		ID:         event.WorkflowRun.ID,
		Name:       event.WorkflowRun.Name,
		Repository: event.Repository.FullName,
		Status:     metrics.WorkflowRunStatus(event.WorkflowRun.Status),
		Conclusion: metrics.WorkflowRunConclusion(event.WorkflowRun.Conclusion),
		StartedAt:  startedAt,
		UpdatedAt:  updatedAt,
		Branch:     refName,
		Trigger:    event.WorkflowRun.Event,
		RefType:    refType,
	}

	// Process the workflow run
	if err := processor.ProcessWorkflowRun(c.Request.Context(), run); err != nil {
		logger.Error("Failed to process workflow run",
			zap.Error(err),
			zap.Int64("runID", run.ID),
			zap.String("repository", run.Repository))
	} else {
		logger.Debug("Successfully processed workflow run",
			zap.Int64("runID", run.ID),
			zap.String("status", string(run.Status)))
	}
}
