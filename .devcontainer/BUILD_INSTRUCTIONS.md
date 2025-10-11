# Development Container Instructions

This document provides instructions for using the Dev Container for the cursor-rules project.

## Overview

The `.devcontainer/Dockerfile` provides a **complete development environment** for building and developing the cursor-rules CLI tool and VSCode extension.

**Base Image:** `golang:1.25.2-bookworm` (Official Go image)  
**Purpose:** Development, testing, and building  
**Not for:** Production deployments (use `make release-binaries` instead)

## Prerequisites

- Docker 20.10+ or compatible container runtime
- VS Code or Cursor with Dev Containers extension
- OR: Docker CLI for manual container usage

## Using with VS Code/Cursor (Recommended)

### 1. Open in Dev Container

**VS Code:**
1. Install "Dev Containers" extension
2. Open project folder
3. Press `F1` → "Dev Containers: Reopen in Container"
4. Wait for container to build and start

**Cursor:**
1. Open project folder
2. Cursor will detect `.devcontainer/devcontainer.json`
3. Click "Reopen in Container" when prompted
4. Wait for container to build and start

### 2. Verify Environment

Once inside the container:

```bash
# Check Go version
go version
# Should output: go version go1.25.2 linux/amd64

# Check installed tools
gopls version
dlv version
staticcheck -version

# Build the project
make build

# Run tests
make test
```

## Manual Docker Usage (Advanced)

If you prefer to use Docker directly without VS Code/Cursor:

### Build the Container

```bash
# From project root
docker build -f .devcontainer/Dockerfile -t cursor-rules-dev .
```

### Run the Container

```bash
# Run with workspace mounted
docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  cursor-rules-dev \
  /bin/bash

# Inside container, you can now:
go build ./cmd/cursor-rules
go test ./...
make build
```

## What's Included

### Base Image

**golang:1.25.2-bookworm** - Official Go image based on Debian Bookworm

### Development Tools

- **Go 1.25.2** - Full Go toolchain
- **git** - Version control
- **make** - Build automation
- **build-essential** - C/C++ compiler toolchain (for CGO if needed)
- **curl, wget** - Download tools
- **vim, nano** - Text editors

### Go Development Tools (Pre-installed)

- **gopls** - Go language server for IDE features
- **delve (dlv)** - Go debugger
- **staticcheck** - Go linter
- **goimports** - Import formatter

### Environment Configuration

- **GOPATH:** `/go`
- **Working Directory:** `/workspace`
- **PATH:** Includes `$GOPATH/bin` and Go binaries

## Dev Container Configuration

The `.devcontainer/devcontainer.json` configures the development environment:

```json
{
  "name": "cursor-rules (Go 1.25.2)",
  "build": {
    "context": "..",
    "dockerfile": "Dockerfile"
  },
  "workspaceFolder": "/workspace",
  "customizations": {
    "vscode": {
      "settings": {
        "go.gopath": "/go",
        "go.useLanguageServer": true,
        "go.formatTool": "goimports",
        "editor.formatOnSave": true
      },
      "extensions": ["golang.go"]
    }
  },
  "postCreateCommand": "go mod download && go mod verify"
}
```

**Key Features:**
- Automatically installs Go extension
- Configures Go language server (gopls)
- Enables format-on-save with goimports
- Downloads dependencies on container creation

## Testing the Container

### Quick Test

```bash
# Build container
docker build -f .devcontainer/Dockerfile -t cursor-rules-dev .

# Verify Go installation
docker run --rm cursor-rules-dev go version

# Verify Go tools
docker run --rm cursor-rules-dev gopls version

# Run with workspace mounted
docker run --rm -v $(pwd):/workspace -w /workspace cursor-rules-dev go test ./...
```

### Full Development Test

```bash
# Start interactive session
docker run -it --rm -v $(pwd):/workspace -w /workspace cursor-rules-dev /bin/bash

# Inside container:
go mod download
make build
make test
./bin/cursor-rules --help
```

