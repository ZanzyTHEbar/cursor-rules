# Contributing to cursor-rules

Thank you for your interest in contributing to cursor-rules! This document provides guidelines and instructions for contributors.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- (Optional but recommended) `golangci-lint` for linting
- (Optional but recommended) `goimports` for import formatting

### Install Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest
```

### Clone and Setup

```bash
git clone https://github.com/ZanzyTHEbar/cursor-rules.git
cd cursor-rules

# Install git hooks (optional but recommended)
bash scripts/install-hooks.sh
```

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run only unit tests (fast)
make test-unit

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Code Quality Checks

```bash
# Format code
go fmt ./...
goimports -w .

# Run linter
make lint

# Run all quality checks
make check
```

### Building

```bash
# Build CLI binary
make build

# Build extension
make ext-build

# Build everything
make build-all
```

## Git Hooks

We provide pre-commit hooks that run automatically before each commit:

### Installing Hooks

```bash
bash scripts/install-hooks.sh
```

### What the Pre-Commit Hook Does

The hook runs the following checks:
1. **gofmt** - Ensures code is properly formatted
2. **goimports** - Checks import formatting
3. **go vet** - Runs Go's built-in static analyzer
4. **golangci-lint** - Runs comprehensive linting (if installed)
5. **go test** - Runs all tests in short mode

If any check fails, the commit is blocked.

### Bypassing Hooks

In rare cases where you need to bypass hooks:

```bash
git commit --no-verify
```

**Note:** Only use `--no-verify` when absolutely necessary, as it skips important quality checks.

## Coding Standards

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Use `goimports` for import organization
- Keep functions focused and under 50 lines when possible
- Write clear, self-documenting code

### Error Handling

- Always check errors; don't ignore them with `_`
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return errors early (guard clauses)

### Testing

- Write table-driven tests for multiple test cases
- Use `t.Helper()` in test helper functions
- Aim for high test coverage of critical paths
- Tests should be deterministic and independent

### Security

- Use `internal/security` package for path validation
- Never construct file paths from user input without validation
- Use `security.SafeJoin` instead of `filepath.Join` for user paths
- Validate all input at boundaries

### Comments

- Package-level documentation for all packages
- Exported functions must have doc comments
- Comments should explain *why*, not *what*
- Keep comments up-to-date with code changes

## Pull Request Process

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards

3. **Run all checks** locally:
   ```bash
   make check
   ```

4. **Commit your changes** with conventional commits:
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve bug in parser"
   git commit -m "docs: update README"
   ```

5. **Push and create PR**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Ensure CI passes** - all tests and linting must pass

7. **Request review** - wait for maintainer review

### Conventional Commits

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Build process/tooling changes
- `perf:` - Performance improvements

## Project Structure

```
cursor-rules/
├── cli/              # CLI framework and app context
├── cmd/              # Command-line entry points
│   └── cursor-rules/ # Main CLI application
│       └── commands/ # Cobra commands
├── internal/         # Internal packages
│   ├── config/       # Configuration management
│   ├── core/         # Core business logic
│   ├── manifest/     # Manifest parsing
│   ├── security/     # Path validation and security
│   ├── testutil/     # Test utilities
│   └── transform/    # Format transformation
├── extension/        # VSCode extension
├── scripts/          # Build and utility scripts
└── Makefile          # Build targets
```

## Questions?

If you have questions or need help:

- Open an issue for bugs or feature requests
- Check existing issues and PRs first
- Be respectful and constructive in all interactions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
