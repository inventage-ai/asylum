## MODIFIED Requirements

### Requirement: Static Dockerfile priority
Each kit SHALL have a `DockerPriority` integer field that represents its preferred position in the Dockerfile. Lower values indicate earlier placement (more stable, more expensive layers). Agent installs SHALL NOT have a priority and SHALL NOT participate in priority-based ordering.

#### Scenario: Kit with explicit priority
- **WHEN** a kit is registered with `DockerPriority: 10`
- **THEN** the ordering system uses `10` as that kit's priority value

#### Scenario: Kit with default priority
- **WHEN** a kit is registered without setting `DockerPriority`
- **THEN** the ordering system uses `50` as that kit's priority value (Go zero-value is not used; default is applied during ordering)

#### Scenario: Agents excluded from priority ordering
- **WHEN** the active sources include agent installs
- **THEN** those agents are not assigned a priority and are not included in the priority-ordered, state-tracked source set

### Requirement: Source identifier scheme
The ordering system SHALL identify each state-tracked Dockerfile snippet source with a string of the form `"kit:<name>"`. Agent installs are not state-tracked sources and have no ordering identifier.

#### Scenario: Kit identifier
- **WHEN** the active kits include a kit named `java`
- **THEN** its source identifier is `"kit:java"`

#### Scenario: Sub-kit identifier
- **WHEN** the active kits include a sub-kit named `java/maven`
- **THEN** its source identifier is `"kit:java/maven"`

#### Scenario: Agents have no ordering identifier
- **WHEN** the active agents include an agent named `claude`
- **THEN** no `"agent:claude"` identifier is added to the tracked source set or persisted in `docker_source_order`

### Requirement: Dockerfile assembly uses computed order
The `assembleDockerfile` function SHALL use the computed source order (not the kit resolution order) when concatenating kit DockerSnippets, then append the agent snippet block after all kit snippets.

#### Scenario: Order differs from resolution order
- **WHEN** kit resolution returns [java, docker, node] but the computed order is [java, node, docker]
- **THEN** the assembled Dockerfile places java's snippet first, then node's, then docker's

#### Scenario: Core and tail position unchanged
- **WHEN** the Dockerfile is assembled with any source order
- **THEN** `Dockerfile.core` is always first and `Dockerfile.tail` is always last, with ordered kit snippets and then the agent block between them

## ADDED Requirements

### Requirement: Agent snippets appended after kits
Agent install DockerSnippets SHALL be emitted as a contiguous block placed after all ordered kit snippets and before `Dockerfile.tail`, so that agent version changes invalidate only the agent layers and the tail, never any kit layer.

#### Scenario: Agents follow kits in the base image
- **WHEN** the base image is assembled with both kit snippets and agent snippets
- **THEN** every kit snippet appears before every agent snippet, and every agent snippet appears before the tail

#### Scenario: Agent version bump preserves kit layers
- **WHEN** only an agent's pinned version changes between builds
- **THEN** the ordered kit snippets are byte-identical and unchanged in position, so their Docker layers remain cached

#### Scenario: Deterministic agent block order
- **WHEN** the agent block is assembled for a fixed set of active agents
- **THEN** the agents are emitted in a deterministic order (claude first, remaining agents by name)

## REMOVED Requirements

### Requirement: Project image agents before kits
**Reason**: Project images never install agents — `generateProjectDockerfile` takes no agent installs (agents are inherited from the base image). The requirement described behavior that does not exist in the code.
**Migration**: None. No project-image behavior changes; agents continue to be installed only in the base image, now after all base kits.
