## ADDED Requirements

### Requirement: Three-layer config loading
The config system SHALL load config from `~/.asylum/config.yaml`, the project layer, and the local layer in order, merging each layer on top of the previous. Before loading, each file SHALL be migrated from v1 format if necessary.

For the project layer, the loader SHALL accept either `$project/.asylum` (canonical) or `$project/.asylum.yaml` (alias). For the local layer, the loader SHALL accept either `$project/.asylum.local` (canonical) or `$project/.asylum.local.yaml` (alias). The two filenames within a layer are equivalent — behavior, migration, and merging are identical regardless of which is found.

If both the canonical filename and its `.yaml` alias exist for the same layer, the loader SHALL return an error rather than choose one or merge them.

#### Scenario: All three files present (canonical names)
- **WHEN** `~/.asylum/config.yaml`, `$project/.asylum`, and `$project/.asylum.local` all exist with different values
- **THEN** values are merged according to merge semantics (scalars last-wins, lists concat, maps merge per-key with field-level merge within KitConfig)

#### Scenario: Project layer uses .yaml alias
- **WHEN** `$project/.asylum.yaml` exists and `$project/.asylum` does not
- **THEN** `.asylum.yaml` is loaded as the project layer with identical behavior to `.asylum`

#### Scenario: Local layer uses .yaml alias
- **WHEN** `$project/.asylum.local.yaml` exists and `$project/.asylum.local` does not
- **THEN** `.asylum.local.yaml` is loaded as the local layer with identical behavior to `.asylum.local`

#### Scenario: Mixed canonical and alias across layers
- **WHEN** `$project/.asylum` exists for the project layer AND `$project/.asylum.local.yaml` exists for the local layer
- **THEN** both are loaded; layer behavior is independent

#### Scenario: Both canonical and alias present in the same layer
- **WHEN** both `$project/.asylum` and `$project/.asylum.yaml` exist
- **THEN** loading SHALL return an error identifying the conflict and instructing the user to remove one

#### Scenario: Missing files are skipped
- **WHEN** one or more config files do not exist
- **THEN** loading succeeds with values from the files that do exist

#### Scenario: Invalid YAML
- **WHEN** a config file contains invalid YAML
- **THEN** an error is returned

#### Scenario: Project kits supplement global kits
- **WHEN** global config has `kits: {node: {}, openspec: {}}` and project config has `kits: {shell: {}}`
- **THEN** the merged result has all three kits active

### Requirement: Scalar merge semantics
Scalar config values (agent, release-channel) SHALL use last-value-wins when merging layers.

#### Scenario: Agent override
- **WHEN** global config sets `agent: claude` and project config sets `agent: gemini`
- **THEN** the merged result has `agent: gemini`

### Requirement: List merge semantics
List config values (ports, volumes) SHALL be concatenated across layers.

#### Scenario: Ports concatenation
- **WHEN** global config has `ports: ["3000"]` and project config has `ports: ["8080"]`
- **THEN** the merged result has `ports: ["3000", "8080"]`

### Requirement: CLI flag overlay
CLI scalar flags SHALL override all config layers. CLI list flags SHALL be appended to merged config values.

#### Scenario: Agent flag overrides config
- **WHEN** config sets `agent: claude` and CLI flag sets `-a codex`
- **THEN** the final agent is `codex`

#### Scenario: Kits flag overrides config
- **WHEN** config has `kits: {java: {}, python: {}}` and CLI passes `--kits java`
- **THEN** the final kits map contains only java

### Requirement: Release channel config field
The config system SHALL support an optional `release-channel` scalar field with values `stable` or `dev`. It follows scalar merge semantics (last value wins across layers).

#### Scenario: Release channel set in global config
- **WHEN** `~/.asylum/config.yaml` contains `release-channel: dev`
- **THEN** the loaded config has `ReleaseChannel` set to `"dev"`

#### Scenario: Not set defaults to empty
- **WHEN** no config file sets `release-channel`
- **THEN** the loaded config has `ReleaseChannel` set to `""` (callers treat empty as stable)

### Requirement: Read Java version from .tool-versions
The config system SHALL read `.tool-versions` from the project directory and use the Java version as `kits.java.default-version` when not already set by asylum config or CLI flags.

#### Scenario: .tool-versions provides Java version
- **WHEN** `.tool-versions` contains `java 21.0.2` and no asylum config sets java's default-version
- **THEN** the loaded config has `kits.java.default-version` set to `"21.0.2"`

#### Scenario: Asylum config overrides .tool-versions
- **WHEN** `.tool-versions` contains `java 21.0.2` and config sets `kits: {java: {default-version: "17"}}`
- **THEN** the loaded config has java default-version set to `"17"`

### Requirement: default-resume config key
The config system SHALL accept an optional top-level boolean key `default-resume`. When set to `true`, asylum's default agent invocation auto-resumes the prior session (the pre-default-new-session behaviour). When unset or `false`, the default is to start a new session. The key SHALL participate in the standard three-layer merge (global → project → local → CLI), last-wins semantics for scalars.

#### Scenario: Unset defaults to false
- **WHEN** no config layer sets `default-resume`
- **THEN** the resolved value is `false`

#### Scenario: Global config opts in
- **WHEN** `~/.asylum/config.yaml` contains `default-resume: true` and no other layer sets it
- **THEN** the resolved value is `true`

#### Scenario: Project layer overrides global
- **WHEN** global config sets `default-resume: true` and `$project/.asylum` sets `default-resume: false`
- **THEN** the resolved value is `false`

#### Scenario: Local layer overrides project
- **WHEN** project config sets `default-resume: false` and `$project/.asylum.local` sets `default-resume: true`
- **THEN** the resolved value is `true`

#### Scenario: Migration dialog writes to global layer
- **WHEN** the resume-migration dialog writes `default-resume: true` on the user's behalf
- **THEN** it is written to `~/.asylum/config.yaml` (the global layer), preserving any other keys already present in that file
