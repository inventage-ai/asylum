## Why

Asylum needs a Go project foundation before any features can be implemented. This sets up the module, directory structure, build system, and verifies cross-compilation works for all four targets.

## What Changes

- Initialize Go module (`github.com/binaryben/asylum`) with `gopkg.in/yaml.v3` dependency
- Create directory structure per PLAN.md section 8: `cmd/asylum/`, `internal/` packages, `assets/`
- Create `Makefile` with `build`, `build-all`, `clean`, and `test` targets
- Create minimal `main.go` that prints version
- Verify `make build` and `make build-all` produce working binaries

## Capabilities

### New Capabilities
- `project-scaffold`: Go module init, directory layout, Makefile with cross-compilation, and version-printing entry point

### Modified Capabilities

None.

## Impact

- Creates the entire project skeleton from scratch
- All subsequent changes build on this foundation
- No external impact — this is the initial setup
