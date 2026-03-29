## Context

Ports are currently a simple `[]string` on `Config`, passed through `appendPorts` into `-p` flags. Users manually specify them. There is no dynamic allocation — every port must be explicitly configured.

The kit system already supports `DefaultOn`, `RulesSnippet`, `Tools`, and per-kit config via `KitConfig`. Container state is stored in `~/.asylum/projects/<container-name>/`. The container name is a deterministic hash of the project directory.

## Goals / Non-Goals

**Goals:**
- Pre-allocate a range of high host ports per project, so agents can start web servers without manual config
- Persist allocations globally so no two projects collide even when running concurrently
- Make the allocated ports visible to the agent via the sandbox rules file
- Allow users to configure the number of ports per project

**Non-Goals:**
- Replacing the existing `ports:` config mechanism (it continues to work independently)
- Dynamic port discovery (detecting which ports the agent actually uses)
- Port ranges per kit or per service (one flat range per project)

## Decisions

### 1. Global port registry at `~/.asylum/ports.json`

A JSON file mapping project directory → allocated port range. Format:

```json
{
  "ranges": [
    {"project": "/home/user/myapp", "start": 10000, "count": 5},
    {"project": "/home/user/other", "start": 10005, "count": 5}
  ]
}
```

File-locked during reads and writes (same `syscall.Flock` pattern used by the session counter). The project directory is the key (not the container name), because it's what the user understands and it's stable.

*Alternative: Store per-project in `~/.asylum/projects/<container>/ports`.* Rejected — we need a global view to allocate non-overlapping ranges. A single file is simpler than scanning directories.

### 2. High starting port, sequential allocation

Allocations start at port 10000 and grow upward. Each new project gets the next `count` ports after the highest allocated range. Port count defaults to 5 but is configurable via `kits: { ports: { count: 10 } }` using the existing `KitConfig` — we reuse the generic integer field approach, reading it from a new config accessor.

*Alternative: Random high ports.* Rejected — sequential is predictable and easier to reason about. Collisions with other software in the 10000+ range are unlikely.

### 3. Ranges are stable per project, reclaimed on cleanup

Once a project is allocated a range, it keeps it forever (across container restarts). Ranges are only reclaimed when `asylum cleanup` explicitly removes a project's state. This means a user's bookmarked `http://localhost:10000` stays valid.

If the configured count increases, the existing range is extended (not reallocated) — the new ports are appended after the current range, as long as they don't collide with another project's range. If they do collide, warn and keep the old count.

### 4. Port allocation as a standalone package

New package `internal/ports` with:
- `Allocate(projectDir string, count int) (Range, error)` — returns existing or allocates new
- `Release(projectDir string) error` — removes allocation (called from cleanup)
- `Load() ([]Range, error)` — reads current state

This keeps allocation logic separate from the kit and container packages.

### 5. Kit wires into RunArgs via a new field

The ports kit doesn't use `DockerSnippet` or `EntrypointSnippet` — it has nothing to install or configure at image build time. Instead, it contributes allocated ports to the container at start time.

`RunOpts` gains an `AllocatedPorts []PortMapping` field (host→container pairs). The kit's allocation runs in `main.go` before `RunArgs`, and the results are passed through. `RunArgs` appends them as `-p` flags alongside the user-configured ports.

### 6. Dynamic RulesSnippet

The ports kit can't use a static `RulesSnippet` because the actual port numbers are only known after allocation. Instead, the kit provides a function to generate the snippet at runtime. We add the snippet directly in `generateSandboxRules` by accepting the allocated ports and formatting them there, rather than changing the kit assembly pattern.

### 7. Container ports mirror host ports

For simplicity, if the host port is 10000, the container port is also 10000. This avoids confusion — the agent sees the same port number the user accesses. The `-p 10000:10000` pattern is already how the existing port shorthand works.

## Risks / Trade-offs

- **Port exhaustion**: With 5 ports per project and starting at 10000, we have ~11,000 projects before hitting 65535. Acceptable for any realistic workload.
- **Port conflicts**: Another application might use a port in our range. The user would see a Docker bind error on container start. Mitigation: high starting port minimizes risk; if it happens, user can adjust the starting port or the conflicting app.
- **Stale allocations**: If a user deletes a project without running `asylum cleanup`, its range stays allocated. Not a real problem given the port space available, and `cleanup` already handles state removal.
- **Concurrent allocation**: Two `asylum` processes starting simultaneously could race on `ports.json`. Mitigation: file locking (same pattern as session counter).
