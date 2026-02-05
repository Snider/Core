package mcp

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewTCPTransport_Defaults(t *testing.T) {
	// Test that empty string gets replaced with default address constant
	// Note: We can't actually bind to 9100 as it may be in use,
	// so we verify the address is set correctly before Listen is called
	if DefaultTCPAddr != "127.0.0.1:9100" {
		t.Errorf("Expected default constant 127.0.0.1:9100, got %s", DefaultTCPAddr)
	}

	// Test with a dynamic port to verify transport creation works
	tr, err := NewTCPTransport("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create transport with dynamic port: %v", err)
	}
	defer tr.listener.Close()

	// Verify we got a valid address
	if tr.addr != "127.0.0.1:0" {
		t.Errorf("Expected address to be set, got %s", tr.addr)
	}
}

func TestNewTCPTransport_Warning(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	// Trigger warning
	tr, err := NewTCPTransport("0.0.0.0:9101")
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}
	defer tr.listener.Close()

	// Restore stderr
	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "WARNING") {
		t.Error("Expected warning for binding to 0.0.0.0, but didn't find it in stderr")
	}
}

func TestServeTCP_Connection(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a random port for testing to avoid collisions
	addr := "127.0.0.1:0"

	// Create transport first to get the actual address if we use :0
	tr, err := NewTCPTransport(addr)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}
	actualAddr := tr.listener.Addr().String()
	tr.listener.Close() // Close it so ServeTCP can re-open it or use the same address

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ServeTCP(ctx, actualAddr)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", actualAddr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Verify we can write to it
	_, err = conn.Write([]byte("{}\n"))
	if err != nil {
		t.Errorf("Failed to write to connection: %v", err)
	}

	// Shutdown server
	cancel()
	err = <-errCh
	if err != nil {
		t.Errorf("ServeTCP returned error: %v", err)
	}
}

func TestRun_TCPTrigger(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set MCP_ADDR to empty to trigger default TCP
	os.Setenv("MCP_ADDR", "")
	defer os.Unsetenv("MCP_ADDR")

	// We use a random port for testing, but Run will try to use 127.0.0.1:9100 by default if we don't override.
	// Since 9100 might be in use, we'll set MCP_ADDR to use :0 (random port)
	os.Setenv("MCP_ADDR", "127.0.0.1:0")

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Run(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Since we can't easily get the actual port used by Run (it's internal),
	// we just verify it didn't immediately fail.
	select {
	case err := <-errCh:
		t.Fatalf("Run failed immediately: %v", err)
	default:
		// still running, which is good
	}

	cancel()
	_ = <-errCh
}

func TestServeTCP_MultipleConnections(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "127.0.0.1:0"
	tr, err := NewTCPTransport(addr)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}
	actualAddr := tr.listener.Addr().String()
	tr.listener.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ServeTCP(ctx, actualAddr)
	}()

	time.Sleep(100 * time.Millisecond)

	// Connect multiple clients
	const numClients = 3
	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", actualAddr)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		defer conn.Close()
		_, err = conn.Write([]byte("{}\n"))
		if err != nil {
			t.Errorf("Client %d failed to write: %v", i, err)
		}
	}

	cancel()
	err = <-errCh
	if err != nil {
		t.Errorf("ServeTCP returned error: %v", err)
	}
}
