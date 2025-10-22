# Core

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v3.0.0--alpha.36-blue)](https://wails.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A Wails v3 service library starter project providing core utilities and helper functions. This library can be integrated into any Wails v3 application to provide backend functionality accessible from the frontend.

## Features

- 🚀 Ready-to-use Wails v3 service structure
- ✅ Comprehensive test coverage
- 📝 Well-documented code with examples
- 🔧 Makefile for common development tasks
- 🎯 GitHub Actions CI/CD pipeline
- 📦 Example Wails application included
- 🌐 Interactive frontend demo

## What is a Wails v3 Service?

Wails v3 services are Go packages that can be registered with a Wails application and called from the frontend JavaScript/TypeScript code. This allows you to write backend logic in Go and expose it to your web-based UI seamlessly.

## Installation

```bash
go get github.com/Snider/Core
```

### Prerequisites

- Go 1.19 or higher
- Wails v3 CLI (for running the example):
  ```bash
  go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.36
  ```

## Usage

### Integrating into Your Wails Application

```go
package main

import (
    "github.com/Snider/Core"
    "github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
    // Create the Core service
    coreService := core.NewCoreService()

    // Create Wails application with the service
    app := application.New(application.Options{
        Name:        "My App",
        Description: "My Wails Application",
        Services: []application.Service{
            application.NewService(coreService),
        },
    })

    // Create and run your application...
    app.Run()
}
```

### Calling from Frontend

Once registered, you can call the service methods from your frontend JavaScript:

```javascript
// Get version
const version = await wails.Call.ByName('github.com.Snider.Core.GetVersion');

// Greet user
const greeting = await wails.Call.ByName('github.com.Snider.Core.Greet', 'World');
console.log(greeting); // "Hello, World!"

// Perform calculations
const sum = await wails.Call.ByName('github.com.Snider.Core.Add', 10, 20);
const product = await wails.Call.ByName('github.com.Snider.Core.Multiply', 5, 6);

// Complex calculations with error handling
try {
    const result = await wails.Call.ByName('github.com.Snider.Core.Calculate', 'divide', 100, 4);
    console.log(result); // 25
} catch (error) {
    console.error('Calculation error:', error);
}
```

## Available Service Methods

The `CoreService` provides the following methods:

- **`ServiceName() string`** - Returns the service identifier
- **`GetVersion() string`** - Returns the current version
- **`Greet(name string) string`** - Returns a greeting message
- **`Add(a, b int) int`** - Adds two integers
- **`Multiply(a, b int) int`** - Multiplies two integers
- **`Calculate(operation string, a, b int) (int, error)`** - Performs calculations
  - Supported operations: `"add"`, `"subtract"`, `"multiply"`, `"divide"`

## Running the Example

The example demonstrates a complete Wails application using the Core service:

```bash
cd examples
go run main.go
```

This will launch a Wails application window with an interactive UI that calls the Core service methods.

## Development

### Building

```bash
make build
```

### Testing

Run all tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-coverage
```

Run benchmark tests:

```bash
make test-bench
```

### Code Quality

Format code:

```bash
make fmt
```

Run static analysis:

```bash
make vet
```

Run linter (requires golangci-lint):

```bash
make lint
```

### Installing Development Tools

```bash
make install-tools
```

## Project Structure

```
.
├── .github/
│   └── workflows/
│       └── ci.yml          # GitHub Actions CI/CD
├── examples/
│   ├── frontend/
│   │   └── index.html      # Example UI
│   └── main.go             # Example Wails application
├── .gitignore              # Git ignore rules
├── .golangci.yml           # Linter configuration
├── core.go                 # Core service implementation
├── core_test.go            # Tests
├── go.mod                  # Go module definition
├── LICENSE                 # MIT License
├── Makefile                # Development tasks
└── README.md               # This file
```

## Creating Your Own Service

Use this project as a template to create your own Wails v3 service:

1. Fork or clone this repository
2. Update the module name in `go.mod`
3. Modify `core.go` to implement your service logic
4. Update the `ServiceName()` method to return your service identifier
5. Add your methods (they will be automatically exposed to the frontend)
6. Write tests for your methods
7. Update the example to demonstrate your service

### Service Method Requirements

For methods to be callable from the frontend:

- Must be exported (start with uppercase letter)
- Must belong to a struct that implements the service interface
- Can accept basic types and structs as parameters
- Can return values and errors

## API Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/Snider/Core).

Generate documentation locally:

```bash
go doc -all
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Testing Guidelines

- Write tests for all new functionality
- Maintain or improve code coverage
- Run `make test` before submitting PRs
- Include benchmark tests for performance-critical code

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Resources

- [Wails v3 Documentation](https://wails.io)
- [Wails v3 GitHub Repository](https://github.com/wailsapp/wails)
- [Go Documentation](https://golang.org/doc/)

## Acknowledgments

- Built with ❤️ using Go and Wails v3
- Inspired by Wails best practices and community standards