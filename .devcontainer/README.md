# Dev Container Configuration

This directory contains the development container configuration for the cursor-rules project.

## Files

- **Dockerfile** - Single-stage development container with Go 1.25.2
- **devcontainer.json** - VS Code/Cursor Dev Container configuration
- **.dockerignore** - Files to exclude from Docker build context
- **BUILD_INSTRUCTIONS.md** - Detailed usage instructions
- **CHANGELOG_DOCKERFILE.md** - Version history and changes

## Quick Start

### Using VS Code/Cursor (Recommended)

1. Open project in VS Code or Cursor
2. Install "Dev Containers" extension
3. Press `F1` → "Dev Containers: Reopen in Container"
4. Wait for container to build and start

### Manual Docker Build

```bash
# From project root
cd .devcontainer
docker build -t cursor-rules-dev .

# Run container
docker run -it --rm \
  -v $(pwd)/..:/workspace \
  -w /workspace \
  cursor-rules-dev \
  /bin/bash
```

## Configuration Details

### Build Context

- **Context:** `.devcontainer` directory
- **Dockerfile:** `Dockerfile` (in same directory)
- **Working Directory:** `/workspace` (project root mounted here)

### Environment

- **Base Image:** `golang:1.25.2-bookworm`
- **Go Version:** 1.25.2
- **GOPATH:** `/go`
- **Pre-installed Tools:**
  - gopls (language server)
  - delve (debugger)
  - staticcheck (linter)
  - goimports (formatter)

### VS Code Integration

The devcontainer automatically:
- Installs Go extension
- Configures language server
- Enables format-on-save
- Downloads Go dependencies on container creation

## Troubleshooting

### Build Fails

If the container fails to build:

1. Check Docker is running: `docker ps`
2. Verify Go version in Dockerfile matches `go.mod`
3. Check build context is correct (should be `.devcontainer` directory)

### Container Won't Start

If VS Code/Cursor fails to start the container:

1. Rebuild: `F1` → "Dev Containers: Rebuild Container"
2. Check logs: `F1` → "Dev Containers: Show Container Log"
3. Verify devcontainer.json syntax (JSONC format with comments)

### Dependencies Not Found

If `go build` fails with missing dependencies:

```bash
# Inside container
go mod download
go mod verify
```

Or rebuild the container (runs `postCreateCommand` automatically).

## Notes

- This is a **development container only** - not for production
- For production builds, use `make release-binaries`
- The container includes full Go toolchain (~1.2 GB)
- Source code is mounted, not copied (for live editing)

## See Also

- [BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md) - Detailed build instructions
- [CHANGELOG_DOCKERFILE.md](CHANGELOG_DOCKERFILE.md) - Version history
- [Dev Containers Documentation](https://containers.dev/)
