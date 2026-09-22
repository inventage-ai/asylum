# iterm-status-kit Specification

## Purpose
Lets a Claude Code session running inside an asylum container drive the iTerm2 status bar of the terminal it is attached to, by forwarding its hook payloads to the host's own iTerm2 status binary through the host broker, so a sandboxed session shows the same working/idle state and tool detail a host session does.
## Requirements
### Requirement: Opt-in iTerm status kit

The system SHALL provide an `iterm` kit that is inactive unless the user explicitly enables it in project or user configuration. The kit SHALL contribute a broker route, a container-side shim, and a rules snippet describing the integration to the agent.

#### Scenario: Inactive without configuration

- **WHEN** a project has no `iterm` entry in its configuration
- **THEN** the kit contributes no route, no mount, and no environment, and the container behaves as it does today

#### Scenario: Active when enabled

- **WHEN** a project enables the `iterm` kit
- **THEN** the broker serves the status route and the container carries the shim

### Requirement: Container shim occupies the host integration's hook path

iTerm2 configures its Claude Code integration by writing hook entries into the user's Claude settings that name a fixed executable path. Under shared agent config isolation the container reads those same settings, and the named executable is a macOS binary that cannot run on Linux. When the kit is active the system SHALL place an executable shim at that same path inside the container, so the hook resolves to the shim in the container and to the host binary on the host.

The system SHALL NOT modify the user's Claude settings.

#### Scenario: Hook path resolves to the shim in the container

- **WHEN** the kit is active and a hook configured by iTerm2 fires inside the container
- **THEN** the shim runs instead of the host binary, and no `exec format error` occurs

#### Scenario: Host sessions are unaffected

- **WHEN** the kit is active and the user runs Claude Code directly on the host
- **THEN** the same hook path resolves to the host binary and the host integration behaves as it did before

#### Scenario: Settings are not rewritten

- **WHEN** the kit is active
- **THEN** the user's Claude settings file is unchanged

### Requirement: The shim never blocks or fails a session

A status indicator MUST NOT be able to interfere with the work it reports on. The shim SHALL exit with status `0` regardless of what the broker does, SHALL bound the time it waits for a response, and SHALL produce no output on the hook's stdout or stderr.

#### Scenario: Broker unreachable

- **WHEN** a hook fires and the broker does not answer
- **THEN** the shim exits `0` within its timeout and the agent's action proceeds normally

#### Scenario: Route returns an error

- **WHEN** the route responds with an error status
- **THEN** the shim still exits `0` and the agent's action proceeds normally

### Requirement: Host status route runs the host status binary

The system SHALL serve a broker route that accepts a hook payload from the container and runs the host's iTerm2 status binary with that payload on standard input and the calling session's identity in its environment.

The route SHALL pass no value received from the container as a command-line argument to the host process. The executable path SHALL be fixed by the system, not supplied by the request. The request body SHALL be size-capped.

#### Scenario: Payload reaches the host binary

- **WHEN** the container posts a hook payload to the route with a valid session envelope
- **THEN** the host status binary runs with that payload on standard input and the iTerm2 status bar for that session updates

#### Scenario: Container input never becomes an argument

- **WHEN** the container posts a payload containing values that resemble command-line flags
- **THEN** the host process is invoked with the system's fixed argument list and the values appear only on standard input

#### Scenario: Oversized payload rejected

- **WHEN** the container posts a body larger than the cap
- **THEN** the route rejects the request and no host process runs

### Requirement: Status updates preserve event order

Hook events describe a sequence of state transitions, and applying them out of order leaves the status bar showing a state the session is not in. The route SHALL complete each status update before accepting the effect of a later one for the same session.

#### Scenario: Rapid consecutive events

- **WHEN** a session emits a tool-start event immediately followed by a tool-end event
- **THEN** the status bar ends in the state described by the later event

### Requirement: Integration disables itself outside iTerm2

The integration depends on the host terminal being an iTerm2 session. When the session has no iTerm2 identity the system SHALL inject no session envelope, and the shim SHALL do nothing.

#### Scenario: Started from a non-iTerm2 terminal

- **WHEN** a session is started from a terminal that is not iTerm2, or from a non-interactive context
- **THEN** no session envelope is injected, the shim performs no request, and hooks succeed silently

#### Scenario: Host without the integration installed

- **WHEN** the kit is enabled on a host where the iTerm2 status binary is not present
- **THEN** the route reports the integration unavailable and the session is otherwise unaffected

