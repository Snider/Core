# MCP Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `core mcp serve` command with RAG and metrics tools, then configure the agentic-flows plugin to use it.

**Architecture:** Create a new `mcp` command package that starts the pkg/mcp server with extended tools. RAG tools call the existing exported functions in internal/cmd/rag. Metrics tools call pkg/ai directly. The agentic-flows plugin gets a `.mcp.json` that spawns `core mcp serve`.

**Tech Stack:** Go 1.25, github.com/modelcontextprotocol/go-sdk/mcp, pkg/rag, pkg/ai

---

## Task 1: Add RAG tools to pkg/mcp

**Files:**
- Create: `pkg/mcp/tools_rag.go`
- Modify: `pkg/mcp/mcp.go:99-101` (registerTools)
- Test: `pkg/mcp/tools_rag_test.go`

**Step 1: Write the failing test**

Create `pkg/mcp/tools_rag_test.go`:

```go
package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRAGQueryTool_Good(t *testing.T) {
	// This test verifies the tool is registered and callable.
	// It doesn't require Qdrant/Ollama running - just checks structure.
	s, err := New(WithWorkspaceRoot(""))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Check that rag_query tool is registered
	tools := s.Server().ListTools()
	found := false
	for _, tool := range tools {
		if tool.Name == "rag_query" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rag_query tool not registered")
	}
}

func TestRAGQueryInput_Good(t *testing.T) {
	input := RAGQueryInput{
		Question:   "how do I deploy?",
		Collection: "hostuk-docs",
		TopK:       5,
	}
	if input.Question == "" {
		t.Error("Question should not be empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestRAGQueryTool ./pkg/mcp/... -v`
Expected: FAIL with "rag_query tool not registered"

**Step 3: Create tools_rag.go with types and tool registration**

Create `pkg/mcp/tools_rag.go`:

```go
package mcp

import (
	"context"
	"fmt"

	ragcmd "github.com/host-uk/core/internal/cmd/rag"
	"github.com/host-uk/core/pkg/rag"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RAG tool input/output types

// RAGQueryInput contains parameters for querying the vector database.
type RAGQueryInput struct {
	Question   string `json:"question"`
	Collection string `json:"collection,omitempty"`
	TopK       int    `json:"top_k,omitempty"`
}

// RAGQueryOutput contains the query results.
type RAGQueryOutput struct {
	Results []RAGResult `json:"results"`
	Context string      `json:"context"`
}

// RAGResult represents a single search result.
type RAGResult struct {
	Content  string            `json:"content"`
	Score    float32           `json:"score"`
	Source   string            `json:"source"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RAGIngestInput contains parameters for ingesting documents.
type RAGIngestInput struct {
	Path       string `json:"path"`
	Collection string `json:"collection,omitempty"`
	Recreate   bool   `json:"recreate,omitempty"`
}

// RAGIngestOutput contains the ingestion results.
type RAGIngestOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Chunks  int    `json:"chunks"`
	Message string `json:"message,omitempty"`
}

// RAGCollectionsInput contains parameters for listing collections.
type RAGCollectionsInput struct {
	ShowStats bool `json:"show_stats,omitempty"`
}

// RAGCollectionsOutput contains the list of collections.
type RAGCollectionsOutput struct {
	Collections []CollectionInfo `json:"collections"`
}

// CollectionInfo describes a Qdrant collection.
type CollectionInfo struct {
	Name        string `json:"name"`
	PointsCount uint64 `json:"points_count,omitempty"`
	Status      string `json:"status,omitempty"`
}

// registerRAGTools adds RAG tools to the MCP server.
func (s *Service) registerRAGTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_query",
		Description: "Query the vector database for relevant documents using semantic search",
	}, s.ragQuery)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_ingest",
		Description: "Ingest a file or directory into the vector database",
	}, s.ragIngest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_collections",
		Description: "List available vector database collections",
	}, s.ragCollections)
}

