## MODIFIED Requirements

### Requirement: Host broker HTTP server
The system SHALL provide a host-side HTTP server (the "broker") that serves routes contributed by enabled kits. The broker SHALL bind a host-only endpoint chosen for the host platform — a Unix domain socket on a native (shared-kernel) Linux engine, or a `127.0.0.1` TCP address on a VM-backed engine (Docker Desktop and macOS) — reachable from the container via `host.docker.internal` or a bind-mounted socket. The broker SHALL NOT bind an address reachable beyond the host (no `0.0.0.0`). It SHALL start only when at least one route is registered.

#### Scenario: Broker starts when a kit registers a route
- **WHEN** a container is started for a project with at least one broker-using kit enabled
- **THEN** the broker is running on the host and serving that kit's routes

#### Scenario: No broker without routes
- **WHEN** no enabled kit registers a broker route
- **THEN** no broker server is started

#### Scenario: Not reachable beyond the host
- **WHEN** the broker is running
- **THEN** it is bound to a Unix socket or `127.0.0.1`, and no host network interface exposes the broker port to other hosts

### Requirement: Broker connection parameters in container environment
At container creation the system SHALL bake the broker connection parameters and token into the container environment. For the Unix-socket transport it SHALL set `ASYLUM_BROKER_SOCK` to the container-side socket path; for the TCP transport it SHALL set `ASYLUM_BROKER_HOST` and `ASYLUM_BROKER_PORT`. `ASYLUM_BROKER_TOKEN` SHALL be set for both. These values SHALL be stable for the container's lifetime and SHALL be recoverable from the running container (e.g. via `docker inspect`).

#### Scenario: Unix-socket parameters
- **WHEN** the broker uses the Unix-socket transport
- **THEN** the container environment has `ASYLUM_BROKER_SOCK` and `ASYLUM_BROKER_TOKEN`

#### Scenario: TCP parameters
- **WHEN** the broker uses the TCP transport
- **THEN** the container environment has `ASYLUM_BROKER_HOST`, `ASYLUM_BROKER_PORT`, and `ASYLUM_BROKER_TOKEN`

### Requirement: Idempotent broker start and self-healing
Every asylum session SHALL ensure a broker is running for its container, on both the container-create path and the attach path. Ensuring SHALL be idempotent: if a live broker already answers, no second broker is started; if none answers, one SHALL be started. A second concurrent start attempt SHALL NOT produce a second serving broker.

#### Scenario: Attach does not double-start
- **WHEN** a session attaches to a container whose broker is already running
- **THEN** no additional broker process serves and the existing broker continues

#### Scenario: Respawn after broker death
- **WHEN** a session attaches to a running container whose broker is no longer answering
- **THEN** the session starts a new broker bound to the container's baked endpoint

## ADDED Requirements

### Requirement: Platform-aware transport selection
The system SHALL choose the broker transport from the host platform. On a native (shared-kernel) Linux engine it SHALL use a Unix domain socket whose containing directory is bind-mounted into that container only. On a VM-backed engine — always on macOS, and on Linux when the engine reports as Docker Desktop — it SHALL use a `127.0.0.1` TCP endpoint reached via `host.docker.internal`. A misdetected transport SHALL degrade to the feature being unavailable and SHALL NOT cause the broker to bind an externally reachable address.

#### Scenario: Native Linux uses a Unix socket
- **WHEN** the host runs a native Linux Docker engine
- **THEN** the broker listens on a Unix socket and the socket directory is bind-mounted into the container

#### Scenario: Docker Desktop uses loopback TCP
- **WHEN** the host runs Docker Desktop (macOS, or a Linux engine reporting as Docker Desktop)
- **THEN** the broker listens on `127.0.0.1` and the container reaches it via `host.docker.internal`

#### Scenario: Misdetection stays safe
- **WHEN** the transport does not match the actual engine and the container cannot reach the broker
- **THEN** the open request fails without any externally reachable bind, and callers fall back to printing the URL
