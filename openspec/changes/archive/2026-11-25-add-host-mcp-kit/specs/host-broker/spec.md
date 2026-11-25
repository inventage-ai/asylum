# Spec Delta

## MODIFIED Requirements

### Requirement: Kit configuration available to route handlers
Because the broker runs as a detached host process that does not read project configuration itself, the session that starts a broker SHALL pass the effective kit configuration its route handlers depend on to that process, and route handlers SHALL validate requests against the values passed at spawn time. The values SHALL be derived from the same merged configuration the session uses. A broker started without such values SHALL fall back to its handlers' built-in defaults rather than failing to start. As an exception, a route handler whose configuration must come from the user's global configuration only SHALL read `~/.asylum/config.yaml` itself at request time and SHALL NOT use project configuration layers.

#### Scenario: Handler sees configured values
- **WHEN** a project configures a kit option that its broker route depends on and a container is started
- **THEN** the broker process serving that container validates requests using the configured value

#### Scenario: Missing configuration falls back to defaults
- **WHEN** a broker process is started with no kit configuration values
- **THEN** it serves its routes using each handler's default behavior

#### Scenario: Global-only handler ignores project layers
- **WHEN** a handler reads its configuration from the global config, and a project `.asylum` sets a value for the same kit
- **THEN** the handler uses only the value from `~/.asylum/config.yaml`

### Requirement: Broker configuration fixed for the container's lifetime
A running broker SHALL keep the configuration it was started with. Because a live broker is not restarted while its container runs, changes to kit configuration that affect broker route handlers SHALL take effect for the next container start. Handlers that read the global configuration at request time are excluded: they SHALL see changes to `~/.asylum/config.yaml` on the next request.

#### Scenario: Config change during a running container
- **WHEN** kit configuration affecting a broker route is edited while the container and its broker are running
- **THEN** the running broker keeps its previous configuration and the new value applies once the container is restarted

#### Scenario: Global config change seen by a request-time handler
- **WHEN** the user edits `~/.asylum/config.yaml` while the container runs, and a request-time handler serves the next request
- **THEN** that request is handled with the edited configuration

## ADDED Requirements

### Requirement: Streaming routes
The broker SHALL support routes that, after authentication, upgrade the HTTP connection to a raw bidirectional byte stream. A streaming route SHALL be able to reject a request with a regular HTTP error status before upgrading. The stream SHALL work over both the Unix-socket and the TCP transport.

#### Scenario: Upgrade after authentication
- **WHEN** an authenticated request asks a streaming route to upgrade
- **THEN** the broker answers `101 Switching Protocols` and relays bytes in both directions until either side closes

#### Scenario: Rejection before upgrade
- **WHEN** a streaming route refuses a request
- **THEN** the broker answers with an HTTP error status and a reason, and does not upgrade the connection

#### Scenario: Unauthenticated upgrade refused
- **WHEN** an upgrade request carries no or a wrong token
- **THEN** the broker responds `401`, and the route handler does not run

### Requirement: Project directory known to the broker
The session that starts a broker SHALL pass the container's project directory to the broker process, so that route handlers can use it as a working directory and for per-project decisions.

#### Scenario: Handler uses the project directory
- **WHEN** a route handler needs the project directory of the container it serves
- **THEN** the broker context provides the host path of that project directory
