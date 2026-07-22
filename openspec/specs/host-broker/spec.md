# host-broker Specification

## Purpose
Provides a host-side HTTP broker that serves kit-contributed routes to containers over an authenticated, container-scoped channel reachable via `host.docker.internal`.

## Requirements

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

### Requirement: Per-container token authentication
The broker SHALL authenticate every request with a per-container secret token generated at container creation. Requests SHALL present the token as an `Authorization: Bearer <token>` header, and the broker SHALL compare it in constant time. Requests with a missing or incorrect token SHALL be rejected with `401` and SHALL NOT reach any route handler.

#### Scenario: Valid token accepted
- **WHEN** a request carries the container's token
- **THEN** the broker dispatches it to the matching route handler

#### Scenario: Missing or wrong token rejected
- **WHEN** a request omits the token or carries a different container's token
- **THEN** the broker responds `401` and no handler runs

### Requirement: Broker connection parameters in container environment
At container creation the system SHALL bake the broker connection parameters and token into the container environment. For the Unix-socket transport it SHALL set `ASYLUM_BROKER_SOCK` to the container-side socket path; for the TCP transport it SHALL set `ASYLUM_BROKER_HOST` and `ASYLUM_BROKER_PORT`. `ASYLUM_BROKER_TOKEN` SHALL be set for both. These values SHALL be stable for the container's lifetime and SHALL be recoverable from the running container (e.g. via `docker inspect`).

#### Scenario: Unix-socket parameters
- **WHEN** the broker uses the Unix-socket transport
- **THEN** the container environment has `ASYLUM_BROKER_SOCK` and `ASYLUM_BROKER_TOKEN`

#### Scenario: TCP parameters
- **WHEN** the broker uses the TCP transport
- **THEN** the container environment has `ASYLUM_BROKER_HOST`, `ASYLUM_BROKER_PORT`, and `ASYLUM_BROKER_TOKEN`

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
- **THEN** the session starts a new broker bound to the container's baked endpoint

### Requirement: Kit-contributed routes
The kit system SHALL let a kit contribute broker routes (path and handler). The broker SHALL mount the routes of all enabled kits under token authentication. Route handlers SHALL execute on the host and SHALL receive a broker-supplied context exposing the container's name and a request to forward a host loopback port into the container.

#### Scenario: Enabled kit's routes are served
- **WHEN** an enabled kit contributes a route
- **THEN** the broker serves that route under token authentication

#### Scenario: Disabled kit's routes are not served
- **WHEN** a kit that contributes a route is not enabled for the project
- **THEN** the broker does not serve that route

#### Scenario: Handler receives broker context
- **WHEN** the broker dispatches a request to a route handler
- **THEN** the handler is given the container name and a loopback-forward request bound to that container

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

### Requirement: Loopback callback forwarding
The broker SHALL provide a best-effort service that forwards a host loopback port to the same port on the container's loopback, so an OAuth callback opened in the host browser reaches a callback server listening inside the container. The forward SHALL bind only a loopback address on the host (`127.0.0.1`, or `[::1]` for an IPv6 request) and SHALL relay each connection into the container over `docker exec` (no host port publishing and no sidecar container). If the host port cannot be bound, the broker SHALL skip the forward without failing the caller.

#### Scenario: Forward established for a container callback
- **WHEN** a forward is requested for loopback port `P` and the host can bind it
- **THEN** connections to the host's loopback `P` are relayed to the container's loopback `P`

#### Scenario: Host port unavailable
- **WHEN** a forward is requested for a port the host cannot bind
- **THEN** the broker logs and skips the forward, and the triggering request still succeeds

#### Scenario: IPv6 loopback
- **WHEN** a forward is requested for an IPv6 loopback (`::1`) port
- **THEN** the broker binds `[::1]` on the host and relays to the container's IPv6 loopback

### Requirement: Time-boxed forwards
Each loopback forward SHALL remain active for five minutes and SHALL be extended by another five minutes each time the same port is requested again. On expiry the broker SHALL stop listening on that host port. All forwards SHALL be torn down when the broker stops (i.e. when the container stops).

#### Scenario: Forward expires
- **WHEN** five minutes pass with no further request for a forwarded port
- **THEN** the broker stops listening on that host port

#### Scenario: Timer reset on repeat request
- **WHEN** a forward for a port is requested again before it expires
- **THEN** its lifetime is extended by five minutes from the repeat request

#### Scenario: Teardown with the container
- **WHEN** the container stops
- **THEN** all of its forwards are torn down
