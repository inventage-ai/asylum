# claude-ide-integration Specification

## Purpose
Lets a Claude Code session running inside an asylum container connect to an IDE running on the host, so `/ide` provides selection context, diagnostics, and the diff view from within the sandbox.
## Requirements
### Requirement: Claude containers resolve the IDE host to the host machine

Claude Code discovers an IDE from lock files in its configuration directory and dials the discovered port on a host it resolves at connection time, defaulting to the container's own loopback address where no IDE listens. The system SHALL contribute an environment variable to every Claude container that redirects this resolution to `host.docker.internal`, the address at which the host machine is reachable from the container.

The variable SHALL be contributed unconditionally for the Claude agent. On engines where `host.docker.internal` does not reach host loopback services, the connection fails exactly as it does without the variable, so no engine detection is required.

#### Scenario: Claude container receives the IDE host override

- **WHEN** a container is started with `claude` as the agent
- **THEN** the container environment contains `CLAUDE_CODE_IDE_HOST_OVERRIDE` set to `host.docker.internal`

#### Scenario: Other agents are unaffected

- **WHEN** a container is started with an agent other than `claude`
- **THEN** no `CLAUDE_CODE_IDE_HOST_OVERRIDE` variable is contributed by that agent

#### Scenario: IDE running on the host is reachable

- **WHEN** an IDE on the host publishes a lock file that the container can read, the container runs on an engine whose `host.docker.internal` reaches host loopback services, and the user connects via `/ide`
- **THEN** the session connects to the IDE's server on the host and IDE-provided tools become available

### Requirement: Documented preconditions for IDE integration

IDE integration depends on conditions the system does not enforce: the container must read the same lock files the host IDE writes, which holds only under `shared` agent config isolation, and the container engine must route `host.docker.internal` to host loopback services, which holds on Docker Desktop but not on a native Linux engine. The in-container reference documentation SHALL state that `/ide` is supported and name both preconditions, so a user whose `/ide` list is empty can identify the cause without inspecting asylum's source.

#### Scenario: Reference documentation covers IDE integration

- **WHEN** a user reads the in-container asylum reference documentation
- **THEN** it states that `/ide` connects to an IDE running on the host, and names both the `shared` agent config isolation requirement and the Docker Desktop engine requirement

