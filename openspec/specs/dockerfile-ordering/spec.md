## ADDED Requirements

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

### Requirement: State-tracked source order
The system SHALL persist the Dockerfile source order in `state.json` as a `docker_source_order` field (string array) after every successful base image build.

#### Scenario: First build with no previous order
- **WHEN** `state.json` has no `docker_source_order` field (or it is empty)
- **THEN** all active sources are treated as new and sorted by priority

#### Scenario: Order persisted after successful build
- **WHEN** a base image build completes successfully
- **THEN** `state.json` is updated with the `docker_source_order` reflecting the source order used for that build

#### Scenario: Order not persisted after failed build
- **WHEN** a base image build fails
- **THEN** `state.json` retains the previous `docker_source_order`

#### Scenario: State contains unknown identifiers
- **WHEN** `docker_source_order` contains identifiers not present in the current active sources
- **THEN** those identifiers are ignored during ordering (filtered out)

### Requirement: Retained sources preserve previous order
When computing the Dockerfile order, sources that were present in the previous build SHALL retain their relative order from the previous build.

#### Scenario: No changes to active sources
- **WHEN** the active sources are identical to the previous build's sources
- **THEN** the order is unchanged from the previous build

#### Scenario: Subset of sources unchanged
- **WHEN** sources A, B, C were in the previous order and all three are still active
- **THEN** A, B, C appear in the same relative order as before

### Requirement: New sources appended at end
Sources that are active but were not present in the previous build's source order SHALL be appended after all retained sources, sorted by priority among themselves.

#### Scenario: Single new source added
- **WHEN** the previous order was [A, B, C] and source D (priority 30) is newly active
- **THEN** the resulting order is [A, B, C, D]

#### Scenario: Multiple new sources added
- **WHEN** the previous order was [A, B] and sources C (priority 40) and D (priority 20) are newly added
- **THEN** the resulting order is [A, B, D, C] (new sources sorted by priority, lower first)

### Requirement: Re-sort suffix after source removal
When a previously-present source is removed, all retained sources from the earliest removal point onward SHALL be re-sorted by their static priority.

#### Scenario: Source removed from middle
- **WHEN** the previous order was [A(10), B(30), C(20), D(40)] and B is removed
- **THEN** A retains position 1 (before the removal point), and [C(20), D(40)] are re-sorted by priority, yielding [A, C, D]

#### Scenario: Source removed from beginning
- **WHEN** the previous order was [A(10), B(20), C(30)] and A is removed
- **THEN** all remaining sources are re-sorted by priority: [B(20), C(30)]

#### Scenario: Multiple sources removed
- **WHEN** the previous order was [A(10), B(30), C(20), D(40), E(15)] and B and D are removed
- **THEN** the earliest removal point is B's position (index 1), so [C(20), E(15)] are re-sorted: [A, E, C]

#### Scenario: Source removed and new source added
- **WHEN** the previous order was [A(10), B(30), C(20)] and B is removed while D(25) is added
- **THEN** the prefix [A] is preserved, the suffix [C(20)] is re-sorted, and D(25) is appended: [A, C, D]

### Requirement: Priority-based tie-breaking
When sources have the same priority and are being sorted (either as new sources or in a re-sorted suffix), they SHALL be ordered alphabetically by their source identifier as a stable tie-breaker.

#### Scenario: Same priority sources
- **WHEN** sources `kit:foo` (priority 50) and `kit:bar` (priority 50) need to be sorted
- **THEN** they are ordered as [`kit:bar`, `kit:foo`] (alphabetical)

### Requirement: Dockerfile assembly uses computed order
The `assembleDockerfile` function SHALL use the computed source order (not the kit resolution order) when concatenating kit DockerSnippets, then append the agent snippet block after all kit snippets.

#### Scenario: Order differs from resolution order
- **WHEN** kit resolution returns [java, docker, node] but the computed order is [java, node, docker]
- **THEN** the assembled Dockerfile places java's snippet first, then node's, then docker's

#### Scenario: Core and tail position unchanged
- **WHEN** the Dockerfile is assembled with any source order
- **THEN** `Dockerfile.core` is always first and `Dockerfile.tail` is always last, with ordered kit snippets and then the agent block between them

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
