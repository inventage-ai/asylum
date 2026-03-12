## Context

Asylum is a new Go CLI tool with no existing code. We need to establish the project skeleton that all subsequent features build on.

## Goals / Non-Goals

**Goals:**
- Working Go module with correct module path
- Directory structure matching PLAN.md section 8
- Makefile that cross-compiles for all four targets (linux/darwin × amd64/arm64)
- Minimal `main.go` that prints version (proves the build works)

**Non-Goals:**
- No CLI argument parsing beyond version output
- No actual feature code — just the skeleton
- No CI/CD setup

## Decisions

- **Module path**: `github.com/binaryben/asylum` — standard Go convention for the project
- **Version injection**: Use `-ldflags -X` to inject version at build time via Makefile. Define a `version` variable in `main.go` with a default dev value
- **Build output**: Binaries go to `build/` directory, gitignored. `make build` produces a single binary for the current platform; `make build-all` produces all four
- **Dependency**: Add `gopkg.in/yaml.v3` to go.mod now (needed by config system in the next change)

## Risks / Trade-offs

- None significant — this is straightforward scaffolding
