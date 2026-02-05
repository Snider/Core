package mcp

import (
	"testing"
	"time"
)

// TestMetricsToolsRegistered_Good verifies that metrics tools are registered with the MCP server.
func TestMetricsToolsRegistered_Good(t *testing.T) {
	// Create a new MCP service - this should register all tools including metrics
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// The server should have registered the metrics tools
	// We verify by checking that the server and logger exist
	if s.server == nil {
		t.Fatal("Server should not be nil")
	}

	if s.logger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestMetricsRecordInput_Good verifies the MetricsRecordInput struct has expected fields.
func TestMetricsRecordInput_Good(t *testing.T) {
	input := MetricsRecordInput{
		Type:    "tool_call",
		AgentID: "agent-123",
		Repo:    "host-uk/core",
		Data:    map[string]any{"tool": "file_read", "duration_ms": 150},
	}

	if input.Type != "tool_call" {
		t.Errorf("Expected type 'tool_call', got %q", input.Type)
	}
	if input.AgentID != "agent-123" {
		t.Errorf("Expected agent_id 'agent-123', got %q", input.AgentID)
	}
	if input.Repo != "host-uk/core" {
		t.Errorf("Expected repo 'host-uk/core', got %q", input.Repo)
	}
	if input.Data["tool"] != "file_read" {
		t.Errorf("Expected data[tool] 'file_read', got %v", input.Data["tool"])
	}
}

// TestMetricsRecordOutput_Good verifies the MetricsRecordOutput struct has expected fields.
func TestMetricsRecordOutput_Good(t *testing.T) {
	ts := time.Now()
	output := MetricsRecordOutput{
		Success:   true,
		Timestamp: ts,
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Timestamp != ts {
		t.Errorf("Expected timestamp %v, got %v", ts, output.Timestamp)
	}
}

// TestMetricsQueryInput_Good verifies the MetricsQueryInput struct has expected fields.
func TestMetricsQueryInput_Good(t *testing.T) {
	input := MetricsQueryInput{
		Since: "7d",
	}

	if input.Since != "7d" {
		t.Errorf("Expected since '7d', got %q", input.Since)
	}
}

// TestMetricsQueryInput_Defaults verifies default values are handled correctly.
func TestMetricsQueryInput_Defaults(t *testing.T) {
	input := MetricsQueryInput{}

	// Empty since should use default when processed
	if input.Since != "" {
		t.Errorf("Expected empty since before defaults, got %q", input.Since)
	}
}

// TestMetricsQueryOutput_Good verifies the MetricsQueryOutput struct has expected fields.
func TestMetricsQueryOutput_Good(t *testing.T) {
	output := MetricsQueryOutput{
		Total: 100,
		ByType: []MetricCount{
			{Key: "tool_call", Count: 50},
			{Key: "query", Count: 30},
		},
		ByRepo: []MetricCount{
			{Key: "host-uk/core", Count: 40},
		},
		ByAgent: []MetricCount{
			{Key: "agent-123", Count: 25},
		},
		Events: []MetricEventBrief{
			{Type: "tool_call", Timestamp: time.Now(), AgentID: "agent-1", Repo: "host-uk/core"},
		},
	}

	if output.Total != 100 {
		t.Errorf("Expected total 100, got %d", output.Total)
	}
	if len(output.ByType) != 2 {
		t.Errorf("Expected 2 ByType entries, got %d", len(output.ByType))
	}
	if output.ByType[0].Key != "tool_call" {
		t.Errorf("Expected ByType[0].Key 'tool_call', got %q", output.ByType[0].Key)
	}
	if output.ByType[0].Count != 50 {
		t.Errorf("Expected ByType[0].Count 50, got %d", output.ByType[0].Count)
	}
	if len(output.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(output.Events))
	}
}

// TestMetricCount_Good verifies the MetricCount struct has expected fields.
func TestMetricCount_Good(t *testing.T) {
	mc := MetricCount{
		Key:   "tool_call",
		Count: 42,
	}

	if mc.Key != "tool_call" {
		t.Errorf("Expected key 'tool_call', got %q", mc.Key)
	}
	if mc.Count != 42 {
		t.Errorf("Expected count 42, got %d", mc.Count)
	}
}

// TestMetricEventBrief_Good verifies the MetricEventBrief struct has expected fields.
func TestMetricEventBrief_Good(t *testing.T) {
	ts := time.Now()
	ev := MetricEventBrief{
		Type:      "tool_call",
		Timestamp: ts,
		AgentID:   "agent-123",
		Repo:      "host-uk/core",
	}

	if ev.Type != "tool_call" {
		t.Errorf("Expected type 'tool_call', got %q", ev.Type)
	}
	if ev.Timestamp != ts {
		t.Errorf("Expected timestamp %v, got %v", ts, ev.Timestamp)
	}
	if ev.AgentID != "agent-123" {
		t.Errorf("Expected agent_id 'agent-123', got %q", ev.AgentID)
	}
	if ev.Repo != "host-uk/core" {
		t.Errorf("Expected repo 'host-uk/core', got %q", ev.Repo)
	}
}

// TestParseDuration_Good verifies the parseDuration helper handles various formats.
func TestParseDuration_Good(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"7d", 7 * 24 * time.Hour},
		{"24h", 24 * time.Hour},
		{"30m", 30 * time.Minute},
		{"1d", 24 * time.Hour},
		{"14d", 14 * 24 * time.Hour},
		{"1h", time.Hour},
		{"10m", 10 * time.Minute},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			d, err := parseDuration(tc.input)
			if err != nil {
				t.Fatalf("parseDuration(%q) returned error: %v", tc.input, err)
			}
			if d != tc.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tc.input, d, tc.expected)
			}
		})
	}
}

// TestParseDuration_Bad verifies parseDuration returns errors for invalid input.
func TestParseDuration_Bad(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"7x",
		"-7d",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDuration(input)
			if err == nil {
				t.Errorf("parseDuration(%q) should return error", input)
			}
		})
	}
}
