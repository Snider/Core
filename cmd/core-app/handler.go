package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dunglas/frankenphp"
)

// PHPHandler implements http.Handler by delegating to FrankenPHP.
// It resolves URLs to files (like Caddy's try_files) before passing
// requests to the PHP runtime.
type PHPHandler struct {
	docRoot     string
	laravelRoot string
}

// NewPHPHandler extracts the embedded Laravel app, prepares the environment,
// initialises FrankenPHP with worker mode, and returns the handler.
func NewPHPHandler() (*PHPHandler, *AppEnvironment, func(), error) {
	// Extract embedded Laravel to temp directory
	laravelRoot, err := extractLaravel()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("extract Laravel: %w", err)
	}

	// Prepare persistent environment
	env, err := PrepareEnvironment(laravelRoot)
	if err != nil {
		os.RemoveAll(laravelRoot)
		return nil, nil, nil, fmt.Errorf("prepare environment: %w", err)
	}

	docRoot := filepath.Join(laravelRoot, "public")

	log.Printf("Laravel root: %s", laravelRoot)
	log.Printf("Document root: %s", docRoot)
	log.Printf("Data directory: %s", env.DataDir)
	log.Printf("Database: %s", env.DatabasePath)

	// Try Octane worker mode first, fall back to standard mode.
	// Worker mode keeps Laravel booted in memory — sub-ms response times.
	workerScript := filepath.Join(laravelRoot, "vendor", "laravel", "octane", "bin", "frankenphp-worker.php")
	workerEnv := map[string]string{
		"APP_BASE_PATH":     laravelRoot,
		"FRANKENPHP_WORKER": "1",
	}

	workerMode := false
	if _, err := os.Stat(workerScript); err == nil {
		if err := frankenphp.Init(
			frankenphp.WithNumThreads(4),
			frankenphp.WithWorkers("laravel", workerScript, 2, workerEnv, nil),
			frankenphp.WithPhpIni(map[string]string{
				"display_errors": "Off",
				"opcache.enable": "1",
			}),
		); err != nil {
			log.Printf("Worker mode init failed (%v), falling back to standard mode", err)
		} else {
			workerMode = true
		}
	}

	if !workerMode {
		if err := frankenphp.Init(
			frankenphp.WithNumThreads(4),
			frankenphp.WithPhpIni(map[string]string{
				"display_errors": "Off",
				"opcache.enable": "1",
			}),
		); err != nil {
			os.RemoveAll(laravelRoot)
			return nil, nil, nil, fmt.Errorf("init FrankenPHP: %w", err)
		}
	}

	if workerMode {
		log.Println("FrankenPHP initialised (Octane worker mode, 2 workers)")
	} else {
		log.Println("FrankenPHP initialised (standard mode, 4 threads)")
	}

	cleanup := func() {
		frankenphp.Shutdown()
		os.RemoveAll(laravelRoot)
	}

	handler := &PHPHandler{
		docRoot:     docRoot,
		laravelRoot: laravelRoot,
	}

	return handler, env, cleanup, nil
}

func (h *PHPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	filePath := filepath.Join(h.docRoot, filepath.Clean(urlPath))

	info, err := os.Stat(filePath)
	if err == nil && info.IsDir() {
		// Directory → try index.php inside it
		urlPath = strings.TrimRight(urlPath, "/") + "/index.php"
	} else if err != nil && !strings.HasSuffix(urlPath, ".php") {
		// File not found and not a .php request → front controller
		urlPath = "/index.php"
	}

	// Serve static assets directly (CSS, JS, images)
	if !strings.HasSuffix(urlPath, ".php") {
		staticPath := filepath.Join(h.docRoot, filepath.Clean(urlPath))
		if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, staticPath)
			return
		}
	}

	// Route to FrankenPHP
	r.URL.Path = urlPath

	req, err := frankenphp.NewRequestWithContext(r,
		frankenphp.WithRequestDocumentRoot(h.docRoot, false),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("FrankenPHP request error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := frankenphp.ServeHTTP(w, req); err != nil {
		http.Error(w, fmt.Sprintf("FrankenPHP serve error: %v", err), http.StatusInternalServerError)
	}
}
