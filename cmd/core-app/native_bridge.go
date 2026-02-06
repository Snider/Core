package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

// NativeBridge provides a localhost HTTP API that PHP code can call
// to access native desktop capabilities (file dialogs, notifications, etc.).
//
// Livewire renders server-side in PHP, so it can't call Wails bindings
// (window.go.*) directly. Instead, PHP makes HTTP requests to this bridge.
// The bridge port is injected into Laravel's .env as NATIVE_BRIDGE_URL.
type NativeBridge struct {
	server *http.Server
	port   int
	app    *AppService
}

// NewNativeBridge creates and starts the bridge on a random available port.
func NewNativeBridge(appService *AppService) (*NativeBridge, error) {
	mux := http.NewServeMux()
	bridge := &NativeBridge{app: appService}

	// Register bridge endpoints
	mux.HandleFunc("POST /bridge/version", bridge.handleVersion)
	mux.HandleFunc("POST /bridge/data-dir", bridge.handleDataDir)
	mux.HandleFunc("POST /bridge/show-window", bridge.handleShowWindow)
	mux.HandleFunc("GET /bridge/health", bridge.handleHealth)

	// Listen on a random available port (localhost only)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	bridge.port = listener.Addr().(*net.TCPAddr).Port
	bridge.server = &http.Server{Handler: mux}

	go func() {
		if err := bridge.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Native bridge error: %v", err)
		}
	}()

	log.Printf("Native bridge listening on http://127.0.0.1:%d", bridge.port)
	return bridge, nil
}

// Port returns the port the bridge is listening on.
func (b *NativeBridge) Port() int {
	return b.port
}

// URL returns the full base URL of the bridge.
func (b *NativeBridge) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", b.port)
}

// Shutdown gracefully stops the bridge server.
func (b *NativeBridge) Shutdown(ctx context.Context) error {
	return b.server.Shutdown(ctx)
}

func (b *NativeBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (b *NativeBridge) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"version": b.app.GetVersion()})
}

func (b *NativeBridge) handleDataDir(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"path": b.app.GetDataDir()})
}

func (b *NativeBridge) handleShowWindow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.app.ShowWindow(req.Name)
	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
