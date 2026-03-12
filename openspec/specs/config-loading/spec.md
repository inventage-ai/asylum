## ADDED Requirements

### Requirement: Three-layer config loading
The config system SHALL load config from `~/.asylum/config.yaml`, `$project/.asylum`, and `$project/.asylum.local` in order, merging each layer on top of the previous.

#### Scenario: All three files present
- **WHEN** all three config files exist with different values
- **THEN** values are merged according to merge semantics (scalars last-wins, lists concat, maps-of-lists concat)

#### Scenario: Missing files are skipped
- **WHEN** one or more config files do not exist
- **THEN** loading succeeds with values from the files that do exist

#### Scenario: Invalid YAML
- **WHEN** a config file contains invalid YAML
- **THEN** an error is returned

### Requirement: Scalar merge semantics
Scalar config values (agent, versions.java) SHALL use last-value-wins when merging layers.

#### Scenario: Agent override
- **WHEN** global config sets `agent: claude` and project config sets `agent: gemini`
- **THEN** the merged result has `agent: gemini`

### Requirement: List merge semantics
List config values (ports, volumes) SHALL be concatenated across layers.

#### Scenario: Ports concatenation
- **WHEN** global config has `ports: ["3000"]` and project config has `ports: ["8080"]`
- **THEN** the merged result has `ports: ["3000", "8080"]`

### Requirement: Map-of-lists merge semantics
Map-of-lists values (packages) SHALL have each sub-list concatenated independently across layers.

#### Scenario: Packages sub-list concatenation
- **WHEN** global config has `packages.apt: [curl]` and project config has `packages.apt: [jq]`
- **THEN** the merged result has `packages.apt: [curl, jq]`

### Requirement: CLI flag overlay
CLI scalar flags SHALL override all config layers. CLI list flags SHALL be appended to merged config values.

#### Scenario: Agent flag overrides config
- **WHEN** config sets `agent: claude` and CLI flag sets `-a codex`
- **THEN** the final agent is `codex`

#### Scenario: Port flag appends to config
- **WHEN** config has `ports: ["3000"]` and CLI flag adds `-p 9090`
- **THEN** the final ports are `["3000", "9090"]`
