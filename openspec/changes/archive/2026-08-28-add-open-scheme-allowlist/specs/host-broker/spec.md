## ADDED Requirements

### Requirement: Kit configuration available to route handlers
Because the broker runs as a detached host process that does not read project configuration itself, the session that starts a broker SHALL pass the effective kit configuration its route handlers depend on to that process, and route handlers SHALL validate requests against the values passed at spawn time. The values SHALL be derived from the same merged configuration the session uses. A broker started without such values SHALL fall back to its handlers' built-in defaults rather than failing to start.

#### Scenario: Handler sees configured values
- **WHEN** a project configures a kit option that its broker route depends on and a container is started
- **THEN** the broker process serving that container validates requests using the configured value

#### Scenario: Missing configuration falls back to defaults
- **WHEN** a broker process is started with no kit configuration values
- **THEN** it serves its routes using each handler's default behavior

### Requirement: Broker configuration fixed for the container's lifetime
A running broker SHALL keep the configuration it was started with. Because a live broker is not restarted while its container runs, changes to kit configuration that affect broker route handlers SHALL take effect for the next container start.

#### Scenario: Config change during a running container
- **WHEN** kit configuration affecting a broker route is edited while the container and its broker are running
- **THEN** the running broker keeps its previous configuration and the new value applies once the container is restarted