## Build Optimization

### Layer Caching

The Dockerfile is optimized for fast rebuilds:

1. Base image with Go 1.25.2 (cached)
2. Install system dependencies (cached)
3. Install Go tools (cached)
4. Your source code is mounted at runtime (not copied)

### .dockerignore

A `.dockerignore` file excludes unnecessary files from build context:

- Build artifacts (`bin/`, `dist/`)
- Node modules (installed inside container)
- Temporary files
- OS files

This reduces build context size and improves build speed.

### Workspace Mounting

The devcontainer mounts your workspace with `consistency=cached` for better performance on macOS/Windows.

## Troubleshooting

### Container Build Fails

**Issue:** Docker build fails with errors

**Solution:** Ensure you're building from project root:

```bash
# From project root
docker build -f .devcontainer/Dockerfile -t cursor-rules-dev .
```

### Go Version Mismatch

**Issue:** `go.mod` requires different Go version

**Solution:** Update Dockerfile to match `go.mod`:

```dockerfile
FROM golang:1.25.2-bookworm
```

And update `go.mod`:

```go
go 1.25.2
```

### Dev Container Won't Start

**Issue:** VS Code/Cursor fails to start container

**Solutions:**
1. Check Docker is running: `docker ps`
2. Rebuild container: `F1` → "Dev Containers: Rebuild Container"
3. Check logs: `F1` → "Dev Containers: Show Container Log"

### Dependencies Not Found

**Issue:** `go build` fails with missing dependencies

**Solution:** Run inside container:

```bash
go mod download
go mod verify
```

Or rebuild container (runs `postCreateCommand` automatically).

## CI/CD Integration

### GitHub Actions Example

For CI/CD, use the official Go action instead of building the devcontainer:

```yaml
name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25.2'
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Build
      run: make build
```

**Note:** The devcontainer is for local development only. CI/CD should use native Go actions for better performance.

## Best Practices

### Development Workflow

1. **Use Dev Containers** - Consistent environment across team
2. **Mount Workspace** - Don't copy source into container
3. **Install Extensions** - Configure in `devcontainer.json`
4. **Format on Save** - Enable `goimports` for consistent formatting
5. **Run Tests Frequently** - Use `make test` or `go test ./...`

### Container Management

1. **Rebuild When Needed** - After Dockerfile changes or dependency updates
2. **Use .dockerignore** - Keep build context small
3. **Pin Go Version** - Match Dockerfile and `go.mod`
4. **Cache Dependencies** - `postCreateCommand` downloads on creation

### Production Builds

For production deployments, **don't use the devcontainer**. Instead:

```bash
# Use make for production builds
make release-binaries VERSION=1.0.0

# Or build directly
CGO_ENABLED=0 go build -ldflags="-s -w" -o cursor-rules ./cmd/cursor-rules
```

## Container Size

**Development Container:** ~1.2 GB (includes full Go toolchain and dev tools)

This is appropriate for development but **not for production**. Use `make release-binaries` to create optimized production binaries.

## Additional Resources

- [Dev Containers Documentation](https://containers.dev/)
- [VS Code Dev Containers](https://code.visualstudio.com/docs/devcontainers/containers)
- [Go Official Docker Images](https://hub.docker.com/_/golang)
- [Go Development in Containers](https://docs.docker.com/language/golang/)

## Quick Reference

```bash
# Open in Dev Container (VS Code/Cursor)
F1 → "Dev Containers: Reopen in Container"

# Build manually
docker build -f .devcontainer/Dockerfile -t cursor-rules-dev .

# Run manually
docker run -it --rm -v $(pwd):/workspace -w /workspace cursor-rules-dev /bin/bash

# Inside container
go mod download
make build
make test
```

---

**Last Updated:** 2025-01-11  
**Go Version:** 1.25.2  
**Purpose:** Development environment only (not for production)
