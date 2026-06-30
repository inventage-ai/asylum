# agent-container-resolution Specification

## Purpose

Resolve which container serves a given agent in a project. Each container is labelled with the agents baked into its image; when the requested agent is not supported by the project's primary container, asylum spills to a separate, portless secondary container (named for the project plus agent set) instead of failing or rebuilding the primary. This lets an agent be run ad-hoc without disturbing a running session, while properly configured agents stay first-class in the primary container.

## Requirements

### Requirement: Agent support label on containers
New containers SHALL carry a label `asylum.agents` whose value is the sorted, comma-separated set of agents baked into the container's image.

#### Scenario: Single configured agent
- **WHEN** a container is started for a project whose image bakes only `claude`
- **THEN** the container has label `asylum.agents=claude`

#### Scenario: Multiple baked agents
- **WHEN** the image bakes `claude` and `pi`
- **THEN** the container has label `asylum.agents=claude,pi` (sorted)

### Requirement: Requested agent support check
On startup, after locating a project's container, the system SHALL determine whether the requested agent is supported by reading the `asylum.agents` label. The requested agent SHALL be considered supported when it is a member of the label's comma-separated list. A container with no `asylum.agents` label SHALL be treated as supporting only the default agent (`claude`).

#### Scenario: Requested agent present
- **WHEN** the located container has label `asylum.agents=claude,pi` and the requested agent is `pi`
- **THEN** the agent is supported and the container is reused

#### Scenario: Requested agent absent
- **WHEN** the located container has label `asylum.agents=claude` and the requested agent is `pi`
- **THEN** the agent is not supported

#### Scenario: Legacy container without label
- **WHEN** the located container has no `asylum.agents` label and the requested agent is `claude`
- **THEN** the agent is treated as supported and the container is reused

#### Scenario: Legacy container with non-default agent
- **WHEN** the located container has no `asylum.agents` label and the requested agent is `pi`
- **THEN** the agent is not supported and the legacy container is left untouched

### Requirement: Two-pass container resolution
The system SHALL resolve a project's container in two passes. First it SHALL look up the primary container named for the project's configured agent set. If that container is running and supports the requested agent, it SHALL be reused. Otherwise the system SHALL derive a secondary container name from `sha256(project_dir + sorted_agents)` for the requested agent (plus its companions) and repeat the running-container lookup against that name, starting a new container there if none is running. The primary container SHALL NOT be stopped or removed as part of this resolution.

#### Scenario: Reuse primary when supported
- **WHEN** the primary container is running and its label includes the requested agent
- **THEN** the session execs into the primary container and no secondary is created

#### Scenario: Spill to a secondary container
- **WHEN** the primary container is running but does not support the requested agent, and no secondary container exists for `project + agents`
- **THEN** a new secondary container named for `sha256(project_dir + sorted_agents)` is started, and the primary container keeps running

#### Scenario: Reuse an existing secondary
- **WHEN** a secondary container for `project + agents` is already running and supports the requested agent
- **THEN** the session execs into that secondary container

#### Scenario: Primary preserved across spill
- **WHEN** a secondary container is created because the requested agent is unsupported by the primary
- **THEN** the primary container is neither stopped nor removed

### Requirement: Secondary containers receive no ports
Secondary containers SHALL NOT be allocated a host port range and SHALL NOT include port-forwarding arguments. Only the primary container participates in port allocation.

#### Scenario: No port allocation for secondary
- **WHEN** a secondary container is started for an ad-hoc agent
- **THEN** the system does not call port allocation and the docker run args contain no `-p` flags for that container

#### Scenario: Promotion restores ports
- **WHEN** an agent is added to the project or user config and baked into the primary image
- **THEN** that agent is supported by the primary container, no secondary is created, and the agent runs with the primary container's allocated ports
