package agentic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
var testTask = Task{
	ID:          "task-123",
	Title:       "Implement feature X",
	Description: "Add the new feature X to the system",
	Priority:    PriorityHigh,
	Status:      StatusPending,
	Labels:      []string{"feature", "backend"},
	Files:       []string{"pkg/feature/feature.go"},
	CreatedAt:   time.Now().Add(-24 * time.Hour),
	Project:     "core",
}

var testTasks = []Task{
	testTask,
	{
		ID:          "task-456",
		Title:       "Fix bug Y",
		Description: "Fix the bug in component Y",
		Priority:    PriorityCritical,
		Status:      StatusPending,
		Labels:      []string{"bug", "urgent"},
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		Project:     "core",
	},
}

func TestNewClient_Good(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")

	assert.Equal(t, "https://api.example.com", client.BaseURL)
	assert.Equal(t, "test-token", client.Token)
	assert.NotNil(t, client.HTTPClient)
}

func TestNewClient_Good_TrailingSlash(t *testing.T) {
	client := NewClient("https://api.example.com/", "test-token")

	assert.Equal(t, "https://api.example.com", client.BaseURL)
}

func TestNewClientFromConfig_Good(t *testing.T) {
	cfg := &Config{
		BaseURL: "https://api.example.com",
		Token:   "config-token",
		AgentID: "agent-001",
	}

	client := NewClientFromConfig(cfg)

	assert.Equal(t, "https://api.example.com", client.BaseURL)
	assert.Equal(t, "config-token", client.Token)
	assert.Equal(t, "agent-001", client.AgentID)
}

func TestClient_ListTasks_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/tasks", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testTasks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tasks, err := client.ListTasks(context.Background(), ListOptions{})

	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "task-123", tasks[0].ID)
	assert.Equal(t, "task-456", tasks[1].ID)
}

func TestClient_ListTasks_Good_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "pending", query.Get("status"))
		assert.Equal(t, "high", query.Get("priority"))
		assert.Equal(t, "core", query.Get("project"))
		assert.Equal(t, "10", query.Get("limit"))
		assert.Equal(t, "bug,urgent", query.Get("labels"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Task{testTask})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	opts := ListOptions{
		Status:   StatusPending,
		Priority: PriorityHigh,
		Project:  "core",
		Limit:    10,
		Labels:   []string{"bug", "urgent"},
	}

	tasks, err := client.ListTasks(context.Background(), opts)

	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestClient_ListTasks_Bad_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIError{Message: "internal error"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tasks, err := client.ListTasks(context.Background(), ListOptions{})

	assert.Error(t, err)
	assert.Nil(t, tasks)
	assert.Contains(t, err.Error(), "internal error")
}

func TestClient_GetTask_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/tasks/task-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testTask)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	task, err := client.GetTask(context.Background(), "task-123")

	require.NoError(t, err)
	assert.Equal(t, "task-123", task.ID)
	assert.Equal(t, "Implement feature X", task.Title)
	assert.Equal(t, PriorityHigh, task.Priority)
}

func TestClient_GetTask_Bad_EmptyID(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")
	task, err := client.GetTask(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task ID is required")
}

func TestClient_GetTask_Bad_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{Message: "task not found"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	task, err := client.GetTask(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task not found")
}

func TestClient_ClaimTask_Good(t *testing.T) {
	claimedTask := testTask
	claimedTask.Status = StatusInProgress
	claimedTask.ClaimedBy = "agent-001"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/tasks/task-123/claim", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ClaimResponse{Task: &claimedTask})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.AgentID = "agent-001"
	task, err := client.ClaimTask(context.Background(), "task-123")

	require.NoError(t, err)
	assert.Equal(t, StatusInProgress, task.Status)
	assert.Equal(t, "agent-001", task.ClaimedBy)
}

func TestClient_ClaimTask_Good_SimpleResponse(t *testing.T) {
	// Some APIs might return just the task without wrapping
	claimedTask := testTask
	claimedTask.Status = StatusInProgress

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claimedTask)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	task, err := client.ClaimTask(context.Background(), "task-123")

	require.NoError(t, err)
	assert.Equal(t, "task-123", task.ID)
}

func TestClient_ClaimTask_Bad_EmptyID(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")
	task, err := client.ClaimTask(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task ID is required")
}

func TestClient_ClaimTask_Bad_AlreadyClaimed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(APIError{Message: "task already claimed"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	task, err := client.ClaimTask(context.Background(), "task-123")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task already claimed")
}

func TestClient_UpdateTask_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/api/tasks/task-123", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var update TaskUpdate
		err := json.NewDecoder(r.Body).Decode(&update)
		require.NoError(t, err)
		assert.Equal(t, StatusInProgress, update.Status)
		assert.Equal(t, 50, update.Progress)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.UpdateTask(context.Background(), "task-123", TaskUpdate{
		Status:   StatusInProgress,
		Progress: 50,
		Notes:    "Making progress",
	})

	assert.NoError(t, err)
}

func TestClient_UpdateTask_Bad_EmptyID(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")
	err := client.UpdateTask(context.Background(), "", TaskUpdate{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task ID is required")
}

func TestClient_CompleteTask_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/tasks/task-123/complete", r.URL.Path)

		var result TaskResult
		err := json.NewDecoder(r.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "Feature implemented", result.Output)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.CompleteTask(context.Background(), "task-123", TaskResult{
		Success:   true,
		Output:    "Feature implemented",
		Artifacts: []string{"pkg/feature/feature.go"},
	})

	assert.NoError(t, err)
}

func TestClient_CompleteTask_Bad_EmptyID(t *testing.T) {
	client := NewClient("https://api.example.com", "test-token")
	err := client.CompleteTask(context.Background(), "", TaskResult{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task ID is required")
}

func TestClient_Ping_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.Ping(context.Background())

	assert.NoError(t, err)
}

func TestClient_Ping_Bad_ServerDown(t *testing.T) {
	client := NewClient("http://localhost:99999", "test-token")
	client.HTTPClient.Timeout = 100 * time.Millisecond

	err := client.Ping(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}

func TestAPIError_Error_Good(t *testing.T) {
	err := &APIError{
		Code:    404,
		Message: "task not found",
	}

	assert.Equal(t, "task not found", err.Error())

	err.Details = "task-123 does not exist"
	assert.Equal(t, "task not found: task-123 does not exist", err.Error())
}

func TestTaskStatus_Good(t *testing.T) {
	assert.Equal(t, TaskStatus("pending"), StatusPending)
	assert.Equal(t, TaskStatus("in_progress"), StatusInProgress)
	assert.Equal(t, TaskStatus("completed"), StatusCompleted)
	assert.Equal(t, TaskStatus("blocked"), StatusBlocked)
}

func TestTaskPriority_Good(t *testing.T) {
	assert.Equal(t, TaskPriority("critical"), PriorityCritical)
	assert.Equal(t, TaskPriority("high"), PriorityHigh)
	assert.Equal(t, TaskPriority("medium"), PriorityMedium)
	assert.Equal(t, TaskPriority("low"), PriorityLow)
}
