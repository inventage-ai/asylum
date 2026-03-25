## Why

The existing integration tests (`integration/`) exercise the Docker image directly via `docker run`, but never run the actual `asylum` binary. This means the full flow — config loading, kit/agent resolution, image building, container lifecycle, agent exec, session tracking — is only tested manually. E2e tests that invoke the compiled binary would catch wiring bugs, flag parsing issues, and config interaction problems that unit tests and image-level integration tests miss.

The new kit system makes fast e2e tests practical: `kits: {}` produces a minimal image (core only, no languages), and a dummy `echo` agent can verify the agent exec path without real credentials.

## What Changes

- New `echo` agent: a minimal agent implementation that runs `echo` with any provided args. Not installed by default (no AgentInstall) — it's a testing-only agent for verifying the exec path.
- New e2e test suite (`e2e/`): builds the binary, runs it against a temp project directory with a minimal config (`kits: {}`, `agents: {}`), and verifies the full lifecycle
- Test cases: binary builds, help flag works, image builds with minimal config, container starts, agent exec runs and exits, container cleaned up after last session, shell mode works, run mode works, config flags apply

## Capabilities

### New Capabilities
- `e2e-testing`: End-to-end test framework that builds and runs the asylum binary with a dummy agent against a minimal Docker image

## Impact

- **internal/agent/echo.go** (new): Echo agent implementation (testing-only)
- **e2e/** (new directory): E2e test files, build tag `e2e`
- **Makefile**: New `test-e2e` target
