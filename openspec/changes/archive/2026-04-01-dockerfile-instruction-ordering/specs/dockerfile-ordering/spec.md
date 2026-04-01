## ADDED Requirements

### Requirement: Static Dockerfile priority
Each kit and agent install SHALL have a `DockerPriority` integer field that represents its preferred position in the Dockerfile. Lower values indicate earlier placement (more stable, more expensive layers).

#### Scenario: Kit with explicit priority
- **WHEN** a kit is registered with `DockerPriority: 10`
- **THEN** the ordering system uses `10` as that kit's priority value

#### Scenario: Kit with default priority
- **WHEN** a kit is registered without setting `DockerPriority`
- **THEN** the ordering system uses `50` as that kit's priority value (Go zero-value is not used; default is applied during ordering)

#### Scenario: Agent install with priority
- **WHEN** an agent install is registered with `DockerPriority: 20`
- **THEN** the ordering system uses `20` as that agent's priority value

### Requirement: Source identifier scheme
The ordering system SHALL identify each Dockerfile snippet source with a string of the form `"kit:<name>"` for kits and `"agent:<name>"` for agent installs.

#### Scenario: Kit identifier
- **WHEN** the active kits include a kit named `java`
- **THEN** its source identifier is `"kit:java"`

#### Scenario: Agent identifier
- **WHEN** the active agents include an agent named `claude`
- **THEN** its source identifier is `"agent:claude"`

#### Scenario: Sub-kit identifier
- **WHEN** the active kits include a sub-kit named `java/maven`
- **THEN** its source identifier is `"kit:java/maven"`

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
The `assembleDockerfile` function SHALL use the computed source order (not the kit resolution order) when concatenating DockerSnippets.

#### Scenario: Order differs from resolution order
- **WHEN** kit resolution returns [java, docker, node] but the computed order is [java, node, docker]
- **THEN** the assembled Dockerfile places java's snippet first, then node's, then docker's

#### Scenario: Core and tail position unchanged
- **WHEN** the Dockerfile is assembled with any source order
- **THEN** `Dockerfile.core` is always first and `Dockerfile.tail` is always last, with ordered kit/agent snippets between them

### Requirement: Project image agents before kits
In project image Dockerfile generation, agent install snippets SHALL be placed before project kit snippets. The full state-tracked ordering algorithm does not apply to project images, but the agents-before-kits rule does.

#### Scenario: Project image with agents and kits
- **WHEN** a project image is generated with agent snippets and project kit snippets
- **THEN** agent snippets appear before kit snippets in the generated Dockerfile

#### Scenario: Project image with only kits
- **WHEN** a project image is generated with project kit snippets but no agent snippets
- **THEN** kit snippets are placed as normal (no change from current behavior)
