# host-broker Specification

## Purpose
Provides a host-side HTTP broker that serves kit-contributed routes to containers over an authenticated, container-scoped channel reachable via `host.docker.internal`.

## Requirements

### Requirement: Host broker HTTP server
The system SHALL provide a host-side HTTP server (the "broker") that serves routes contributed by enabled kits. The broker SHALL bind an address reachable from the container via `host.docker.internal` (`0.0.0.0`), and SHALL start only when at least one route is registered.

#### Scenario: Broker starts when a kit registers a route
- **WHEN** a container is started for a project with at least one broker-using kit enabled
- **THEN** the broker is running on the host and serving that kit's routes

#### Scenario: No broker without routes
- **WHEN** no enabled kit registers a broker route
- **THEN** no broker server is started

### Requirement: Per-container token authentication
The broker SHALL authenticate every request with a per-container secret token generated at container creation. Requests SHALL present the token as an `Authorization: Bearer <token>` header, and the broker SHALL compare it in constant time. Requests with a missing or incorrect token SHALL be rejected with `401` and SHALL NOT reach any route handler.

#### Scenario: Valid token accepted
- **WHEN** a request carries the container's token
- **THEN** the broker dispatches it to the matching route handler

#### Scenario: Missing or wrong token rejected
- **WHEN** a request omits the token or carries a different container's token
- **THEN** the broker responds `401` and no handler runs

### Requirement: Broker connection parameters in container environment
At container creation the system SHALL bake the broker host, port, and token into the container environment as `ASYLUM_BROKER_HOST`, `ASYLUM_BROKER_PORT`, and `ASYLUM_BROKER_TOKEN`. These values SHALL be stable for the container's lifetime and SHALL be recoverable from the running container (e.g. via `docker inspect`).

#### Scenario: Environment available to sessions
- **WHEN** any session (agent, shell, or run) is attached to the container
- **THEN** `ASYLUM_BROKER_HOST`, `ASYLUM_BROKER_PORT`, and `ASYLUM_BROKER_TOKEN` are present in its environment

### Requirement: Container-scoped broker lifetime
The broker's lifetime SHALL be tied to the container, not to any single session. The broker SHALL run as a detached host process that terminates when the container stops. It SHALL NOT be terminated when the session that started it exits while the container is still running.

#### Scenario: Broker survives the starting session
- **WHEN** the session that started the broker exits but the container keeps running with other sessions attached
- **THEN** the broker keeps serving requests

#### Scenario: Broker stops with the container
- **WHEN** the container stops (normal exit, cleanup, or rebuild)
- **THEN** the broker process terminates

### Requirement: Idempotent broker start and self-healing
Every asylum session SHALL ensure a broker is running for its container, on both the container-create path and the attach path. Ensuring SHALL be idempotent: if a live broker already answers, no second broker is started; if none answers, one SHALL be started. A second concurrent start attempt SHALL NOT produce a second serving broker.

#### Scenario: Attach does not double-start
- **WHEN** a session attaches to a container whose broker is already running
- **THEN** no additional broker process serves and the existing broker continues

#### Scenario: Respawn after broker death
- **WHEN** a session attaches to a running container whose broker is no longer answering
- **THEN** the session starts a new broker bound to the container's baked port

### Requirement: Kit-contributed routes
The kit system SHALL let a kit contribute broker routes (path and handler). The broker SHALL mount the routes of all enabled kits under token authentication. Route handlers SHALL execute on the host.

#### Scenario: Enabled kit's routes are served
- **WHEN** an enabled kit contributes a route
- **THEN** the broker serves that route under token authentication

#### Scenario: Disabled kit's routes are not served
- **WHEN** a kit that contributes a route is not enabled for the project
- **THEN** the broker does not serve that route
