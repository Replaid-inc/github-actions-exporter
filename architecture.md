# GitHub Actions Exporter - Project Architecture Specification

## Project Overview

GitHub Actions Exporter is a service that collects and exposes metrics about GitHub Actions workflows and jobs for monitoring with Prometheus. It processes GitHub webhook events to generate metrics on workflow/job performance, status, and cost.

## Core Components

### 1. Webhook Receiver
- Implements an HTTP endpoint (`/webhook`) that accepts GitHub webhook POST requests
- Verifies webhook signatures using GitHub's secret token mechanism
- Parses incoming JSON payloads from GitHub webhook events
- Supports `workflow_run` and `workflow_job` event types
- Sends parsed events to the metrics processor

### 2. Metrics Processor
- Processes GitHub webhook events to update Prometheus metrics
- Maintains state for in-progress workflows/jobs to calculate durations
- Updates the following metrics categories:
  - Workflow status (success, failure, cancelled, etc.)
  - Workflow duration (time from start to completion)
  - Job status (success, failure, queued, etc.)
  - Job duration (time from start to completion)
  - Job queue time (time from creation until execution begins)

### 3. Metrics Exposer
- Implements an HTTP endpoint (`/metrics`) that exposes Prometheus-formatted metrics
- Includes proper content type headers and formatting
- Provides real-time metric values on request

### 4. Configuration System
- Loads configuration from YAML files and environment variables
- Configures the following aspects:
  - HTTP server settings (port, TLS, etc.)
  - GitHub webhook secret for verification
  - Label customization for jobs and workflows
  - Metric name prefixes

## Data Models

### GitHub Webhook Payloads
#### WorkflowRunEvent:
- Contains metadata about workflow executions
- Includes workflow name, repository, status, timing information

#### WorkflowJobEvent:
- Contains metadata about individual job executions
- Includes job name, status, timing information

### Metric Definitions
#### Workflow Metrics:
- `workflow_run_total`: Counter of workflow runs by status
- `workflow_run_duration_seconds`: Histogram of workflow execution times

#### Job Metrics:
- `job_run_total`: Counter of job runs by status
- `job_run_duration_seconds`: Histogram of job execution times
- `job_queue_duration_seconds`: Histogram of job queue times
- `job_cost`: Gauge for estimated cost of job execution

## API Endpoints

### `/webhook` (POST):
- Accepts GitHub webhook payloads
- Requires GitHub signature header for verification
- Returns 200 OK if processed successfully

### `/metrics` (GET):
- Returns Prometheus-formatted metrics
- No authentication required (should be protected at network level)

### `/health` (GET):
- Returns health status of the service
- Used by monitoring systems to verify service is running

## Deployment Aspects
- Service should be containerized for easy deployment
- Configuration through environment variables for container compatibility
- Low resource requirements (memory/CPU)
- Persistent storage not required (metrics reset on restart is acceptable)
- Should handle webhook event bursts without dropping events

## Error Handling
- Log invalid webhook payloads but don't crash
- Graceful handling of GitHub API changes
- Proper error responses for invalid requests
- Metric for tracking webhook processing errors

## Security Considerations
- Webhook signature verification is mandatory
- No storage of repository secrets or sensitive information
- Metrics endpoint should be protected (e.g., network controls)
- No outbound API calls to GitHub (only processes incoming webhooks)

## Technical Implementation

### Programming Language
- The service will be implemented in Go (Golang)
- Target Go version: 1.21 or newer
- Follow standard Go project layout conventions

### Framework & Library Choices

#### HTTP Server
- Use [Gin](https://github.com/gin-gonic/gin) web framework for handling HTTP requests
- Benefits: High performance, middleware support, robust routing

#### Prometheus Integration
- Use the official [Prometheus Go client](https://github.com/prometheus/client_golang)
- Implement custom metrics collectors as needed

#### Configuration Management
- Use [Viper](https://github.com/spf13/viper) for configuration management
  - Support for YAML, environment variables, and flags
  - Hot reloading of configuration

#### Logging
- Use [zap](https://github.com/uber-go/zap) for structured, performant logging
- Configure log levels based on environment (development/production)

#### Testing
- Use standard Go testing package with [testify](https://github.com/stretchr/testify) for assertions
- Implement unit tests for all core components
- Use [httptest](https://golang.org/pkg/net/http/httptest/) for API endpoint testing

#### Dependency Management
- Use Go modules for dependency management
- Vendor dependencies for reproducible builds

#### Error Handling
- Use structured error handling with appropriate context
- Implement custom error types for domain-specific errors
- Use [errors](https://github.com/pkg/errors) package for error wrapping

### Development Practices
- Follow Go best practices and idioms
- Use Go linters (golangci-lint) for code quality
- Document public APIs with godoc-compatible comments
- Use context for timeouts and cancellation

### Build & Deployment
- Create a minimal Docker image using multi-stage builds
- Use [Task](https://taskfile.dev) instead of Make for common development tasks
  - Define tasks for build, test, lint, run, and other operations
  - Provides a cross-platform alternative to Makefiles
  - Uses YAML for better readability and maintainability
- Support graceful shutdown for container orchestration