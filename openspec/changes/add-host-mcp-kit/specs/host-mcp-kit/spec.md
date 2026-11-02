# Spec Delta

## Purpose

Lets agents in the container use stdio MCP servers that can only run on the host (for example XcodeBuildMCP). The broker relays their stdio over the authenticated host broker, and the servers are started on demand per connection.

## ADDED Requirements

### Requirement: Opt-in kit
The system SHALL provide a `host-mcp` kit that is inactive unless the user enables it. No MCP server SHALL be available through the kit unless one is defined in the user's global configuration.

#### Scenario: Kit not enabled
- **WHEN** the `host-mcp` kit is not enabled for a project
- **THEN** the container has no `asylum-mcp` shim, and the broker serves no host MCP route

#### Scenario: Enabled without servers
- **WHEN** the kit is enabled but no servers are defined in `~/.asylum/config.yaml`
- **THEN** every `asylum-mcp <name>` invocation fails with an error naming the unknown server

### Requirement: Server definitions from global config only
Host MCP servers SHALL be defined under `kits.host-mcp.servers` in `~/.asylum/config.yaml` only. Each entry has a name, a `command` (argument list), an optional `env` map and an optional `projects` list. Server definitions in `.asylum`, `.asylum.local` or CLI flags SHALL be ignored and SHALL NOT add, replace or alter a server.

#### Scenario: Project config cannot define a server
- **WHEN** a project `.asylum` defines `kits.host-mcp.servers.evil` with a command
- **THEN** `asylum-mcp evil` fails as an unknown server, and no host process is started

#### Scenario: Project config cannot alter a server
- **WHEN** the global config defines server `xcodebuild` and a project `.asylum` sets a different `command` or `env` for it
- **THEN** the server runs with the global definition only

### Requirement: Absolute command path
A server's `command[0]` SHALL be an absolute path to an executable. The broker SHALL NOT resolve the command through `PATH`. A server whose `command[0]` is missing or not absolute SHALL be rejected when a connection requests it, and no host process SHALL start.

#### Scenario: Relative command rejected
- **WHEN** a server is defined with `command: ["npx", "some-mcp"]` and the shim requests it
- **THEN** the request fails with an error stating that the command must be an absolute path, and nothing is spawned

### Requirement: Project allowlist
When a server defines `projects`, the broker SHALL start it only for containers whose project directory is one of the listed absolute paths. When `projects` is absent, the server SHALL be available to every project that enables the kit.

#### Scenario: Project not allowed
- **WHEN** server `xcodebuild` lists `projects: ["/code/app"]` and the shim requests it from a container for `/code/other`
- **THEN** the request is rejected, and no host process starts

#### Scenario: No allowlist
- **WHEN** a server defines no `projects`
- **THEN** any project with the kit enabled can start it

### Requirement: Container shim
The kit SHALL install an `asylum-mcp` command in the container. `asylum-mcp <name>` SHALL connect to the broker with the container's token over the transport described by the container environment, request server `<name>`, and relay its own stdin to the server's stdin and the server's stdout to its own stdout without modifying the bytes. If the broker rejects the request, the shim SHALL print the reason to stderr and exit non-zero.

#### Scenario: Registered as a stdio MCP server
- **WHEN** the agent registers `asylum-mcp xcodebuild` as a stdio MCP server and starts a session
- **THEN** the agent talks MCP to the host's XcodeBuildMCP process through the shim

#### Scenario: Both transports
- **WHEN** the container uses the Unix-socket transport, or the TCP transport
- **THEN** the shim reaches the broker in either case without extra configuration

#### Scenario: Rejection reported
- **WHEN** the shim requests an unknown or disallowed server
- **THEN** it writes the broker's reason to stderr and exits with a non-zero status

### Requirement: Per-connection server process
Each shim connection SHALL start its own server process on the host, with the project directory as the working directory. The process SHALL be terminated, including any child processes it started, when the shim connection closes, or when the broker stops.

#### Scenario: Connection closed
- **WHEN** the agent session ends and the shim's connection closes
- **THEN** the host server process and its children are terminated

#### Scenario: Container stops
- **WHEN** the container stops while a host MCP server is running
- **THEN** the server process and its children are terminated

#### Scenario: Concurrent sessions
- **WHEN** two sessions in the same container each start `asylum-mcp xcodebuild`
- **THEN** two independent server processes run, one per connection

### Requirement: Minimal server environment
A spawned server SHALL receive only `HOME`, `PATH`, `LANG`, `TMPDIR`, `USER` and `LOGNAME`, each taken from the broker's environment when set, plus the server's configured `env`, which overrides those on conflict. The broker token and any other broker environment SHALL NOT be passed to the server.

#### Scenario: Token not leaked
- **WHEN** a host MCP server is spawned
- **THEN** its environment does not contain `ASYLUM_BROKER_TOKEN`

#### Scenario: Configured env applied
- **WHEN** a server defines `env: { XCODEBUILDMCP_SENTRY_DISABLED: "true" }`
- **THEN** the spawned process sees that variable

### Requirement: Server stderr log
A server's stderr SHALL be appended to a log file in the container's directory under `~/.asylum/projects/<container>/` on the host. Stderr SHALL NOT be relayed into the shim's stdout.

#### Scenario: Server writes to stderr
- **WHEN** a host MCP server writes diagnostics to stderr
- **THEN** they appear in the per-container log file, and the MCP stream on stdout is unaffected

### Requirement: Sandbox rules describe the servers
When the kit is active, the sandbox rules SHALL explain how to register a host MCP server with `asylum-mcp <name>`. The system SHALL NOT write MCP registrations into any agent configuration.

#### Scenario: Agent learns the registration command
- **WHEN** the kit is active
- **THEN** the sandbox rules contain the `asylum-mcp <name>` registration instructions

### Requirement: Automatic registration in Claude sessions
When the kit is active and asylum starts a Claude session, asylum SHALL register every server from the global config that the project may start, for that session only, as a stdio server named after the server and run as `asylum-mcp <name>`. Servers the project may not start, or whose command is invalid, SHALL NOT be registered. The registration SHALL NOT be persisted in any file.

#### Scenario: Allowed servers registered
- **WHEN** the global config defines `xcodebuild` and `cupertino`, and asylum starts a Claude session for a project that may start both
- **THEN** both servers are available in that session without any manual `claude mcp add`

#### Scenario: Disallowed server not registered
- **WHEN** a server lists `projects` that do not include the session's project
- **THEN** the server is not registered for that session

#### Scenario: Nothing persisted
- **WHEN** a server is removed from the global config
- **THEN** the next Claude session asylum starts no longer has it, and no agent configuration file has to be cleaned up

#### Scenario: Kit inactive or no servers
- **WHEN** the kit is not active, or no server is allowed for the project
- **THEN** the Claude command line carries no MCP configuration from asylum

#### Scenario: Unreadable global config
- **WHEN** the global config cannot be read or parsed while a session starts
- **THEN** asylum warns and starts the session without host MCP servers
