Thank you for your interest in contributing to cursor-rules!

Guidelines

-   Fork the repository and open a pull request against `main`.
-   Write clear, focused PR descriptions referencing any related issues.
-   Run `make test` and ensure all tests pass.
-   Follow existing code style and run `gofmt` before committing.

Development

### Prerequisites

-   Go 1.25.2+ (required)
-   Node.js and pnpm (for extension development)
-   Docker (optional, for Dev Container)
-   pre-commit (optional, for git hooks)

### Building

-   Use `make build` to build the CLI binary.
-   Use `make ext-build` to build the VSCode extension.
-   Use `make test` to run the test suite.

### Pre-Commit Hooks (Recommended)

We use pre-commit hooks to ensure code quality before commits:

```bash
# Install pre-commit (if not already installed)
pip install pre-commit
# or
brew install pre-commit

# Install the git hooks
pre-commit install

# Run manually on all files
pre-commit run --all-files
```

The hooks will automatically:
- Format Go code with `gofmt` and `goimports`
- Run `go vet` for static analysis
- Run quick tests (`go test -short`)
- Check YAML and Markdown formatting
- Fix trailing whitespace and line endings

### Dev Container (Recommended)

For a consistent development environment:

1. Open project in VS Code or Cursor
2. Install "Dev Containers" extension
3. Press `F1` → "Dev Containers: Reopen in Container"
4. Container includes Go 1.25.2, gopls, delve, and all dev tools

See [.devcontainer/BUILD_INSTRUCTIONS.md](.devcontainer/BUILD_INSTRUCTIONS.md) for detailed instructions.

License

This project is licensed under the MIT License. See `LICENSE` for details.
