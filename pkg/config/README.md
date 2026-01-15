# Config Module

[![Go CI](https://github.com/Snider/config/actions/workflows/ci.yml/badge.svg)](https://github.com/Snider/config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/Snider/config/branch/main/graph/badge.svg)](https://codecov.io/gh/Snider/config)

This repository is a config module for the Core Framework. It includes a Go backend, an Angular custom element, and a full release cycle configuration.

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/Snider/config.git
    ```

2.  **Install the dependencies:**
    ```bash
    cd config
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

The `config` service provides a generic way to load and save configuration files in various formats. This is useful for other packages that need to persist their own configuration without being tied to a specific format.

### Supported Formats

*   JSON (`.json`)
*   YAML (`.yml`, `.yaml`)
*   INI (`.ini`)
*   XML (`.xml`)

### Saving Configuration

To save a set of key-value pairs, you can use the `SaveKeyValues` method. The format is determined by the file extension of the `key` you provide.

```go
package main

import (
	"log"

	"github.com/Snider/config/pkg/config"
)

func main() {
	// Get a new config service instance
	configSvc, err := config.New()
	if err != nil {
		log.Fatalf("Failed to create config service: %v", err)
	}

	// Example data to save
	data := map[string]interface{}{
		"setting1": "value1",
		"enabled":  true,
		"retries":  3,
	}

	// Save as a JSON file
	if err := configSvc.SaveKeyValues("my-app-settings.json", data); err != nil {
		log.Fatalf("Failed to save JSON config: %v", err)
	}

	// Save as a YAML file
	if err := configSvc.SaveKeyValues("my-app-settings.yaml", data); err != nil {
		log.Fatalf("Failed to save YAML config: %v", err)
	}

	// For INI, keys are typically in `section.key` format
	iniData := map[string]interface{}{
		"general.setting1": "value1",
		"general.enabled":  true,
		"network.retries":  3,
	}
	if err := configSvc.SaveKeyValues("my-app-settings.ini", iniData); err != nil {
		log.Fatalf("Failed to save INI config: %v", err)
	}

	// Save as an XML file
	if err := configSvc.SaveKeyValues("my-app-settings.xml", data); err != nil {
		log.Fatalf("Failed to save XML config: %v", err)
	}
}
```

### Loading Configuration

To load a configuration file, use the `LoadKeyValues` method. It will automatically parse the file based on its extension and return a `map[string]interface{}`.

```go
package main

import (
	"fmt"
	"log"

	"github.com/Snider/config/pkg/config"
)

func main() {
	// Get a new config service instance
	configSvc, err := config.New()
	if err != nil {
		log.Fatalf("Failed to create config service: %v", err)
	}

	// Load a JSON file
	jsonData, err := configSvc.LoadKeyValues("my-app-settings.json")
	if err != nil {
		log.Fatalf("Failed to load JSON config: %v", err)
	}
	fmt.Printf("Loaded from JSON: %v\n", jsonData)
	// Note: Numbers from JSON are unmarshaled as float64

	// Load a YAML file
	yamlData, err := configSvc.LoadKeyValues("my-app-settings.yaml")
	if err != nil {
		log.Fatalf("Failed to load YAML config: %v", err)
	}
	fmt.Printf("Loaded from YAML: %v\n", yamlData)
	// Note: Numbers from YAML without decimals are unmarshaled as int

	// Load an INI file
	iniData, err := configSvc.LoadKeyValues("my-app-settings.ini")
	if err != nil {
		log.Fatalf("Failed to load INI config: %v", err)
	}
	fmt.Printf("Loaded from INI: %v\n", iniData)
	// Note: All values from INI are loaded as strings

	// Load an XML file
	xmlData, err := configSvc.LoadKeyValues("my-app-settings.xml")
	if err != nil {
		log.Fatalf("Failed to load XML config: %v", err)
	}
	fmt.Printf("Loaded from XML: %v\n", xmlData)
	// Note: All values from XML are loaded as strings
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the EUPL-1.2 License - see the [LICENSE](LICENSE) file for details.
