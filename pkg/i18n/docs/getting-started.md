# Getting Started

This guide will help you set up and run the `i18n` project on your local machine.

## Prerequisites

Before you begin, ensure you have the following installed:

*   [Go](https://golang.org/doc/install) (version 1.16 or later recommended)
*   [Node.js](https://nodejs.org/) and npm (for the frontend)

## Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/snider/i18n.git
    cd i18n
    ```

2.  **Install Go dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Install Frontend dependencies:**

    Navigate to the `ui` directory and install the npm packages:

    ```bash
    cd ui
    npm install
    cd ..
    ```

## Running the Application

The application consists of a Go backend and an Angular frontend. You can run them together using the `serve` command.

1.  **Build the Frontend:**

    First, you need to build the Angular application so that the Go server can serve the static files.

    ```bash
    cd ui
    npm run build
    cd ..
    ```

2.  **Start the Server:**

    Run the Go application using the `serve` command:

    ```bash
    go run ./cmd/i18n serve
    ```

    You should see output indicating that the server is listening on port 8080:

    ```text
    Listening on :8080...
    ```

3.  **Access the Application:**

    Open your web browser and navigate to `http://localhost:8080`. You should see the `i18n-element` application running.

    There is also a demo API endpoint available at `http://localhost:8080/api/v1/demo`.
