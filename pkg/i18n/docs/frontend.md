# Frontend Documentation

The frontend of the `i18n` project is an Angular application located in the `ui` directory. It is designed to be built as a custom element (Web Component).

## Project Structure

*   **`ui/`**: Root directory for the Angular project.
*   **`ui/src/`**: Source code for the application.
*   **`ui/dist/`**: Output directory for the build artifacts.

## Building the Application

The application is built using standard Angular CLI commands, wrapped in npm scripts.

To build the project:

```bash
cd ui
npm run build
```

This will generate the build artifacts in the `dist/i18n-element` directory. The main build output is typically found in `dist/i18n-element/browser`.

## Integration

The Go backend (`cmd/i18n serve`) is configured to serve the static files from `ui/dist/i18n-element/browser`. This allows the Angular application to be hosted directly by the Go server.

## Development

The Angular project is named `i18n-element`. You can run standard Angular CLI commands within the `ui` directory if you have the CLI installed globally, or use `npm run` to execute scripts defined in `package.json`.
