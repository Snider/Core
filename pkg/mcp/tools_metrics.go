package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Default values for metrics operations.
const (
	DefaultMetricsSince = "7d"
	DefaultMetricsLimit = 10
)

// MetricsRecordInput contains parameters for recording a metrics event.
type MetricsRecordInput struct {
	Type    string         `json:"type"`               // Event type (required)
	AgentID string         `json:"agent_id,omitempty"` // Agent identifier
	Repo    string         `json:"repo,omitempty"`     // Repository name
	Data    map[string]any `json:"data,omitempty"`     // Additional event data
}

// MetricsRecordOutput contains the result of recording a metrics event.
type MetricsRecordOutput struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricsQueryInput contains parameters for querying metrics.
type MetricsQueryInput struct {
	Since string `json:"since,omitempty"` // Time range like "7d", "24h", "30m" (default: "7d")
}

// MetricsQueryOutput contains the results of a metrics query.
type MetricsQueryOutput struct {
	Total   int                `json:"total"`
	ByType  []MetricCount      `json:"by_type"`
	ByRepo  []MetricCount      `json:"by_repo"`
	ByAgent []MetricCount      `json:"by_agent"`
	Events  []MetricEventBrief `json:"events"` // Most recent 10 events
}

// MetricCount represents a count for a specific key.
type MetricCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// MetricEventBrief represents a brief summary of an event.
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
		Description: "Record a metrics event for AI/security tracking. Events are stored in daily JSONL files.",
	}, s.metricsRecord)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "metrics_query",
		Description: "Query metrics events and get aggregated statistics by type, repo, and agent.",
	}, s.metricsQuery)
}

// metricsRecord handles the metrics_record tool call.
func (s *Service) metricsRecord(ctx context.Context, req *mcp.CallToolRequest, input MetricsRecordInput) (*mcp.CallToolResult, MetricsRecordOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "metrics_record", "type", input.Type, "agent_id", input.AgentID, "repo", input.Repo, "user", log.Username())

	// Validate input
	if input.Type == "" {
		return nil, MetricsRecordOutput{}, fmt.Errorf("type cannot be empty")
	}

	// Create the event
	event := ai.Event{
		Type:      input.Type,
		Timestamp: time.Now(),
		AgentID:   input.AgentID,
		Repo:      input.Repo,
		Data:      input.Data,
	}

	// Record the event
	if err := ai.Record(event); err != nil {
		log.Error("mcp: metrics record failed", "type", input.Type, "err", err)
		return nil, MetricsRecordOutput{}, fmt.Errorf("failed to record metrics: %w", err)
	}

	return nil, MetricsRecordOutput{
		Success:   true,
		Timestamp: event.Timestamp,
	}, nil
}

// metricsQuery handles the metrics_query tool call.
func (s *Service) metricsQuery(ctx context.Context, req *mcp.CallToolRequest, input MetricsQueryInput) (*mcp.CallToolResult, MetricsQueryOutput, error) {
	// Apply defaults
	since := input.Since
	if since == "" {
		since = DefaultMetricsSince
	}

	s.logger.Info("MCP tool execution", "tool", "metrics_query", "since", since, "user", log.Username())

	// Parse the duration
	duration, err := parseDuration(since)
	if err != nil {
		return nil, MetricsQueryOutput{}, fmt.Errorf("invalid since value: %w", err)
	}

	sinceTime := time.Now().Add(-duration)

	// Read events
	events, err := ai.ReadEvents(sinceTime)
	if err != nil {
		log.Error("mcp: metrics query failed", "since", since, "err", err)
		return nil, MetricsQueryOutput{}, fmt.Errorf("failed to read metrics: %w", err)
	}

	// Get summary
	summary := ai.Summary(events)

	// Build output
	output := MetricsQueryOutput{
		Total:   summary["total"].(int),
		ByType:  convertMetricCounts(summary["by_type"]),
		ByRepo:  convertMetricCounts(summary["by_repo"]),
		ByAgent: convertMetricCounts(summary["by_agent"]),
		Events:  make([]MetricEventBrief, 0, DefaultMetricsLimit),
	}

	// Get recent events (last 10, most recent first)
	startIdx := len(events) - DefaultMetricsLimit
	if startIdx < 0 {
		startIdx = 0
	}
	for i := len(events) - 1; i >= startIdx; i-- {
		ev := events[i]
		output.Events = append(output.Events, MetricEventBrief{
			Type:      ev.Type,
			Timestamp: ev.Timestamp,
			AgentID:   ev.AgentID,
			Repo:      ev.Repo,
		})
	}

	return nil, output, nil
}

// convertMetricCounts converts the summary map format to MetricCount slice.
func convertMetricCounts(data any) []MetricCount {
	if data == nil {
		return []MetricCount{}
	}

	items, ok := data.([]map[string]any)
	if !ok {
		return []MetricCount{}
	}

	result := make([]MetricCount, len(items))
	for i, item := range items {
		key, _ := item["key"].(string)
		count, _ := item["count"].(int)
		result[i] = MetricCount{Key: key, Count: count}
	}
	return result
}

// parseDuration parses a duration string like "7d", "24h", "30m".
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format: %q", s)
	}

	// Get the numeric part and unit
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration number: %q", numStr)
	}

	if num <= 0 {
		return 0, fmt.Errorf("duration must be positive: %d", num)
	}

	switch unit {
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	case 'm':
		return time.Duration(num) * time.Minute, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %q (expected d, h, or m)", string(unit))
	}
}
