## ADDED Requirements

### Requirement: Go module initialization
The project SHALL have a valid Go module at `github.com/binaryben/asylum` with `gopkg.in/yaml.v3` as a dependency.

#### Scenario: Module is valid
- **WHEN** `go mod tidy` is run in the project root
- **THEN** it completes without errors

### Requirement: Directory structure
The project SHALL have the directory layout specified in PLAN.md section 8: `cmd/asylum/`, `internal/agent/`, `internal/config/`, `internal/container/`, `internal/docker/`, `internal/image/`, `internal/log/`, `internal/ssh/`, and `assets/`.

#### Scenario: All directories exist
- **WHEN** the project is checked out
- **THEN** all specified directories exist (with placeholder files where needed for git)

### Requirement: Cross-compilation Makefile
The Makefile SHALL provide `build`, `build-all`, `clean`, and `test` targets. `build-all` SHALL produce binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64.

#### Scenario: Build for current platform
- **WHEN** `make build` is run
- **THEN** a single binary is produced in `build/` for the current OS/architecture

#### Scenario: Cross-compile all targets
- **WHEN** `make build-all` is run
- **THEN** four binaries are produced: `asylum-linux-amd64`, `asylum-linux-arm64`, `asylum-darwin-amd64`, `asylum-darwin-arm64`

### Requirement: Version output
The entry point SHALL print the tool version when run with no arguments (placeholder behavior for this change only).

#### Scenario: Version display
- **WHEN** the built binary is executed
- **THEN** it prints the version string to stdout
