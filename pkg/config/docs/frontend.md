# Frontend Documentation (`ui`)

The frontend of this project is an Angular application located in the `ui` directory. It is designed to be built as a custom element (Web Component) or a standard Angular application.

## Prerequisites

Ensure you have the following installed:
- Node.js (Latest LTS recommended)
- npm (comes with Node.js)

## Setup

Navigate to the `ui` directory and install the dependencies:

```bash
cd ui
npm install
```

## Development

To start the local development server:

```bash
ng serve
```

This will run the application at `http://localhost:4200/`. The application automatically reloads if you change any of the source files.

## Building

The project can be built for production using the standard Angular CLI build command:

```bash
ng build
```

The build artifacts will be stored in the `dist/` directory.

### Custom Element Build

*Note: If the application is configured as a Custom Element (Web Component), the build output in `dist/` typically includes a main JavaScript file that can be included in other HTML pages.*

## Testing

### Unit Tests

Unit tests are written using Jasmine and run with Karma. To execute them:

```bash
ng test
```

### End-to-End Tests

End-to-end tests can be run via:

```bash
ng e2e
```

## Project Structure

- `src/`: Source code of the Angular application.
- `angular.json`: Angular CLI configuration.
- `package.json`: Project dependencies and scripts.
