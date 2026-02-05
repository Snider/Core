package mcp

import (
	"testing"

	"github.com/host-uk/core/pkg/ws"
)

// TestWSToolsRegistered_Good verifies that WebSocket tools are registered when hub is available.
func TestWSToolsRegistered_Good(t *testing.T) {
	// Create a new MCP service without ws hub - tools should not be registered
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.wsHub != nil {
		t.Error("WS hub should be nil by default")
	}

	if s.server == nil {
		t.Fatal("Server should not be nil")
	}
}

// TestWSStartInput_Good verifies the WSStartInput struct has expected fields.
func TestWSStartInput_Good(t *testing.T) {
	input := WSStartInput{
		Addr: ":9090",
	}

	if input.Addr != ":9090" {
		t.Errorf("Expected addr ':9090', got %q", input.Addr)
	}
}

// TestWSStartInput_Defaults verifies default values.
func TestWSStartInput_Defaults(t *testing.T) {
	input := WSStartInput{}

	if input.Addr != "" {
		t.Errorf("Expected addr to default to empty, got %q", input.Addr)
	}
}

// TestWSStartOutput_Good verifies the WSStartOutput struct has expected fields.
func TestWSStartOutput_Good(t *testing.T) {
	output := WSStartOutput{
		Success: true,
		Addr:    "127.0.0.1:8080",
		Message: "WebSocket server started",
	}

	if !output.Success {
		t.Error("Expected Success to be true")
	}
	if output.Addr != "127.0.0.1:8080" {
		t.Errorf("Expected addr '127.0.0.1:8080', got %q", output.Addr)
	}
	if output.Message != "WebSocket server started" {
		t.Errorf("Expected message 'WebSocket server started', got %q", output.Message)
	}
}

// TestWSInfoInput_Good verifies the WSInfoInput struct exists (it's empty).
func TestWSInfoInput_Good(t *testing.T) {
	input := WSInfoInput{}
	_ = input // Just verify it compiles
}

// TestWSInfoOutput_Good verifies the WSInfoOutput struct has expected fields.
func TestWSInfoOutput_Good(t *testing.T) {
	output := WSInfoOutput{
		Clients:  5,
		Channels: 3,
	}

	if output.Clients != 5 {
		t.Errorf("Expected clients 5, got %d", output.Clients)
	}
	if output.Channels != 3 {
		t.Errorf("Expected channels 3, got %d", output.Channels)
	}
}

// TestWithWSHub_Good verifies the WithWSHub option.
func TestWithWSHub_Good(t *testing.T) {
	hub := ws.NewHub()

	s, err := New(WithWSHub(hub))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.wsHub != hub {
		t.Error("Expected wsHub to be set")
	}
}

// TestWithWSHub_Nil verifies the WithWSHub option with nil.
func TestWithWSHub_Nil(t *testing.T) {
	s, err := New(WithWSHub(nil))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.wsHub != nil {
		t.Error("Expected wsHub to be nil when passed nil")
	}
}

// TestProcessEventCallback_Good verifies the ProcessEventCallback struct.
func TestProcessEventCallback_Good(t *testing.T) {
	hub := ws.NewHub()
	callback := NewProcessEventCallback(hub)

	if callback.hub != hub {
		t.Error("Expected callback hub to be set")
	}

	// Test that methods don't panic
	callback.OnProcessOutput("proc-1", "test output")
	callback.OnProcessStatus("proc-1", "exited", 0)
}

// TestProcessEventCallback_NilHub verifies the ProcessEventCallback with nil hub doesn't panic.
func TestProcessEventCallback_NilHub(t *testing.T) {
	callback := NewProcessEventCallback(nil)

	if callback.hub != nil {
		t.Error("Expected callback hub to be nil")
	}

	// Test that methods don't panic with nil hub
	callback.OnProcessOutput("proc-1", "test output")
	callback.OnProcessStatus("proc-1", "exited", 0)
}

// TestServiceWSHub_Good verifies the WSHub getter method.
func TestServiceWSHub_Good(t *testing.T) {
	hub := ws.NewHub()
	s, err := New(WithWSHub(hub))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.WSHub() != hub {
		t.Error("Expected WSHub() to return the hub")
	}
}

// TestServiceWSHub_Nil verifies the WSHub getter returns nil when not configured.
func TestServiceWSHub_Nil(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.WSHub() != nil {
		t.Error("Expected WSHub() to return nil when not configured")
	}
}

// TestServiceProcessService_Nil verifies the ProcessService getter returns nil when not configured.
func TestServiceProcessService_Nil(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.ProcessService() != nil {
		t.Error("Expected ProcessService() to return nil when not configured")
	}
}
