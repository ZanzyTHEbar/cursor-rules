# ============================================================================
# Development Container for cursor-rules project
# ============================================================================
# This Dockerfile provides a complete development environment for building
# and developing the cursor-rules CLI tool and VSCode extension.
#
# Base: Official Go 1.25.2 image (Debian Bookworm)
# Purpose: Development, testing, and building
# ============================================================================

FROM golang:1.25.2-bookworm

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive \
    GOPATH=/go \
    PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Install development dependencies
# - git: Version control
# - make: Build automation
# - build-essential: C/C++ compiler toolchain (for CGO if needed)
# - curl, wget: Download tools
# - ca-certificates: SSL/TLS certificates
# - vim, nano: Text editors for quick edits
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    build-essential \
    curl \
    wget \
    ca-certificates \
    vim \
    nano \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create GOPATH directory structure with proper permissions
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install common Go development tools
# These will be cached in the image for faster container startup
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest

# Set working directory to /workspace/cursor-rules (Gitpod standard)
WORKDIR /workspace/cursor-rules

# Verify Go installation
RUN go version && \
    echo "Go development environment ready"

# Default command (can be overridden by devcontainer.json)
CMD ["/bin/bash"]
