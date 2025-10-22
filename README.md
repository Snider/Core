# Core

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A Go library starter project with core utilities and helper functions.

## Features

- 🚀 Ready-to-use Go library structure
- ✅ Comprehensive test coverage
- 📝 Well-documented code with examples
- 🔧 Makefile for common development tasks
- 🎯 GitHub Actions CI/CD pipeline
- 📦 Example usage included

## Installation

```bash
go get github.com/Snider/Core
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/Snider/Core"
)

func main() {
    // Get library version
    fmt.Println("Version:", core.GetVersion())
    
    // Use the Greeter
    greeter := core.NewGreeter("World")
    fmt.Println(greeter.Greet()) // Output: Hello, World!
    
    // Use utility functions
    sum := core.Add(10, 20)
    fmt.Println("Sum:", sum) // Output: Sum: 30
    
    product := core.Multiply(5, 6)
    fmt.Println("Product:", product) // Output: Product: 30
}
```

### Running the Example

```bash
make example
```

Or manually:

```bash
cd examples
go run main.go
```

## Development

### Prerequisites

- Go 1.19 or higher

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
│   └── main.go             # Example usage
├── .gitignore              # Git ignore rules
├── core.go                 # Main library code
├── core_test.go            # Tests
├── go.mod                  # Go module definition
├── LICENSE                 # License file
├── Makefile                # Development tasks
└── README.md               # This file
```

## API Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/Snider/Core).

You can also generate documentation locally:

```bash
go doc -all
```

### Available Functions

- `GetVersion() string` - Returns the library version
- `NewGreeter(name string) *Greeter` - Creates a new Greeter instance
- `Greeter.Greet() string` - Returns a greeting message
- `Add(a, b int) int` - Adds two integers
- `Multiply(a, b int) int` - Multiplies two integers

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

## Acknowledgments

- Built with ❤️ using Go
- Inspired by Go best practices and community standards