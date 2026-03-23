## Context

Shadow volumes (from the `shadow-node-modules` change) isolate the container from host-built native binaries by mounting empty named volumes over `node_modules`. This means dependencies must be installed inside the container. For agents, this should happen automatically before the agent starts working.

## Goals / Non-Goals

**Goals:**
- Prompt the user before installing (credentials or build tools may be required).
- Install runs inside the container where the correct platform binaries are built.
- Single consolidated prompt for all detected projects (not one per project).
- Feature can be disabled per-project.

**Non-Goals:**
- Auto-installing without prompting (too risky — may need credentials or specific build tools).
- Supporting non-agent modes (shell users can install manually).
- Detecting non-Node.js dependency managers.

## Decisions

### 1. Prompt on host, install in container

The prompt runs in Go on the host (where the user's terminal is), before the `docker exec` that starts the agent. The install commands are prepended to the agent's exec command via `bash -c`. This avoids entrypoint complexity and ensures the install happens in the interactive session.

### 2. PreRunCmds in ExecOpts

`ExecOpts` gains a `PreRunCmds []string` field. When set, the agent command is wrapped: `bash -c "source ~/.bashrc; cmd1 && cmd2 ; exec agent_binary"`. The PATH setup (`fnm env`) is included so `pnpm`, `npm`, etc. are available.

### 3. Reuse FindNodeModulesDirs for lockfile detection

The same function that finds `node_modules` paths for shadowing is reused to find projects. Each path's parent directory is checked for lockfiles (package-lock.json, pnpm-lock.yaml, yarn.lock, bun.lock/bun.lockb). Projects without lockfiles are skipped.

### 4. Shadow volume ownership fix

Docker creates named volumes as root. After the container starts, asylum runs `docker exec -u root chown claude:claude <path>` for each shadow volume so the container user can write to them.

### 5. Feature flag for opt-out

`features: { auto-install-node-modules: false }` in config disables the prompt entirely. The check uses `FeatureOff()` (default-on pattern).

## Risks / Trade-offs

- **Install failures don't block the agent**: The `bash -c` script uses `&&` between install commands but `;` before `exec agent`, so the agent starts even if installs fail.
- **PATH setup is duplicated**: The `fnm env` setup in the exec wrapper mirrors the Dockerfile's `.bashrc` setup. If the PATH setup changes in the Dockerfile, it needs updating here too.
- **Prompt only on first container start**: If the container is already running (second session), the prompt doesn't appear and installs don't run. This is correct — deps persist in the named volume.
