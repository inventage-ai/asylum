## Context

Asylum currently handles Node.js dependency installation through a fragile chain: Go prompts the user, passes install commands as `PreRunCmds`, which wraps the agent command in `bash -c` with duplicated PATH setup. This is hard to extend and mixes concerns. Shadow volume ownership is fixed in the entrypoint, which is a separate concern (container setup) and stays there.

## Goals / Non-Goals

**Goals:**
- Clean separation: scan → prompt → execute, all in Go.
- Extensible: adding a new task (Python, credentials) means adding a struct, not modifying dispatch logic.
- State tracking: don't re-prompt for completed tasks unless inputs change.
- Controllable: global disable, per-task disable, CLI skip flag.

**Non-Goals:**
- Task dependencies (not needed yet — tasks are independent).
- Running onboarding in non-agent modes (shell/run users handle setup themselves).
- Shadow volume ownership (that's container setup, stays in entrypoint).

## Decisions

### 1. Task interface

Each onboarding task is a Go struct implementing:

```go
type Task interface {
    Name() string
    Detect(projectDir string) []Workload
}

type Workload struct {
    Label      string   // display name, e.g. "eamportal-view/.../angular"
    Command    []string // e.g. ["pnpm", "install", "--frozen-lockfile"]
    Dir        string   // working directory inside container
    HashInputs []string // paths to files whose hash determines re-run (e.g. lockfile)
    Phase      Phase    // PostContainer (PreContainer reserved for future use)
}
```

`Detect` returns zero or more workloads. Each workload has its own hash inputs for change detection. Per-task disable is handled centrally in `Run()` by checking `opts.Onboarding[task.Name()]` — no need for a separate `FeatureFlag` method since the task name is the key.

**Why workloads separate from tasks?** A single "npm" task can produce multiple workloads (one per package.json in a monorepo). The task decides detection logic; the framework handles prompting, state, and execution uniformly.

### 2. Orchestration in main.go

The onboarding runs between container start and session exec:

```
container started (or already running)
    │
    ▼
onboarding.Run(opts)
    ├── load state from ~/.asylum/projects/<cname>/onboarding.json
    ├── for each registered task:
    │     ├── check feature flag → skip if disabled
    │     ├── task.Detect(projectDir) → workloads
    │     ├── for each workload:
    │     │     └── compare hash → skip if unchanged
    │     └── collect pending workloads
    ├── if no pending workloads → return
    ├── prompt user (consolidated list)
    ├── if declined → return
    ├── execute accepted workloads:
    │     ├── PreContainer: exec on host (future)
    │     └── PostContainer: docker exec into container
    ├── save updated state
    └── return
    │
    ▼
docker exec agent/shell/command
```

### 3. Execution via docker exec from Go

Post-container tasks run via `docker.Exec(container, user, cmd...)` — the helper we already have. The Go process captures stdout/stderr and streams it to the user's terminal. Errors are reported but non-fatal: a failed install doesn't prevent the agent from starting.

**Why not a bash wrapper?** Running `docker exec` per workload from Go gives us:
- Individual error handling per workload
- Clean agent command (no `bash -c` wrapper)
- PATH setup handled once in the exec command, not duplicated

**PATH resolution**: `docker exec` inherits the container's initial environment, not the entrypoint's exports. The entrypoint sets up fnm, mise, and other tools that modify PATH dynamically. To make this available to `docker exec` calls:

1. The entrypoint writes its final resolved PATH to `/tmp/asylum-path` after all setup is complete.
2. `WaitReady` polls for this marker file, ensuring the entrypoint has finished before Go proceeds.
3. Before running onboarding tasks, Go reads the file via `docker exec cat /tmp/asylum-path`.
4. Each onboarding `docker exec` runs through `bash -c "export PATH=<resolved>; <command>"` — this ensures PATH is applied before command lookup in the same process (using `-e PATH=` alone would not work because Docker resolves the executable before applying environment variables).

This avoids duplicating PATH knowledge in Go and stays correct as the Dockerfile/entrypoint evolve.

### 4. State persistence

State lives at `~/.asylum/projects/<container-name>/onboarding.json`:

```json
{
  "auto-install-node-modules": {
    "eamportal-view/.../angular": "sha256:abc123",
    "eamportal-admintool/.../build": "sha256:def456"
  }
}
```

Keys are task names (matching feature flags). Values are maps of workload label → hash of inputs. On scan, if the hash matches, the workload is skipped. If the hash differs or the key is missing, the workload is pending.

`--cleanup` wipes the projects directory (already does), which resets onboarding state.

### 5. Skip mechanisms

Three levels of control:

| Mechanism | Scope | Effect |
|-----------|-------|--------|
| `--skip-onboarding` CLI flag | Single invocation | Skip all onboarding this run |
| `features: { onboarding: false }` | Permanent (config) | Disable all onboarding |
| `onboarding: { npm: false }` | Per-task (config) | Disable specific task |

The `features` map handles broad system toggles. The `onboarding` map is a dedicated config section where each key is a task name and the value is a bool (default true). This keeps per-task config separate from system features.

```yaml
# Global disable
features:
  onboarding: false

# Per-task disable
onboarding:
  npm: false
```

The CLI flag is checked first, then the global feature flag, then per-task config. The `onboarding` map follows the same merge semantics as `versions` — scalar last-wins per key.

### 6. npm task implementation

The npm task replaces the current `nodeInstallCmds` function and entrypoint install block:

- **Detect**: Reuses `FindNodeModulesDirs` to find package.json locations, checks each for lockfiles (package-lock.json, pnpm-lock.yaml, yarn.lock, bun.lock/bun.lockb).
- **Hash input**: The lockfile path. Hash changes when deps change.
- **Phase**: PostContainer (needs Linux binaries).
- **Command**: `npm ci`, `pnpm install --frozen-lockfile`, etc.

## Risks / Trade-offs

- **PATH in docker exec**: We need to pass the correct PATH explicitly on each `docker exec` call since it doesn't inherit entrypoint exports. This couples the onboarding package to the container's tool layout. Acceptable since we control both.
- **State file corruption**: Concurrent sessions could race on onboarding.json. Mitigated by: onboarding only runs on first container creation (inside the `!docker.IsRunning` block), and the session counter prevents concurrent first-starts.
- **No rollback**: If an onboarding task partially succeeds (installs some deps, fails on others), there's no undo. The state records the attempt, and re-running with `--cleanup` or changed lockfile will retry.
