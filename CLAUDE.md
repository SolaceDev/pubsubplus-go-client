# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

The Solace Messaging API for Go (`solace.dev/go/messaging`) is a high-performance Go client library for connecting to Solace Event Brokers. It wraps the native Solace C messaging library (CCSMP/libsolclient) via Cgo, providing pub/sub, request/reply, persistent messaging, and cache APIs.

Current version: **1.10.1** (defined in `version.go`)

## Build Commands

```bash
# Build the library (requires Go 1.17+)
go build ./...

# Run unit tests (internal package tests)
go test ./...

# Run unit tests with race detection
go test -race ./...

# Run static analysis
go vet ./...
go install honnef.co/go/tools/cmd/staticcheck@v0.4.7
staticcheck --checks=all ./...

# Check formatting
go fmt ./...
```

## Running Integration Tests

The integration tests live in the `./test` directory as a **separate Go module**. They require Docker for spinning up a Solace broker via testcontainers.

```bash
# First, generate the SEMPv2 client (requires Docker)
cd test/sempclient
go generate .
cd ..

# Install the Ginkgo test runner (pinned version)
go install github.com/onsi/ginkgo/v2/ginkgo@v2.23.4

# Run all integration tests (uses testcontainers to start a broker)
ginkgo

# Run a specific test by regex
ginkgo --focus="mytestregex"

# Run with debug logging
ginkgo -tags enable_debug_logging

# Run against an external broker (instead of testcontainers)
go test -tags remote

# Generate coverage report
go test -coverprofile coverage.out -coverpkg solace.dev/go/messaging/internal/...,solace.dev/go/messaging/pkg/...
go tool cover -html coverage.out
```

**Note:** The test module uses `replace` directives to reference the parent module locally. The Ginkgo suite timeout is 1 hour (vs `go test` default of 10 minutes), so prefer `ginkgo` over `go test` for integration tests.

## Architecture

### Layer Structure

```
User Application
    ↓
Entrypoint  (./)                               ← messaging.NewMessagingServiceBuilder()
    ↓
Public API  (pkg/solace/)                      ← interfaces, errors, lifecycle
    ↓
Implementation  (internal/impl/)               ← concrete service, publisher, receiver impls
    ↓
Core  (internal/impl/core/)                    ← transport, events, metrics, publisher/receiver core
    ↓
CCSMP Bindings  (internal/ccsmp/)              ← Cgo wrappers around libsolclient
    ↓
libsolclient  (internal/ccsmp/lib/{platform}/) ← pre-compiled static C library per platform
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `./` (root) | Module entrypoint: `NewMessagingServiceBuilder()`, `ReplicationGroupMessageIDOf()` |
| `pkg/solace/` | Public API interfaces: `MessagingService`, publishers, receivers, lifecycle |
| `pkg/solace/config/` | Configuration properties and constants |
| `pkg/solace/message/` | Message interfaces and SDT (Structured Data Types) |
| `pkg/solace/message/rgmid/` | Replication Group Message ID type |
| `pkg/solace/metrics/` | Metrics constants and interfaces |
| `pkg/solace/subcode/` | Error subcodes (generated from C headers) |
| `pkg/solace/resource/` | Topic and queue resource types |
| `pkg/solace/logging/` | Logging interfaces |
| `internal/impl/` | MessagingService implementation and builder |
| `internal/impl/publisher/` | Direct, persistent, and request-reply publisher impls |
| `internal/impl/receiver/` | Direct, persistent, and request-reply receiver impls |
| `internal/impl/core/` | Transport layer, session management, metrics |
| `internal/impl/message/` | Message implementation |
| `internal/impl/executor/` | Async task execution |
| `internal/impl/future/` | Future/promise patterns |
| `internal/impl/logging/` | Logging implementation |
| `internal/impl/validation/` | Input validation |
| `internal/impl/provisioner/` | Endpoint provisioning implementation |
| `internal/ccsmp/` | Cgo bindings to the native C library |
| `internal/generator/` | Code generation utilities |

### Native Library Management

Pre-compiled static libraries (`libsolclient.a`) are stored per platform in:
- `internal/ccsmp/lib/darwin/` (macOS arm64/x86_64)
- `internal/ccsmp/lib/linux_amd64/`
- `internal/ccsmp/lib/linux_arm64/`

C headers are in `internal/ccsmp/lib/include/solclient/`.

### OS Support

- Linux x86/x86_64 (glibc and musl/Alpine)
- Linux arm64 (glibc)
- macOS 10.15+ (x86_64), macOS 11.0+ (arm64)
- Windows WSL 2.0

## Generated Code

Certain files are generated from the C API headers. To regenerate after CCSMP updates:

```bash
# Generate subcodes
export SOLCLIENT_H=./internal/ccsmp/lib/include/solclient/solClient.h
cd pkg/solace/subcode
go generate .

