## MODIFIED Requirements

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

## ADDED Requirements

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
