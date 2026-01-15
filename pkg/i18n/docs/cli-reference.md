# CLI Reference

The `i18n` command-line tool helps you manage the application and its services.

## Base Command

```bash
i18n [command]
```

**Description:**
`i18n` is a CLI for managing internationalization and localization. It provides functions for the web server and locale management.

**Usage:**
Run `i18n` without any arguments to see the help message or a greeting (depending on configuration).

## Commands

### `serve`

Start the web server.

```bash
i18n serve
```

**Description:**
The `serve` command starts the Go web server which serves the Angular frontend and the API endpoints.

**Options:**
Currently, the server listens on port **8080** by default.

**Example:**
```bash
go run ./cmd/i18n serve
```

## Help

To get help for any command, use the `--help` or `-h` flag.

```bash
i18n --help
i18n serve --help
```