# Generate CCSMP props/enums
export SOLCACHE_H=./internal/ccsmp/lib/include/solclient/solCache.h
export SOLCLIENT_H=./internal/ccsmp/lib/include/solclient/solClient.h
cd internal/ccsmp
go generate .
```

Generated files follow the `*_generated.go` naming convention.

## Testing Strategy

- **Unit tests** (`*_test.go` alongside implementation): Standard `go test`, run in CI with race detection.
- **Integration tests** (`./test/`): Separate module using Ginkgo/Gomega, driven by testcontainers (Docker-based Solace broker + ToxiProxy for network simulation).
- **Test framework**: Ginkgo v2 with Gomega matchers.
- **SEMPv2 client**: Generated from OpenAPI specs via Docker (not committed to repo; must run `go generate` in `test/sempclient/`).

### Test Environment Variables

| Variable | Purpose |
|----------|---------|
| `PUBSUB_HOST` | Broker hostname (for remote tests) |
| `PUBSUB_PORT_PLAINTEXT` | Plaintext messaging port |
| `PUBSUB_PORT_SSL` | SSL messaging port |
| `PUBSUB_PORT_SEMP` | SEMP management port |
| `PUBSUB_VPN` | Message VPN name |
| `PUBSUB_USERNAME` | Client username |
| `PUBSUB_MGMT_USER` | SEMP management username |
| `PUBSUB_MGMT_PASSWORD` | SEMP management password |

## CI/CD

- **GitHub Actions** (`.github/workflows/test.yml`): Runs on push/PR. Compatibility check with Go 1.17, full lint+test suite on Go 1.22 (Linux).
- **Jenkinsfile**: Internal CI pipeline testing across platforms (Linux x86_64, Linux ARM, Darwin x86_64, Darwin ARM, Linux musl) with multiple Go versions.

## Local Development Setup

### Prerequisites

- Go 1.17+ (1.22+ recommended for latest tooling)
- Docker (for integration tests)
- A Go-enabled editor with format-on-save (VS Code with Go extension recommended)

### IDE Note

The integration tests are a separate Go module (`./test/go.mod`). If using VS Code with gopls, open the `test/` directory in its own workspace to avoid multi-module issues.

### Quick Start

```bash
# Clone and verify build
git clone https://github.com/SolaceDev/solace-messaging-go-client.git
cd solace-messaging-go-client
go build ./...
go test ./...

# Set up Go proxy (recommended)
export GOPROXY=https://proxy.golang.org,direct

# Run integration tests (requires Docker)
cd test/sempclient && go generate . && cd ..
go install github.com/onsi/ginkgo/v2/ginkgo@v2.23.4
ginkgo
```

## Key Files

| File | Purpose |
|------|---------|
| `go.mod` | Module definition (`solace.dev/go/messaging`, Go 1.17) |
| `messaging.go` | Public entrypoint: `NewMessagingServiceBuilder()` |
| `version.go` | API version constant |
| `doc.go` | Package-level godoc |
| `test/go.mod` | Integration test module (separate workspace) |
| `test/data/compose/docker-compose.yml` | Broker + ToxiProxy for local testing |
| `.github/workflows/test.yml` | CI pipeline definition |
| `Jenkinsfile` | Internal multi-platform CI |
| `CONTRIBUTING.md` | Contribution guidelines |
