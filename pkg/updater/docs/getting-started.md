# Getting Started

This guide will help you integrate the `updater` library into your Go application.

## Installation

To install the library, run:

```bash
go get github.com/snider/updater
```

## Basic Usage

The `updater` library provides an `UpdateService` that simplifies the process of checking for and applying updates.

### GitHub-based Updates

If you are hosting your releases on GitHub, you can configure the service to check your repository.

```go
package main

import (
	"fmt"
	"log"

	"github.com/snider/updater"
)

func main() {
	// Configure the update service
	config := updater.UpdateServiceConfig{
		RepoURL:        "https://github.com/your-username/your-repo",
		Channel:        "stable", // or "beta", "alpha", etc.
		CheckOnStartup: updater.CheckAndUpdateOnStartup,
	}

	// Create the service
	updateService, err := updater.NewUpdateService(config)
	if err != nil {
		log.Fatalf("Failed to create update service: %v", err)
	}

	// Start the service (checks for updates and applies them if configured)
	if err := updateService.Start(); err != nil {
		fmt.Printf("Update check/apply failed: %v\n", err)
	} else {
		fmt.Println("Update check completed.")
	}
}
```

### Generic HTTP Updates

If you are hosting your releases on a generic HTTP server, the server must provide a way to check for the latest version.

```go
package main

import (
	"fmt"
	"log"

	"github.com/snider/updater"
)

func main() {
	config := updater.UpdateServiceConfig{
		RepoURL:        "https://your-server.com/updates",
		CheckOnStartup: updater.CheckOnStartup, // Check only, don't apply automatically
	}

	updateService, err := updater.NewUpdateService(config)
	if err != nil {
		log.Fatalf("Failed to create update service: %v", err)
	}

	if err := updateService.Start(); err != nil {
		fmt.Printf("Update check failed: %v\n", err)
	}
}
```

For Generic HTTP updates, the endpoint is expected to return a JSON object with `version` and `url` fields. See [Architecture](architecture.md) for more details.
