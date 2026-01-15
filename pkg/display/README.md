# Display

This repository is a display module for the core web3 framework. It includes a Go backend, an Angular custom element, and a full release cycle configuration.

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/Snider/display.git
    ```

2.  **Install the dependencies:**
    ```bash
    cd display
    go mod tidy
    cd ui
    npm install
    ```

3.  **Run the development server:**
    ```bash
    go run ./cmd/demo-cli serve
    ```
    This will start the Go backend and serve the Angular custom element.

## Building the Custom Element

To build the Angular custom element, run the following command:

```bash
cd ui
npm run build
```

This will create a single JavaScript file in the `dist` directory that you can use in any HTML page.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the EUPL-1.2 License - see the [LICENSE](LICENSE) file for details.
