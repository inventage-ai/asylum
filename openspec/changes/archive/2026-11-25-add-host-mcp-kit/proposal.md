# Proposal

## Why

Some MCP servers can only run on the host: XcodeBuildMCP and `xcrun mcpbridge` need Xcode and the Simulator, and others need host apps or the host keychain. Today the workaround is a manual `mcp-proxy` on the host, reached via `host.docker.internal`. It has no authentication, it has to be started and stopped by hand, and each server needs its own fixed port (#49).

## What Changes

- New opt-in `host-mcp` kit. It runs stdio MCP servers that the user configures in `~/.asylum/config.yaml` on the host, and relays their stdio into the container through the existing host broker.
- New container-side shim `asylum-mcp <name>`, registered like any local stdio MCP server (`claude mcp add <name> -- asylum-mcp <name>`). It reaches the broker with `$ASYLUM_BROKER_TOKEN` over the Unix socket or TCP, whichever the container uses.
- Each shim connection starts its own server process on the host, in the project directory. The process ends when the connection ends or the broker stops.
- Security constraints:
  - Server definitions come only from the global config, never from `.asylum` or `.asylum.local`.
  - `command[0]` must be an absolute path.
  - Servers get a minimal environment plus their configured `env`.
  - An optional `projects` allowlist limits a server to certain projects.
- Server stderr is written to a per-container log file on the host.
- Claude sessions that asylum starts get the servers that are allowed for the project automatically, as a session-only `--mcp-config`. Nothing is written into agent configs. For other agents, or for `claude` started by hand in `asylum shell`, the kit's sandbox rules explain how to register a server with `asylum-mcp <name>`.
- The host broker gains a streaming (connection-upgrade) route type. Handlers may also read the user's global config at request time instead of receiving config at spawn time.

## Capabilities

### New Capabilities
- `host-mcp-kit`: configuring host MCP servers, the container shim, and the per-connection server lifecycle, environment, logging, and trust rules

### Modified Capabilities
- `host-broker`: Route handlers may read configuration from the user's global config file at request time. That configuration is then not fixed for the container's lifetime. The broker also supports routes that upgrade a request to a bidirectional byte stream.

## Impact

- `internal/kit/`: new `host_mcp.go` kit (shim in `DockerSnippet`, rules snippet, broker route)
- `internal/broker/`: streaming route support; the broker is told the project directory
- `cmd/asylum/main.go`: passes the project directory to `__broker`
- `internal/agent/`, `internal/container/`: Claude command gets `--mcp-config` for the allowed host MCP servers
- `internal/config/`: config schema for `host-mcp` servers; a way to load the global layer on its own
- `docs/kits/host-mcp.md`, `CHANGELOG.md`
- Related: #47. Project config is writable by the agent. The "global config only" guarantee holds fully only once #47 is fixed.
