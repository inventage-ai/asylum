## Why

Project onboarding logic (Node.js dependency installation, volume ownership fixes, future credential import and Python env setup) is currently scattered across the entrypoint shell script, bash-c exec wrappers, and Go dispatch code. This makes it fragile, hard to extend, and difficult to handle errors properly. The entrypoint can't reliably prompt the user, the bash wrapper duplicates PATH setup, and there's no way to run pre-container tasks.

We need a structured onboarding system in Go that scans the project, prompts once, and orchestrates tasks across the container lifecycle — with proper error handling, state tracking, and extensibility.

## What Changes

- **Onboarding framework**: A task-based system where each onboarding task (npm install, future pip install, credential import, etc.) implements a common interface with detection, prompting, and execution phases.
- **Scan → Prompt → Execute flow**: On container creation, the system scans for applicable tasks, shows a consolidated prompt, then executes accepted tasks at the appropriate phase (pre-container or post-container via `docker exec`).
- **Project-scoped state**: Completed tasks are tracked in `~/.asylum/projects/<container-name>/onboarding.json` with lockfile hashes so tasks are skipped on subsequent starts unless inputs change.
- **npm install as first task**: The existing auto-install logic is refactored into an onboarding task, removing the bash wrapper and entrypoint hacks.
- **Skip onboarding**: `--skip-onboarding` CLI flag and `features: { onboarding: false }` config option to disable all onboarding. Individual tasks can be disabled in a dedicated `onboarding` config section (e.g., `onboarding: { npm: false }`).

## Capabilities

### New Capabilities

- `project-onboarding`: Framework for detecting, prompting, and executing project setup tasks across the container lifecycle. Includes task interface, state tracking, and consolidated user prompt.
- `onboarding-npm`: Node.js dependency installation as an onboarding task — detects lockfiles, runs the appropriate package manager via `docker exec`.

### Modified Capabilities

- `cli-dispatch`: Main dispatch gains onboarding orchestration between container start and session exec. New `--skip-onboarding` flag.
- `container-assembly`: `ExecArgs` no longer needs `PreRunCmds` / bash wrapper for agent mode.

## Impact

- **New package**: `internal/onboarding/` with task interface, scanner, state persistence, and npm task implementation.
- **CLI dispatch**: `main.go` calls onboarding between container start and `docker exec` for the session.
- **Entrypoint**: Node.js install block removed (volume ownership fix stays — it's container setup, not onboarding).
- **Container exec**: Agent command is no longer wrapped in `bash -c` — runs directly.
- **State storage**: New file `~/.asylum/projects/<container-name>/onboarding.json`.
- **Config**: `features: { onboarding: false }` for global disable. New `onboarding` map in Config struct for per-task control (e.g., `onboarding: { npm: false }`).
