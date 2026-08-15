# kit-config-sync Specification

## Purpose

Writes newly detected kits into an existing config file without disturbing what the user wrote. A kit the user accepted is inserted live; an opt-in kit, or any new kit found in a non-interactive session, is inserted commented out so it is discoverable but inert. Kits supply their own structured YAML nodes for this, and a config with no `kits` mapping gets one created.

## Requirements

### Requirement: Insert active kit into existing config
When a new kit with tier `TierDefault` is activated (by user consent or first-run), the system SHALL insert it into the `kits` mapping of `~/.asylum/config.yaml` using `yaml.Node` tree manipulation, preserving existing comments and key ordering.

#### Scenario: New default kit added to config
- **WHEN** kit `rust` (tier `TierDefault`) is activated and `config.yaml` has a `kits` mapping without `rust`
- **THEN** a `rust` key-value pair is appended to the `kits` mapping node using the kit's `ConfigNodes` output

#### Scenario: Existing config preserved
- **WHEN** a new kit is inserted into `config.yaml` that has user comments and custom ordering
- **THEN** all existing comments, key order, and values are preserved in the output

#### Scenario: Kit already in config
- **WHEN** kit `docker` is detected as new (not in `known_kits`) but `config.yaml` already has a `docker` key in `kits`
- **THEN** no modification is made to the config for that kit

### Requirement: Insert commented kit into existing config
When a new kit with tier `TierOptIn` is detected, the system SHALL insert it as a YAML comment in the `kits` section so the user can see and enable it.

#### Scenario: Opt-in kit added as comment
- **WHEN** kit `apt` (tier `TierOptIn`) is new and `config.yaml` has a `kits` mapping
- **THEN** a commented-out block for `apt` is added after the existing kit entries

### Requirement: Non-interactive default-on kits added as comments
When the session is non-interactive and a new `TierDefault` kit is detected, the system SHALL insert it as a commented-out entry instead of an active entry.

#### Scenario: Non-interactive adds commented
- **WHEN** a new `TierDefault` kit is detected and stdin is not a terminal
- **THEN** the kit is added to the config as a commented-out entry (same as `TierOptIn`)

### Requirement: Config without kits mapping
When `config.yaml` exists but has no `kits` key, the system SHALL create the mapping before inserting kit entries.

#### Scenario: Add kits mapping to minimal config
- **WHEN** `config.yaml` contains only `agent: claude` with no `kits` key
- **THEN** a `kits` mapping is created and new kit entries are inserted into it

### Requirement: Kit provides structured config nodes
Each kit SHALL provide a `ConfigNodes` method that returns `yaml.Node` key-value pairs for insertion into the kits mapping. This replaces text-based `ConfigSnippet` for config modification purposes.

#### Scenario: Simple kit with no options
- **WHEN** kit `docker` provides its config nodes
- **THEN** it returns a scalar key node (`docker`) with a line comment and an empty mapping value node

#### Scenario: Kit with nested options
- **WHEN** kit `java` provides its config nodes
- **THEN** it returns a scalar key node (`java`) and a mapping value node containing `versions` and `default-version` entries
