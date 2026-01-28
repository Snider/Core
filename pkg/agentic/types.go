// Package agentic provides an API client for core-agentic, an AI-assisted task
// management service. It enables developers and AI agents to discover, claim,
// and complete development tasks.
package agentic

import (
	"time"
)

// TaskStatus represents the state of a task in the system.
type TaskStatus string

const (
	// StatusPending indicates the task is available to be claimed.
	StatusPending TaskStatus = "pending"
	// StatusInProgress indicates the task has been claimed and is being worked on.
	StatusInProgress TaskStatus = "in_progress"
	// StatusCompleted indicates the task has been successfully completed.
	StatusCompleted TaskStatus = "completed"
	// StatusBlocked indicates the task cannot proceed due to dependencies.
	StatusBlocked TaskStatus = "blocked"
)

// TaskPriority represents the urgency level of a task.
type TaskPriority string

const (
	// PriorityCritical indicates the task requires immediate attention.
	PriorityCritical TaskPriority = "critical"
	// PriorityHigh indicates the task is important and should be addressed soon.
	PriorityHigh TaskPriority = "high"
	// PriorityMedium indicates the task has normal priority.
	PriorityMedium TaskPriority = "medium"
	// PriorityLow indicates the task can be addressed when time permits.
	PriorityLow TaskPriority = "low"
)

// Task represents a development task in the core-agentic system.
type Task struct {
	// ID is the unique identifier for the task.
	ID string `json:"id"`
	// Title is the short description of the task.
	Title string `json:"title"`
	// Description provides detailed information about what needs to be done.
	Description string `json:"description"`
	// Priority indicates the urgency of the task.
	Priority TaskPriority `json:"priority"`
	// Status indicates the current state of the task.
	Status TaskStatus `json:"status"`
	// Labels are tags used to categorize the task.
	Labels []string `json:"labels,omitempty"`
	// Files lists the files that are relevant to this task.
	Files []string `json:"files,omitempty"`
	// CreatedAt is when the task was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the task was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// ClaimedBy is the identifier of the agent or developer who claimed the task.
	ClaimedBy string `json:"claimed_by,omitempty"`
	// ClaimedAt is when the task was claimed.
	ClaimedAt *time.Time `json:"claimed_at,omitempty"`
	// Project is the project this task belongs to.
	Project string `json:"project,omitempty"`
	// Dependencies lists task IDs that must be completed before this task.
	Dependencies []string `json:"dependencies,omitempty"`
	// Blockers lists task IDs that this task is blocking.
	Blockers []string `json:"blockers,omitempty"`
}

// TaskUpdate contains fields that can be updated on a task.
type TaskUpdate struct {
	// Status is the new status for the task.
	Status TaskStatus `json:"status,omitempty"`
	// Progress is a percentage (0-100) indicating completion.
	Progress int `json:"progress,omitempty"`
	// Notes are additional comments about the update.
	Notes string `json:"notes,omitempty"`
}

// TaskResult contains the outcome of a completed task.
type TaskResult struct {
	// Success indicates whether the task was completed successfully.
	Success bool `json:"success"`
	// Output is the result or summary of the completed work.
	Output string `json:"output,omitempty"`
	// Artifacts are files or resources produced by the task.
	Artifacts []string `json:"artifacts,omitempty"`
	// ErrorMessage contains details if the task failed.
	ErrorMessage string `json:"error_message,omitempty"`
}

// ListOptions specifies filters for listing tasks.
type ListOptions struct {
	// Status filters tasks by their current status.
	Status TaskStatus `json:"status,omitempty"`
	// Labels filters tasks that have all specified labels.
	Labels []string `json:"labels,omitempty"`
	// Priority filters tasks by priority level.
	Priority TaskPriority `json:"priority,omitempty"`
	// Limit is the maximum number of tasks to return.
	Limit int `json:"limit,omitempty"`
	// Project filters tasks by project.
	Project string `json:"project,omitempty"`
	// ClaimedBy filters tasks claimed by a specific agent.
	ClaimedBy string `json:"claimed_by,omitempty"`
}

// APIError represents an error response from the API.
type APIError struct {
	// Code is the HTTP status code.
	Code int `json:"code"`
	// Message is the error description.
	Message string `json:"message"`
	// Details provides additional context about the error.
	Details string `json:"details,omitempty"`
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// ClaimResponse is returned when a task is successfully claimed.
type ClaimResponse struct {
	// Task is the claimed task with updated fields.
	Task *Task `json:"task"`
	// Message provides additional context about the claim.
	Message string `json:"message,omitempty"`
}

// CompleteResponse is returned when a task is completed.
type CompleteResponse struct {
	// Task is the completed task with final status.
	Task *Task `json:"task"`
	// Message provides additional context about the completion.
	Message string `json:"message,omitempty"`
}
