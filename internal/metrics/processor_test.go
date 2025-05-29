package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestWorkflowStatusGauge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := prometheus.NewRegistry()
	processor := NewMetricsProcessor(logger, registry)

	startTime := time.Now().Add(-60 * time.Second)
	endTime := time.Now()

	// Test scenario 1: In-progress workflow run should set gauge to 9.0
	err := processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1001,
		Name:       "gauge-test",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusInProgress,
		StartedAt:  startTime,
		UpdatedAt:  startTime,
		Branch:     "main",
		Trigger:    "push",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 9.0 (in_progress)
	gauge, err := processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test", "main", "push", "branch")
	require.NoError(t, err)
	val := testutil.ToFloat64(gauge)
	assert.Equal(t, 9.0, val, "Expected workflow status gauge to be 9.0 for in_progress")

	// Test scenario 2: Successful completion should set gauge to 10.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1001,
		Name:       "gauge-test",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionSuccess,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "main",
		Trigger:    "push",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 10.0 (success)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test", "main", "push", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 10.0, val, "Expected workflow status gauge to be 10.0 for success")

	// Test scenario 3: Failed completion should set gauge to 1.0
	// First create the in-progress workflow
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1002,
		Name:       "gauge-test-failure",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusInProgress,
		StartedAt:  startTime,
		UpdatedAt:  startTime,
		Branch:     "feature",
		Trigger:    "push",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Then complete it with failure
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1002,
		Name:       "gauge-test-failure",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionFailure,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "feature",
		Trigger:    "push",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 1.0 (failure)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-failure", "feature", "push", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 1.0, val, "Expected workflow status gauge to be 1.0 for failure")

	// Test scenario 4: Cancelled completion should set gauge to 3.0
	// First create the in-progress workflow
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1003,
		Name:       "gauge-test-cancelled",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusInProgress,
		StartedAt:  startTime,
		UpdatedAt:  startTime,
		Branch:     "main",
		Trigger:    "pull_request",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Then complete it with cancelled status
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1003,
		Name:       "gauge-test-cancelled",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionCancelled,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "main",
		Trigger:    "pull_request",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 3.0 (cancelled)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-cancelled", "main", "pull_request", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 3.0, val, "Expected workflow status gauge to be 3.0 for cancelled")

	// Test scenario 5: Neutral conclusion should set gauge to 5.0
	// First create the in-progress workflow
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1004,
		Name:       "gauge-test-neutral",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusInProgress,
		StartedAt:  startTime,
		UpdatedAt:  startTime,
		Branch:     "develop",
		Trigger:    "workflow_dispatch",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Then complete it with neutral conclusion
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1004,
		Name:       "gauge-test-neutral",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionNeutral,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "develop",
		Trigger:    "workflow_dispatch",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 5.0 (neutral)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-neutral", "develop", "workflow_dispatch", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 5.0, val, "Expected workflow status gauge to be 5.0 for neutral")

	// Test scenario 6: Timed out completion should set gauge to 0.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1005,
		Name:       "gauge-test-timedout",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionTimedOut,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "release",
		Trigger:    "schedule",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 0.0 (timed out)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-timedout", "release", "schedule", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 0.0, val, "Expected workflow status gauge to be 0.0 for timed_out")

	// Test scenario 6b: Tag-based workflow with timed out completion
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         10055,
		Name:       "gauge-test-tag",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionTimedOut,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "v1.0.0",
		Trigger:    "push",
		RefType:    "tag",
	})
	require.NoError(t, err)

	// Verify gauge value is 0.0 (timed out) for tag
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-tag", "v1.0.0", "push", "tag")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 0.0, val, "Expected workflow status gauge to be 0.0 for timed_out")

	// Test scenario 7: Startup failure should set gauge to 2.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1006,
		Name:       "gauge-test-startup-failure",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionStartupFailure,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "main",
		Trigger:    "push",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 2.0 (startup failure)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-startup-failure", "main", "push", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 2.0, val, "Expected workflow status gauge to be 2.0 for startup_failure")

	// Test scenario 8: Skipped completion should set gauge to 4.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1007,
		Name:       "gauge-test-skipped",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionSkipped,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "main",
		Trigger:    "pull_request",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 4.0 (skipped)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-skipped", "main", "pull_request", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 4.0, val, "Expected workflow status gauge to be 4.0 for skipped")

	// Test scenario 9: Action required should set gauge to 8.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1008,
		Name:       "gauge-test-action-required",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionActionRequired,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "feature/approval-needed",
		Trigger:    "workflow_dispatch",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 8.0 (action required)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-action-required", "feature/approval-needed", "workflow_dispatch", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 8.0, val, "Expected workflow status gauge to be 8.0 for action_required")

	// Test scenario 10: Stale should set gauge to 6.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1009,
		Name:       "gauge-test-stale",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionStale,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "stale-branch",
		Trigger:    "schedule",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 6.0 (stale)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-stale", "stale-branch", "schedule", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 6.0, val, "Expected workflow status gauge to be 6.0 for stale")

	// Test scenario 11: Null conclusion should set gauge to 7.0
	err = processor.ProcessWorkflowRun(context.Background(), WorkflowRun{
		ID:         1010,
		Name:       "gauge-test-null",
		Repository: "myorg/myrepo",
		Status:     WorkflowRunStatusCompleted,
		Conclusion: WorkflowRunConclusionNull,
		StartedAt:  startTime,
		UpdatedAt:  endTime,
		Branch:     "unknown",
		Trigger:    "unknown",
		RefType:    "branch",
	})
	require.NoError(t, err)

	// Verify gauge value is 7.0 (null)
	gauge, err = processor.workflowStatus.GetMetricWithLabelValues("myorg/myrepo", "gauge-test-null", "unknown", "unknown", "branch")
	require.NoError(t, err)
	val = testutil.ToFloat64(gauge)
	assert.Equal(t, 7.0, val, "Expected workflow status gauge to be 7.0 for null")
}
