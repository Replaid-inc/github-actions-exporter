package models

// WorkflowRunEvent represents the payload for workflow run events from GitHub.
type WorkflowRunEvent struct {
    Action      string `json:"action"`
    Workflow    string `json:"workflow"`
    Repository  Repository `json:"repository"`
    Sender      User `json:"sender"`
    // Add other relevant fields as needed
}

// WorkflowJobEvent represents the payload for workflow job events from GitHub.
type WorkflowJobEvent struct {
    Action      string `json:"action"`
    Job         Job `json:"job"`
    Repository  Repository `json:"repository"`
    Sender      User `json:"sender"`
    // Add other relevant fields as needed
}

// Repository represents a GitHub repository.
type Repository struct {
    Name     string `json:"name"`
    Owner    User `json:"owner"`
    // Add other relevant fields as needed
}

// User represents a GitHub user.
type User struct {
    Login string `json:"login"`
    // Add other relevant fields as needed
}

// Job represents a GitHub Actions job.
type Job struct {
    Name   string `json:"name"`
    Status string `json:"status"`
    // Add other relevant fields as needed
}