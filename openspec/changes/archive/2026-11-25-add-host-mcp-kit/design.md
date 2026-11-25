# Design

## Context

- **Broker:** The host broker (`internal/broker`) is an `http.Serve` server with bearer-token auth. It runs once per container as a detached `asylum __broker` process (`Setsid`) and exits via `os.Exit` when `docker wait` returns. It inherits the full `os.Environ()` of the session that spawned it.
- **Kit config for the broker:** Today it reaches the broker only through spawn-time env (`brokerEnv`), which is derived from the merged config.
- **Container-side shims:** `asylum-open` and the others are `curl` scripts embedded in a `DockerSnippet`. `curl` cannot carry a full-duplex stream.
- **Import direction:** `internal/config` imports `internal/kit`, so kit code cannot use `config.LoadFile`. The config loader does not parse YAML strictly, so unknown keys under a kit are ignored.
- **Project directory:** The container name is derived from the project directory, but the broker only receives the container name.

## Goals / Non-Goals

**Goals:**
- One generic mechanism for any stdio MCP server that has to run on the host
- The agent cannot define or change what runs on the host through files it can write (within the limits of #47)

**Non-Goals:**
- Exposing HTTP or SSE MCP servers from the host
- Sharing one server process across sessions
- Writing MCP registrations into agent configs
- Confining what a server does once it runs. The kit controls who can start a server, not what the server does.

## Decisions

### Shim is a Node script
`asylum-mcp` is a small Node script embedded in the kit's `DockerSnippet`. It sends `GET /mcp/<name>` with `Connection: Upgrade`, `Upgrade: asylum-stdio` and the bearer token. It uses `socketPath` when `ASYLUM_BROKER_SOCK` is set, else `ASYLUM_BROKER_HOST:PORT`. On `upgrade` it pipes stdin→socket and socket→stdout. On a non-101 response it writes the body to stderr and exits 1.
- *Why:* Node is always present (the `node` kit is always-on, and the kit declares `Deps: ["node"]`). Node's `http` client supports upgrades on both transports, so no new binary is needed.
- *Alternatives:*
  - `socat`: The HTTP handshake would have to be built by hand, and response headers could leak into stdout.
  - Cross-compiled Go shim binary: needs per-arch builds and a delivery path into the image.
  - Plain HTTP MCP transport: Clients cannot dial a Unix socket, and the token would end up in agent config.

### Streaming via hijack inside the route handler
The `/mcp/` handler validates the request, spawns the server, then hijacks the connection through `http.Hijacker`. It writes the `101` response itself and copies bytes in both directions. The `Route` type stays unchanged. Auth still runs first, through `authWrap`.
- *Alternative:* A dedicated streaming route type in the broker. It adds an API for a single user.

### Spawn before upgrade
The handler starts the process before it sends `101`. Errors can then still be reported as plain HTTP statuses, which the shim prints:
- unknown server: `404`
- project not allowed or invalid command: `403`
- `exec` failure: `502`

### Handler reads the global config itself, per connection
At each request the handler reads `~/.asylum/config.yaml` with `yaml.v3` into a minimal struct (`kits.host-mcp.servers`). It never sees project layers, and edits apply without a container restart. No field is added to `config.KitConfig`, so the merged config never carries server definitions.
- *Alternatives:*
  - Spawn-time env: It would freeze the definitions and need a second, global-only config path in the session anyway.
  - `config.LoadFile`: blocked by the import cycle.

### Project directory passed to the broker
`ensureBroker` passes `--project <dir>` to `__broker`. `broker.Ctx` gains `ProjectDir()`. The handler uses it as the working directory and to check the `projects` allowlist (exact match on the cleaned absolute path).
- *Alternative:* The sealed session envelope. It is not needed, because the project directory is per container, not per session.

### Process lifetime
- **Process group:** Each server starts with `Setpgid: true`.
- **Connection close:** When the server's stdout ends, teardown starts immediately. When the client closes its side, the handler closes the server's stdin and gives the server 2 s to flush and exit, because closing stdin is how an MCP stdio client ends a session. Teardown then sends `SIGTERM` to `-pgid`, and `SIGKILL` after another 2 s.
- **Broker exit:** The handler registers each process group in a broker-wide set. The `docker wait` goroutine kills every registered group before `os.Exit`.

### Session-only registration via `--mcp-config`
For Claude sessions, the host-side asylum process reads the allowed servers when it builds the exec command. It uses the same global-config read and the same validation as the broker. It passes them as `--mcp-config '{"mcpServers":{"<name>":{"type":"stdio","command":"asylum-mcp","args":["<name>"]}}}'`. `CmdOpts` gains a list of server names, and the agent turns it into its own flag. This follows the existing `KitSkillsDir` → `--add-dir` path. `--mcp-config` is variadic, so it goes before `--add-dir` and after passthrough arguments.
- *Why:* Nothing is persisted, so removed servers disappear with the next session. It works in every isolation mode, and nothing lands in the repo or in a Claude config shared across containers.
- *Alternatives:*
  - `claude mcp add` during onboarding or in the entrypoint: persists into `~/.claude/.claude.json`, which `shared` mode shares across all containers, and leaves stale entries behind.
  - Project `.mcp.json`: lives in the repo and triggers Claude's approval prompt.
- *Limits:* Only sessions asylum starts get the servers. `claude` launched by hand inside `asylum shell` does not. A server that is also registered by hand appears twice. The docs tell users to remove the manual entry.

### Environment and logging
- **Environment:** The server environment is built from scratch: `HOME`, `PATH`, `LANG`, `TMPDIR`, `USER`, `LOGNAME` from the broker's environment, then the configured `env` on top.
- **Stderr:** Stderr is opened in append mode on `~/.asylum/projects/<container>/host-mcp.log`. Each spawn writes a header line with the timestamp and server name.

## Risks / Trade-offs

- **Project config is still writable by the agent (#47).** Through `volumes`, an agent could mount `~/.asylum` writable and change the global config. → The docs reference #47. This kit does not widen the gap.
- **The server itself can be steered by the agent** (project-local config, LLDB in XcodeBuildMCP). → The docs state this plainly. Users pin settings via `env` where the server allows it.
- **The agent can start any configured server**, because the token is in its environment. → `projects` allowlist, opt-in kit, no default servers.
- **`node` may be missing from `PATH`** for agents that spawn MCP servers with a minimal environment. → Verify with Claude, Codex and Gemini during implementation. If needed, the shim falls back to fnm's default alias path.
- **Orphaned servers if the broker is killed with `SIGKILL`.** → Accepted. The process group makes manual cleanup easy.
- **Config tooling might drop the unknown `servers` key** when it rewrites `~/.asylum/config.yaml` (kit sync, migration, `asylum config`). → Verify round-trips, and fix them if needed.