func (s *Service) ragQuery(ctx context.Context, req *mcp.CallToolRequest, input RAGQueryInput) (*mcp.CallToolResult, RAGQueryOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "rag_query", "question", input.Question)

	collection := input.Collection
	if collection == "" {
		collection = "hostuk-docs"
	}
	topK := input.TopK
	if topK <= 0 {
		topK = 5
	}

	results, err := ragcmd.QueryDocs(ctx, input.Question, collection, topK)
	if err != nil {
		return nil, RAGQueryOutput{}, fmt.Errorf("query failed: %w", err)
	}

	// Convert to output format
	out := RAGQueryOutput{
		Results: make([]RAGResult, 0, len(results)),
		Context: rag.FormatResultsContext(results),
	}
	for _, r := range results {
		out.Results = append(out.Results, RAGResult{
			Content:  r.Content,
			Score:    r.Score,
			Source:   r.Source,
			Metadata: r.Metadata,
		})
	}

	return nil, out, nil
}

func (s *Service) ragIngest(ctx context.Context, req *mcp.CallToolRequest, input RAGIngestInput) (*mcp.CallToolResult, RAGIngestOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "rag_ingest", "path", input.Path)

	collection := input.Collection
	if collection == "" {
		collection = "hostuk-docs"
	}

	// Check if path is a file or directory
	info, err := s.medium.Stat(input.Path)
	if err != nil {
		return nil, RAGIngestOutput{}, fmt.Errorf("path not found: %w", err)
	}

	if info.IsDir() {
		err = ragcmd.IngestDirectory(ctx, input.Path, collection, input.Recreate)
		if err != nil {
			return nil, RAGIngestOutput{}, fmt.Errorf("ingest directory failed: %w", err)
		}
		return nil, RAGIngestOutput{
			Success: true,
			Path:    input.Path,
			Message: fmt.Sprintf("Ingested directory into collection %s", collection),
		}, nil
	}

	chunks, err := ragcmd.IngestFile(ctx, input.Path, collection)
	if err != nil {
		return nil, RAGIngestOutput{}, fmt.Errorf("ingest file failed: %w", err)
	}

	return nil, RAGIngestOutput{
		Success: true,
		Path:    input.Path,
		Chunks:  chunks,
		Message: fmt.Sprintf("Ingested %d chunks into collection %s", chunks, collection),
	}, nil
}

func (s *Service) ragCollections(ctx context.Context, req *mcp.CallToolRequest, input RAGCollectionsInput) (*mcp.CallToolResult, RAGCollectionsOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "rag_collections")

	client, err := rag.NewQdrantClient(rag.DefaultQdrantConfig())
	if err != nil {
		return nil, RAGCollectionsOutput{}, fmt.Errorf("connect to Qdrant: %w", err)
	}
	defer func() { _ = client.Close() }()

	names, err := client.ListCollections(ctx)
	if err != nil {
		return nil, RAGCollectionsOutput{}, fmt.Errorf("list collections: %w", err)
	}

	out := RAGCollectionsOutput{
		Collections: make([]CollectionInfo, 0, len(names)),
	}

	for _, name := range names {
		info := CollectionInfo{Name: name}
		if input.ShowStats {
			cinfo, err := client.CollectionInfo(ctx, name)
			if err == nil {
				info.PointsCount = cinfo.PointsCount
				info.Status = cinfo.Status.String()
			}
		}
		out.Collections = append(out.Collections, info)
	}

	return nil, out, nil
}
```

**Step 4: Update mcp.go to call registerRAGTools**

In `pkg/mcp/mcp.go`, modify the `registerTools` function (around line 104) to add:

```go
func (s *Service) registerTools(server *mcp.Server) {
	// File operations (existing)
	// ... existing code ...

	// RAG operations
	s.registerRAGTools(server)
}
```

**Step 5: Run test to verify it passes**

Run: `go test -run TestRAGQuery ./pkg/mcp/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/mcp/tools_rag.go pkg/mcp/tools_rag_test.go pkg/mcp/mcp.go
git commit -m "feat(mcp): add RAG tools (query, ingest, collections)"
```

---

## Task 2: Add metrics tools to pkg/mcp

**Files:**
- Create: `pkg/mcp/tools_metrics.go`
- Modify: `pkg/mcp/mcp.go` (registerTools)
- Test: `pkg/mcp/tools_metrics_test.go`

**Step 1: Write the failing test**

Create `pkg/mcp/tools_metrics_test.go`:

```go
package mcp

import (
	"testing"
)

func TestMetricsRecordTool_Good(t *testing.T) {
	s, err := New(WithWorkspaceRoot(""))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tools := s.Server().ListTools()
	found := false
	for _, tool := range tools {
		if tool.Name == "metrics_record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("metrics_record tool not registered")
	}
}

