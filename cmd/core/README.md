# Core CLI

This directory contains the source code for the `core` command-line interface (CLI), a tool for managing and interacting with the Core framework.

## Purpose

The CLI provides a set of commands to streamline common development tasks, such as:

- Building and running applications.
- Generating tests for the public API.
- Synchronizing API definitions.
- Generating missing documentation.

## Key Packages

The CLI's user interface is built with the help of several key packages:

- **`github.com/charmbracelet/lipgloss`**: A library for styling terminal output. We use it to add color and formatting to make the CLI's output more readable and visually appealing.

- **`github.com/common-nighthawk/go-figure`**: Used to create ASCII art text banners, giving the CLI a distinct and recognizable startup message.

- **`github.com/leaanthony/clir`**: A simple and lightweight library for creating CLI applications, and our primary framework for specific, simpler command needs.

## Examples

```go
package main

import (
	"fmt"

	"github.com/leaanthony/clir"
)

func main() {
	// Create new cli
	cli := clir.NewCli("Flags", "A simple example", "v0.0.1")

	// Name
	name := "Anonymous"
	cli.StringFlag("name", "Your name", &name)

	// Define action for the command
	cli.Action(func() error {
		fmt.Printf("Hello %s!\n", name)
		return nil
	})

	if err := cli.Run(); err != nil {
		fmt.Printf("Error encountered: %v\n", err)
	}
}
```
