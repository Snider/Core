# Help Module

[![Go Reference](https://pkg.go.dev/badge/github.com/Snider/help.svg)](https://pkg.go.dev/github.com/Snider/help)
[![Go Report Card](https://goreportcard.com/badge/github.com/Snider/help)](https://goreportcard.com/report/github.com/Snider/help)
[![Codecov](https://codecov.io/gh/Snider/help/branch/main/graph/badge.svg)](https://codecov.io/gh/Snider/help)
[![Build Status](https://github.com/Snider/help/actions/workflows/go.yml/badge.svg)](https://github.com/Snider/help/actions/workflows/go.yml)
[![License: EUPL-1.2](https://img.shields.io/badge/License-EUPL--1.2-yellow.svg)](https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12)

This repository contains the `help` module, which was formerly part of the `Snider/Core` framework. This module provides assistance and documentation functionality.

## Getting Started

This project uses `mkdocs-material` to build the documentation. To get started, you will need to have Python and `pip` installed.

1. **Install the dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

2. **Run the development server:**
   ```bash
   mkdocs serve
   ```

## Usage

To use the `help` module, you first need to import it in your Go project:

```go
import "github.com/Snider/help"
```

Next, initialize the help service by calling the `New` function. The `New` function accepts an `Options` struct, which allows you to configure the documentation source.

### Using a custom `embed.FS`

You can provide your own `embed.FS` as a documentation source. This is useful when you want to bundle the documentation with your application.

```go
import (
    "embed"
    "github.com/Snider/help"
)

//go:embed all:my-docs/build
var myDocs embed.FS

func main() {
    helpService, err := help.New(help.Options{
        Assets: myDocs,
    })
    if err != nil {
        // Handle error
    }
    // ...
}
```

### Custom Static Site Source

You can also provide a custom directory containing a static website as the documentation source. To do this, set the `Source` field in the `Options` struct to the path of your static site directory:

```go
helpService, err := help.New(help.Options{
    Source: "path/to/your/static/site",
})
if err != nil {
    // Handle error
}
```

Once the help service is initialized, you can use the `Show()` and `ShowAt()` methods to display the documentation.

### Displaying Help

The `Show()` method opens the help window to the main page.

```go
err := helpService.Show()
if err != nil {
    // Handle error
}
```

The `ShowAt()` method opens the help window to a specific anchor. The provided anchor is normalized into a URL. For example, calling `ShowAt("ui/how/settings#resetPassword")` will open the help window to a URL similar to `http://localhost:8080/docs/ui/how/settings/index.html#resetPassword`. The exact URL depends on how your display service resolves these paths.

```go
err := helpService.ShowAt("ui/how/settings#resetPassword")
if err != nil {
    // Handle error
}
```
