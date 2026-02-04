//go:build ide

// core_ide.go imports packages for the Core IDE desktop application.
//
// Build with: go build -tags ide
//
// This is the Wails v3 GUI variant featuring:
//   - System tray with quick actions
//   - Tray panel for status/notifications
//   - Angular frontend
//   - All CLI commands available via IPC

package variants

import (
	// Core IDE GUI
	_ "github.com/host-uk/core/internal/core-ide"

	// CLI commands available via IPC
	_ "github.com/host-uk/core/internal/cmd/ai"
	_ "github.com/host-uk/core/internal/cmd/dev"
	_ "github.com/host-uk/core/internal/cmd/deploy"
	_ "github.com/host-uk/core/internal/cmd/php"
	_ "github.com/host-uk/core/internal/cmd/rag"
)
