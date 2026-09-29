# Design

## Context

The guard runs at `cmd/asylum/main.go:84`, right after argument parsing. It calls `workspace.Resolve(projectDir, home, reuse)`, where `reuse` is true when the forwarded args contain `--continue` or `--resume`. Running that early causes three problems:
- `self-update` (line 131) and `update` (line 358) both reach the guard.
- The agent, its isolation and the mode aren't known yet.
- Reuse depends only on the arguments, not on the mode.

Everything between the guard and the first project `config.Load` (line 192) is independent of `projectDir`, except `self-update`, which should see the unredirected directory anyway.

## Goals / Non-Goals

**Goals:**
- `--continue` from `~` lands where the agent has something to continue.
- `shell` and `run` from `~` join the workspace the user is running.
- Only the container-run path redirects.

**Non-Goals:**
- Considering secondary containers when looking for a running workspace.
- Finding sessions of agents other than the active one.
- Changing how a plain agent launch behaves.

## Decisions

### `Resolve` takes a filter
The new signature is `Resolve(projectDir, home string, accept func(dir string) bool, match Outcome) (string, Outcome, error)`.
- `newest` sorts generated workspaces by modification time, newest first, and returns the first that `accept` allows.
- On a match, `Resolve` returns `match`. The caller passes `Continued` or `Attached`, so the workspace package never needs to know the mode.
- A nil filter means "always fresh".
- `Reused` is replaced by `Continued` and `Attached`.

### Move the guard, skip it for `update`
The guard moves to just after `cliFlags` is built and before the first project `config.Load`. `self-update` has returned by then. `update` skips the guard with an explicit check, because it runs the full image path afterwards. `containerMode` is already computed at that point.

### The session filter uses `config.Load(home, cliFlags, …)`
Loading with the home directory as the project is safe. `resolveProjectConfigPath` ignores directories, so `~/.asylum` is skipped and the result is the global config plus CLI flags. That gives the same agent choice as a normal launch (`--agent`, then the config, then `claude`) and its isolation mode. For each candidate `dir`, the filter computes `agent.ResolveConfigDir(a, isolation, container.ContainerName(dir))` and calls `a.HasSession(configDir, dir)`. The config load happens only when a resume flag is present in agent mode.

### The running filter uses one Docker call
A new `docker.RunningNames() (map[string]bool, error)` runs `docker ps --format '{{.Names}}'` once. The filter checks `names[container.ContainerName(dir)]`. On a Docker error the filter accepts nothing, and a fresh workspace follows, which matches the old behaviour.

### Mode decides the filter

| Mode | Filter | Outcome on match |
|---|---|---|
| Agent, no resume flag | nil | — |
| Agent, `--continue`/`--resume` | session | `Continued` |
| Shell, admin shell, run | running | `Attached` |

## Risks / Trade-offs

- [A workspace only running a secondary container is not seen as running] → Rare in scratch workspaces. `shell` then starts a fresh workspace, and the warning shows its path.
- [`config.Load(home, …)` also runs the v1→v2 migration on the global config] → The normal launch runs it a moment later anyway, and the migration is idempotent.
- [`HasSession` for many candidates reads many directories] → It stops at the first match, newest first, and the checks are cheap directory reads.
