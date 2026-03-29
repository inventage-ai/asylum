# Building from Source

## Prerequisites

- [Go](https://go.dev/dl/) (latest stable)
- [Docker](https://docs.docker.com/get-docker/) (for running, not building)

## Build

```sh
git clone https://github.com/inventage-ai/asylum.git
cd asylum
make build          # Build for current platform
make build-all      # Cross-compile all targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
```

The binary is output to `bin/asylum`.

## Test

```sh
make test           # Unit tests
make test-integration  # Integration tests (requires Docker, slow)
```

Integration tests are gated behind `-tags integration` and excluded from `go test ./...`. They spin up real Docker containers and take several seconds per test.

## Dependencies

Asylum has a single external dependency: `gopkg.in/yaml.v3`. Everything else is Go standard library.
