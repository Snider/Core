# CLI Documentation (`cmd/demo-cli`)

The `demo-cli` is a command-line interface application built to demonstrate the capabilities of the Config Module and serve the frontend application. It uses the `cobra` library for command management.

## Installation

You can run the CLI directly using `go run`:

```bash
go run ./cmd/demo-cli <command>
```

Or build it into a binary:

```bash
go build -o demo-cli ./cmd/demo-cli
./demo-cli <command>
```

## Commands

### `serve`

The `serve` command starts an HTTP server that serves both the Angular frontend and a demo API endpoint.

**Usage:**

```bash
go run ./cmd/demo-cli serve
```

**Features:**
- **Frontend Serving**: Serves static files from `ui/dist/config/browser`.
- **API Endpoint**: Exposes a demo endpoint at `/api/v1/demo`.
- **Port**: Listens on port `8080`.

**Example Output:**
```
Listening on :8080...
```

Access the application at `http://localhost:8080`.

### Root Command

Running the CLI without any subcommands prints the help message or executes the default action (if configured).

```bash
go run ./cmd/demo-cli
```
