package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// WorkflowRunStatus represents the status of a workflow run
type WorkflowRunStatus string

// WorkflowRunConclusion represents the conclusion of a completed workflow run
type WorkflowRunConclusion string

// Constants for workflow run statuses
const (
	WorkflowRunStatusCompleted  WorkflowRunStatus = "completed"
	WorkflowRunStatusInProgress WorkflowRunStatus = "in_progress"
	WorkflowRunStatusRequested  WorkflowRunStatus = "requested"
)

// Constants for workflow run conclusions
const (
	WorkflowRunConclusionActionRequired WorkflowRunConclusion = "action_required"
	WorkflowRunConclusionCancelled      WorkflowRunConclusion = "cancelled"
	WorkflowRunConclusionFailure        WorkflowRunConclusion = "failure"
	WorkflowRunConclusionNeutral        WorkflowRunConclusion = "neutral"
	WorkflowRunConclusionSkipped        WorkflowRunConclusion = "skipped"
	WorkflowRunConclusionStale          WorkflowRunConclusion = "stale"
	WorkflowRunConclusionSuccess        WorkflowRunConclusion = "success"
	WorkflowRunConclusionTimedOut       WorkflowRunConclusion = "timed_out"
	WorkflowRunConclusionStartupFailure WorkflowRunConclusion = "startup_failure"
	WorkflowRunConclusionNull           WorkflowRunConclusion = "null"
)

// WorkflowRun represents a workflow run event
type WorkflowRun struct {
	ID         int64
	Name       string
	Repository string
	Status     WorkflowRunStatus
	Conclusion WorkflowRunConclusion
	StartedAt  time.Time
	UpdatedAt  time.Time
	Branch     string
	Trigger    string
	RefType    string // "branch" or "tag"
}

// MetricsProcessor processes GitHub webhook events and updates metrics
type MetricsProcessor struct {
	logger *zap.Logger

	// Prometheus metrics
	workflowStatus *prometheus.GaugeVec // New gauge metric for workflow status
}

// NewMetricsProcessor creates a new metrics processor
func NewMetricsProcessor(logger *zap.Logger, registry *prometheus.Registry) *MetricsProcessor {

	// Create new gauge for workflow status
	workflowStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "github_workflow_status",
			Help: "Current status of workflow runs (0=timed_out, 1=failure, 2=startup_failure, 3=cancelled, 4=skipped, 5=neutral, 6=stale, 7=null, 8=action_required, 9=in_progress, 10=success)",
		},
		[]string{"repository", "workflow", "branch", "trigger", "ref_type"},
	)

	// Register all metrics with the Prometheus registry
	registry.MustRegister(
		workflowStatus,
	)

	return &MetricsProcessor{
		logger:         logger,
		workflowStatus: workflowStatus,
	}
}

// ProcessWorkflowRun processes a workflow run event
func (p *MetricsProcessor) ProcessWorkflowRun(ctx context.Context, run WorkflowRun) error {

	// Update the workflow status gauge
	statusValue := 9.0 // Default: in_progress
	if run.Status == WorkflowRunStatusCompleted {
		switch run.Conclusion {
		case WorkflowRunConclusionTimedOut:
			statusValue = 0.0 // Timed out
		case WorkflowRunConclusionFailure:
			statusValue = 1.0 // Failure
		case WorkflowRunConclusionStartupFailure:
			statusValue = 2.0 // Startup failure
		case WorkflowRunConclusionCancelled:
			statusValue = 3.0 // Cancelled
		case WorkflowRunConclusionSkipped:
			statusValue = 4.0 // Skipped
		case WorkflowRunConclusionNeutral:
			statusValue = 5.0 // Neutral
		case WorkflowRunConclusionStale:
			statusValue = 6.0 // Stale
		case WorkflowRunConclusionNull:
			statusValue = 7.0 // Null
		case WorkflowRunConclusionActionRequired:
			statusValue = 8.0 // Action required
		case WorkflowRunConclusionSuccess:
			statusValue = 10.0 // Success
		}
	}

	p.workflowStatus.WithLabelValues(run.Repository, run.Name, run.Branch, run.Trigger, run.RefType).Set(statusValue)

	return nil
}