func TestMetricsQueryTool_Good(t *testing.T) {
	s, err := New(WithWorkspaceRoot(""))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tools := s.Server().ListTools()
	found := false
	for _, tool := range tools {
		if tool.Name == "metrics_query" {
			found = true
			break
		}
	}
	if !found {
		t.Error("metrics_query tool not registered")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestMetrics ./pkg/mcp/... -v`
Expected: FAIL

**Step 3: Create tools_metrics.go**

Create `pkg/mcp/tools_metrics.go`:

```go
package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/ai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Metrics tool input/output types

// MetricsRecordInput contains parameters for recording a metric event.
type MetricsRecordInput struct {
	Type    string         `json:"type"`
	AgentID string         `json:"agent_id,omitempty"`
	Repo    string         `json:"repo,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

// MetricsRecordOutput contains the result of recording.
type MetricsRecordOutput struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricsQueryInput contains parameters for querying metrics.
type MetricsQueryInput struct {
	Since string `json:"since,omitempty"` // e.g., "7d", "24h"
}

// MetricsQueryOutput contains the query results.
type MetricsQueryOutput struct {
	Total   int                `json:"total"`
	ByType  []MetricCount      `json:"by_type"`
	ByRepo  []MetricCount      `json:"by_repo"`
	ByAgent []MetricCount      `json:"by_agent"`
	Events  []MetricEventBrief `json:"events,omitempty"`
}

// MetricCount represents a count by key.
type MetricCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// MetricEventBrief is a simplified event for output.
type MetricEventBrief struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	AgentID   string    `json:"agent_id,omitempty"`
	Repo      string    `json:"repo,omitempty"`
}

// registerMetricsTools adds metrics tools to the MCP server.
func (s *Service) registerMetricsTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "metrics_record",
		Description: "Record a metric event (AI task, security scan, job creation, etc.)",
	}, s.metricsRecord)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "metrics_query",
		Description: "Query recorded metrics with aggregation by type, repo, and agent",
	}, s.metricsQuery)
}

func (s *Service) metricsRecord(ctx context.Context, req *mcp.CallToolRequest, input MetricsRecordInput) (*mcp.CallToolResult, MetricsRecordOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "metrics_record", "type", input.Type)

	if input.Type == "" {
		return nil, MetricsRecordOutput{}, fmt.Errorf("type is required")
	}

	event := ai.Event{
		Type:      input.Type,
		Timestamp: time.Now(),
		AgentID:   input.AgentID,
		Repo:      input.Repo,
		Data:      input.Data,
	}

	if err := ai.Record(event); err != nil {
		return nil, MetricsRecordOutput{}, fmt.Errorf("record event: %w", err)
	}

	return nil, MetricsRecordOutput{
		Success:   true,
		Timestamp: event.Timestamp,
	}, nil
}

func (s *Service) metricsQuery(ctx context.Context, req *mcp.CallToolRequest, input MetricsQueryInput) (*mcp.CallToolResult, MetricsQueryOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "metrics_query", "since", input.Since)

	since := input.Since
	if since == "" {
		since = "7d"
	}

	duration, err := parseDuration(since)
	if err != nil {
		return nil, MetricsQueryOutput{}, fmt.Errorf("invalid since value: %w", err)
	}

	sinceTime := time.Now().Add(-duration)
	events, err := ai.ReadEvents(sinceTime)
	if err != nil {
		return nil, MetricsQueryOutput{}, fmt.Errorf("read events: %w", err)
	}

	summary := ai.Summary(events)

	out := MetricsQueryOutput{
		Total: summary["total"].(int),
	}

	// Convert by_type
	if byType, ok := summary["by_type"].([]map[string]any); ok {
		for _, entry := range byType {
			out.ByType = append(out.ByType, MetricCount{
				Key:   entry["key"].(string),
				Count: entry["count"].(int),
			})
		}
	}

	// Convert by_repo
	if byRepo, ok := summary["by_repo"].([]map[string]any); ok {
		for _, entry := range byRepo {
			out.ByRepo = append(out.ByRepo, MetricCount{
				Key:   entry["key"].(string),
				Count: entry["count"].(int),
			})
		}
	}

	// Convert by_agent
	if byAgent, ok := summary["by_agent"].([]map[string]any); ok {
		for _, entry := range byAgent {
			out.ByAgent = append(out.ByAgent, MetricCount{
				Key:   entry["key"].(string),
				Count: entry["count"].(int),
			})
		}
	}

	// Include last 10 events for context
	limit := 10
	if len(events) < limit {
		limit = len(events)
	}
	for i := len(events) - limit; i < len(events); i++ {
		ev := events[i]
		out.Events = append(out.Events, MetricEventBrief{
			Type:      ev.Type,
			Timestamp: ev.Timestamp,
			AgentID:   ev.AgentID,
			Repo:      ev.Repo,
		})
	}

	return nil, out, nil
}

// parseDuration parses a human-friendly duration like "7d", "24h", "30d".
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	unit := s[len(s)-1]
	value := s[:len(s)-1]

	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	if n <= 0 {
		return 0, fmt.Errorf("duration must be positive: %s", s)
	}

	switch unit {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'm':
		return time.Duration(n) * time.Minute, nil
	default:
		return 0, fmt.Errorf("unknown unit %c in duration: %s", unit, s)
	}
}
```

**Step 4: Update mcp.go to call registerMetricsTools**

In `pkg/mcp/mcp.go`, add to `registerTools`:

```go
func (s *Service) registerTools(server *mcp.Server) {
	// ... existing file operations ...

	// RAG operations
	s.registerRAGTools(server)

	// Metrics operations
	s.registerMetricsTools(server)
}
```

**Step 5: Run test to verify it passes**

Run: `go test -run TestMetrics ./pkg/mcp/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/mcp/tools_metrics.go pkg/mcp/tools_metrics_test.go pkg/mcp/mcp.go
git commit -m "feat(mcp): add metrics tools (record, query)"
```

---

## Task 3: Create `core mcp serve` command

**Files:**
- Create: `internal/cmd/mcpcmd/cmd_mcp.go`
- Modify: `internal/variants/full.go` (add import)
- Test: Manual test via `core mcp serve`

**Step 1: Create the mcp command package**

Create `internal/cmd/mcpcmd/cmd_mcp.go`:

```go
package mcpcmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/mcp"
)

func init() {
	cli.RegisterCommands(AddMCPCommands)
}

var (
	mcpWorkspace string
)

var mcpCmd = &cli.Command{
	Use:   "mcp",
	Short: i18n.T("cmd.mcp.short"),
	Long:  i18n.T("cmd.mcp.long"),
}

var serveCmd = &cli.Command{
	Use:   "serve",
	Short: i18n.T("cmd.mcp.serve.short"),
	Long:  i18n.T("cmd.mcp.serve.long"),
	RunE: func(cmd *cli.Command, args []string) error {
		return runServe()
	},
}

func AddMCPCommands(root *cli.Command) {
	initMCPFlags()
	mcpCmd.AddCommand(serveCmd)
	root.AddCommand(mcpCmd)
}

func initMCPFlags() {
	serveCmd.Flags().StringVar(&mcpWorkspace, "workspace", "", i18n.T("cmd.mcp.serve.flag.workspace"))
}

func runServe() error {
	opts := []mcp.Option{}

	if mcpWorkspace != "" {
		opts = append(opts, mcp.WithWorkspaceRoot(mcpWorkspace))
	} else {
		// Default to unrestricted for MCP server
		opts = append(opts, mcp.WithWorkspaceRoot(""))
	}

	svc, err := mcp.New(opts...)
	if err != nil {
		return cli.Wrap(err, "create MCP service")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	return svc.Run(ctx)
}
```

**Step 2: Add i18n strings**

Create or update `pkg/i18n/en.yaml` (if it exists) or add to the existing i18n mechanism:

```yaml
cmd.mcp.short: "MCP (Model Context Protocol) server"
cmd.mcp.long: "Start an MCP server for Claude Code integration with file, RAG, and metrics tools."
cmd.mcp.serve.short: "Start the MCP server"
cmd.mcp.serve.long: "Start the MCP server in stdio mode. Use MCP_ADDR env var for TCP mode."
cmd.mcp.serve.flag.workspace: "Restrict file operations to this directory (empty = unrestricted)"
```

**Step 3: Add import to full.go**

Modify `internal/variants/full.go` to add:

```go
import (
	// ... existing imports ...
	_ "github.com/host-uk/core/internal/cmd/mcpcmd"
)
```

**Step 4: Build and test**

Run: `go build && ./core mcp serve --help`
Expected: Help output showing the serve command

**Step 5: Test MCP server manually**

Run: `echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./core mcp serve`
Expected: JSON response listing all tools including rag_query, metrics_record, etc.

**Step 6: Commit**

```bash
git add internal/cmd/mcpcmd/cmd_mcp.go internal/variants/full.go
git commit -m "feat: add 'core mcp serve' command"
```

---

## Task 4: Configure agentic-flows plugin with .mcp.json

**Files:**
- Create: `/home/shared/hostuk/claude-plugins/plugins/agentic-flows/.mcp.json`
- Modify: `/home/shared/hostuk/claude-plugins/plugins/agentic-flows/.claude-plugin/plugin.json` (optional, add mcpServers)

**Step 1: Create .mcp.json**

Create `/home/shared/hostuk/claude-plugins/plugins/agentic-flows/.mcp.json`:

```json
{
  "core-cli": {
    "command": "core",
    "args": ["mcp", "serve"],
    "env": {
      "MCP_WORKSPACE": ""
    }
  }
}
```

**Step 2: Verify plugin loads**

Restart Claude Code and run `/mcp` to verify the core-cli server appears.

**Step 3: Test MCP tools**

Test that tools are available:
- `mcp__plugin_agentic-flows_core-cli__rag_query`
- `mcp__plugin_agentic-flows_core-cli__rag_ingest`
- `mcp__plugin_agentic-flows_core-cli__rag_collections`
- `mcp__plugin_agentic-flows_core-cli__metrics_record`
- `mcp__plugin_agentic-flows_core-cli__metrics_query`
- `mcp__plugin_agentic-flows_core-cli__file_read`
- etc.

**Step 4: Commit plugin changes**

```bash
cd /home/shared/hostuk/claude-plugins
git add plugins/agentic-flows/.mcp.json
git commit -m "feat(agentic-flows): add MCP server configuration for core-cli"
```

---

## Task 5: Update documentation

**Files:**
- Modify: `/home/claude/.claude/projects/-home-claude/memory/MEMORY.md`
- Modify: `/home/claude/.claude/projects/-home-claude/memory/plugin-dev-notes.md`

**Step 1: Update MEMORY.md**

Add under "Core CLI MCP Server" section:

```markdown
### Core CLI MCP Server
- **Command:** `core mcp serve` (stdio mode) or `MCP_ADDR=:9000 core mcp serve` (TCP)
- **Tools available:**
  - File ops: file_read, file_write, file_edit, file_delete, file_rename, file_exists, dir_list, dir_create
  - RAG: rag_query, rag_ingest, rag_collections
  - Metrics: metrics_record, metrics_query
  - Language: lang_detect, lang_list
- **Plugin config:** `plugins/agentic-flows/.mcp.json`
```

**Step 2: Update plugin-dev-notes.md**

Add section:

```markdown
## MCP Server (core mcp serve)

### Available Tools
| Tool | Description |
|------|-------------|
| file_read | Read file contents |
| file_write | Write file contents |
| file_edit | Edit file (replace string) |
| file_delete | Delete file |
| file_rename | Rename/move file |
| file_exists | Check if file exists |
| dir_list | List directory contents |
| dir_create | Create directory |
| rag_query | Query vector DB |
| rag_ingest | Ingest file/directory |
| rag_collections | List collections |
| metrics_record | Record event |
| metrics_query | Query events |
| lang_detect | Detect file language |
| lang_list | List supported languages |

### Example .mcp.json
```json
{
  "core-cli": {
    "command": "core",
    "args": ["mcp", "serve"]
  }
}
```
```

**Step 3: Commit documentation**

```bash
git add ~/.claude/projects/-home-claude/memory/*.md
git commit -m "docs: update memory with MCP server tools"
```

---

## Summary

| Task | Files | Purpose |
|------|-------|---------|
| 1 | `pkg/mcp/tools_rag.go` | RAG tools (query, ingest, collections) |
| 2 | `pkg/mcp/tools_metrics.go` | Metrics tools (record, query) |
| 3 | `internal/cmd/mcpcmd/cmd_mcp.go` | `core mcp serve` command |
| 4 | `plugins/agentic-flows/.mcp.json` | Plugin MCP configuration |
| 5 | Memory docs | Documentation updates |

## Services Required

- **Qdrant:** localhost:6333 (verified running)
- **Ollama:** localhost:11434 with nomic-embed-text (verified running)
- **InfluxDB:** localhost:8086 (optional, for future time-series metrics)
