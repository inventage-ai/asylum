## Why

Agents running inside the sandbox frequently start web servers (dev servers, preview servers, API endpoints) that the user needs to access from their browser. Currently, users must manually configure port forwarding via `ports:` config or `-p` flags, requiring them to know in advance which ports the agent will use. A default-on `ports` kit that pre-allocates a range of forwarded ports per project eliminates this friction — the agent can pick any of its assigned ports, and the user can immediately access the service.

## What Changes

- New `ports` kit (`internal/kit/ports.go`), default-on, that allocates a configurable number of host ports (default: 5) per project and forwards them to the container.
- Global port registry at `~/.asylum/ports.json` tracking which host port ranges are assigned to which project. Ranges are never reused by another project while assigned.
- Port allocation happens at container start: if the project already has an assigned range, reuse it; otherwise allocate the next available range from a high starting port (e.g., 10000+).
- The kit provides a `RulesSnippet` telling the agent which ports are available and how the user can access them (e.g., `http://localhost:<port>`).
- The number of ports is configurable via `KitConfig` (e.g., `kits: { ports: { count: 10 } }`).

## Capabilities

### New Capabilities
- `port-allocation`: Global port range allocation, per-project assignment, and container port forwarding via the ports kit.

### Modified Capabilities
- `sandbox-rules`: The sandbox rules file gains a dynamic section showing the project's allocated ports.

## Impact

- `internal/kit/ports.go`: New kit definition with `RulesSnippet` generation.
- `internal/container/container.go`: Port allocation integrates with `RunArgs` to inject `-p` flags for the allocated range.
- `~/.asylum/ports.json`: New global state file for port assignments.
- `internal/config/config.go`: New `KitConfig` accessor for port count.
- `cmd/asylum/main.go`: Wiring port allocation into the container start flow.
