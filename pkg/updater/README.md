# Core Element Template

This repository is a template for developers to create custom HTML elements for the core web3 framework. It includes a Go backend, an Angular custom element, and a full release cycle configuration.

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/core-element-template.git
    ```

2.  **Install the dependencies:**
    ```bash
    cd core-element-template
    go mod tidy
    cd ui
    npm install
    ```

3.  **Run the development server:**
    ```bash
    go run ./cmd/demo-cli serve
    ```
    This will start the Go backend and serve the Angular custom element.

## Building the Custom Element

To build the Angular custom element, run the following command:

```bash
cd ui
npm run build
```

This will create a single JavaScript file in the `dist` directory that you can use in any HTML page.

## Usage

To use the updater library in your Go project, you can use the `UpdateService`.

### GitHub-based Updates

```go
package main

import (
	"fmt"
	"log"

	"github.com/snider/updater"
)

func main() {
	config := updater.UpdateServiceConfig{
		RepoURL:        "https://github.com/owner/repo",
		Channel:        "stable",
		CheckOnStartup: updater.CheckAndUpdateOnStartup,
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

### Generic HTTP Updates

For updates from a generic HTTP server, the server should provide a `latest.json` file at the root of the `RepoURL`. The JSON file should have the following structure:

```json
{
  "version": "1.2.3",
  "url": "https://your-server.com/path/to/release-asset"
}
```

You can then configure the `UpdateService` as follows:

```go
package main

import (
	"fmt"
	"log"

	"github.com/snider/updater"
)

func main() {
	config := updater.UpdateServiceConfig{
		RepoURL:        "https://your-server.com",
		CheckOnStartup: updater.CheckAndUpdateOnStartup,
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the EUPL-1.2 License - see the [LICENSE](LICENSE) file for details.
