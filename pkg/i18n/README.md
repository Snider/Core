# i18n

This repository is a template for developers to create custom HTML elements. It includes a Go backend, an Angular custom element, and a full release cycle configuration.

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/snider/i18n.git
    ```

2.  **Install the dependencies:**
    ```bash
    cd i18n
    go mod tidy
    cd ui
    npm install
    ```

3.  **Run the development server:**
    ```bash
    go run ./cmd/i18n serve
    ```
    This will start the Go backend and serve the Angular custom element.

## Usage

To see how to use the `i18n` library in your own Go program, check out the example in the `examples/simple` directory.

To run the example, use the following command:

```bash
go run ./examples/simple
```

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
