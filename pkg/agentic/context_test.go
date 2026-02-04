package agentic

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskContext_Good(t *testing.T) {
	// Create a temp directory with some files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	require.NoError(t, err)

	task := &Task{
		ID:          "test-123",
		Title:       "Test Task",
		Description: "A test task description",
		Priority:    PriorityMedium,
		Status:      StatusPending,
		Files:       []string{"main.go"},
		CreatedAt:   time.Now(),
	}

	ctx, err := BuildTaskContext(task, tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.Equal(t, task, ctx.Task)
	assert.Len(t, ctx.Files, 1)
	assert.Equal(t, "main.go", ctx.Files[0].Path)
	assert.Equal(t, "go", ctx.Files[0].Language)
}

func TestBuildTaskContext_Bad_NilTask(t *testing.T) {
	ctx, err := BuildTaskContext(nil, ".")
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "task is required")
}

func TestGatherRelatedFiles_Good(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"app.go":    "package app\n\nfunc Run() {}\n",
		"config.ts": "export const config = {};\n",
		"README.md": "# Project\n",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	task := &Task{
		ID:    "task-1",
		Title: "Test",
		Files: []string{"app.go", "config.ts"},
	}

	gathered, err := GatherRelatedFiles(task, tmpDir)
	require.NoError(t, err)
	assert.Len(t, gathered, 2)

	// Check languages detected correctly
	foundGo := false
	foundTS := false
	for _, f := range gathered {
		if f.Path == "app.go" {
			foundGo = true
			assert.Equal(t, "go", f.Language)
		}
		if f.Path == "config.ts" {
			foundTS = true
			assert.Equal(t, "typescript", f.Language)
		}
	}
	assert.True(t, foundGo, "should find app.go")
	assert.True(t, foundTS, "should find config.ts")
}

func TestGatherRelatedFiles_Bad_NilTask(t *testing.T) {
	files, err := GatherRelatedFiles(nil, ".")
	assert.Error(t, err)
	assert.Nil(t, files)
}

func TestGatherRelatedFiles_Good_MissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	task := &Task{
		ID:    "task-1",
		Title: "Test",
		Files: []string{"nonexistent.go", "also-missing.ts"},
	}

	// Should not error, just return empty list
	gathered, err := GatherRelatedFiles(task, tmpDir)
	require.NoError(t, err)
	assert.Empty(t, gathered)
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "go"},
		{"app.ts", "typescript"},
		{"app.tsx", "typescript"},
		{"script.js", "javascript"},
		{"script.jsx", "javascript"},
		{"main.py", "python"},
		{"lib.rs", "rust"},
		{"App.java", "java"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"data.json", "json"},
		{"index.html", "html"},
		{"styles.css", "css"},
		{"styles.scss", "scss"},
		{"query.sql", "sql"},
		{"README.md", "markdown"},
		{"unknown.xyz", "text"},
		{"", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int // minimum number of keywords expected
	}{
		{
			name:     "simple title",
			text:     "Add user authentication feature",
			expected: 2,
		},
		{
			name:     "with stop words",
			text:     "The quick brown fox jumps over the lazy dog",
			expected: 3,
		},
		{
			name:     "technical text",
			text:     "Implement OAuth2 authentication with JWT tokens",
			expected: 3,
		},
		{
			name:     "empty",
			text:     "",
			expected: 0,
		},
		{
			name:     "only stop words",
			text:     "the a an and or but in on at",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywords(tt.text)
			assert.GreaterOrEqual(t, len(keywords), tt.expected)
			// Keywords should not exceed 5
			assert.LessOrEqual(t, len(keywords), 5)
		})
	}
}

func TestTaskContext_FormatContext(t *testing.T) {
	task := &Task{
		ID:          "test-456",
		Title:       "Test Formatting",
		Description: "This is a test description",
		Priority:    PriorityHigh,
		Status:      StatusInProgress,
	}

	ctx := &TaskContext{
		Task:          task,
		Files:         []FileContent{{Path: "main.go", Content: "package main", Language: "go"}},
		GitStatus:     " M main.go",
		RecentCommits: "abc123 Initial commit",
		RelatedCode:   []FileContent{{Path: "util.go", Content: "package util", Language: "go"}},
	}

	formatted := ctx.FormatContext()

	assert.Contains(t, formatted, "# Task Context")
	assert.Contains(t, formatted, "test-456")
	assert.Contains(t, formatted, "Test Formatting")
	assert.Contains(t, formatted, "## Task Files")
	assert.Contains(t, formatted, "## Git Status")
	assert.Contains(t, formatted, "## Recent Commits")
	assert.Contains(t, formatted, "## Related Code")
}
