# Frontend Documentation

The frontend is an Angular application located in the `ui/` directory. It is designed to be built as a custom element or a standalone application that runs inside the Wails webview.

## Structure

- **Path:** `ui/`
- **Framework:** Angular
- **Output:** The build process generates artifacts in `ui/dist`.

## Development

The frontend can be developed using standard Angular CLI commands or via the provided demo CLI which serves the built files.

## Build Process

To build the frontend for production or integration with the Go backend:

```bash
cd ui
npm install
npm run build
```

This will compile the Angular application and place the output in the `dist/` directory, which the Go backend can then serve or embed.

## Integration

The frontend communicates with the Go backend through the Wails runtime. It can trigger actions defined in the backend (like opening windows) and receive events.
