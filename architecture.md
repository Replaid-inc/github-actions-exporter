# GitHub Actions Exporter - Project Architecture Specification

## Project Overview

GitHub Actions Exporter is a service that collects and exposes metrics about GitHub Actions workflows for monitoring with Prometheus. It processes GitHub webhook events to generate metrics on workflow status.

## Core Components

### 1. Webhook Receiver
- Implements an HTTP endpoint (`/webhook`) that accepts GitHub webhook POST requests
- Verifies webhook signatures using GitHub's secret token mechanism
- Parses incoming JSON payloads from GitHub webhook events
- Supports `workflow_run` event types
- Sends parsed events to the metrics processor

### 2. Metrics Processor
- Processes GitHub webhook events to update Prometheus metrics
- Updates the following metrics categories:
  - Workflow status (success, failure, cancelled, etc.)

### 3. Metrics Exposer
- Implements an HTTP endpoint (`/metrics`) that exposes Prometheus-formatted metrics
- Includes proper content type headers and formatting
- Provides real-time metric values on request

### Metric Definitions

#### Workflow Metrics:
- `github_workflow_status`: Gauge of current workflow status

## API Endpoints

### `/webhook` (POST):
- Accepts GitHub webhook payloads
- Requires GitHub signature header for verification (when secret is configured)
- Returns 200 OK if processed successfully

### `/metrics` (GET):
- Returns Prometheus-formatted metrics
- No authentication required (should be protected at network level)

### `/health` (GET):
- Returns health status of the service
- Used by monitoring systems to verify service is running

### Framework & Library Choices

#### HTTP Server
- Use [Gin](https://github.com/gin-gonic/gin) web framework for handling HTTP requests
- Benefits: High performance, middleware support, robust routing

#### Prometheus Integration
- Use the official [Prometheus Go client](https://github.com/prometheus/client_golang)
- Implement custom metrics collectors as needed

#### Logging
- Use [zap](https://github.com/uber-go/zap) for structured, performant logging
- Configure log levels based on environment (development/production)

#### Testing
- Use standard Go testing package with [testify](https://github.com/stretchr/testify) for assertions
- Implement unit tests for all core components
- Use [httptest](https://golang.org/pkg/net/http/httptest/) for API endpoint testing
