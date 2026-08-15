# config-migration Specification

## Purpose

Upgrades v1 config files to the v2 schema on load, so an existing install keeps working after asylum's config format changes rather than failing to parse. Global and project files are detected by different signals, and a `.backup` copy is written before any file is rewritten.

## Requirements

### Requirement: Global config migration
When loading `~/.asylum/config.yaml`, the system SHALL detect v1 format (missing `version` field or version < 0.2) and migrate it to v2 format in place.

#### Scenario: v1 global config with profiles and features
- **WHEN** global config has `profiles: [java, node]` and `features: {shadow-node-modules: true}`
- **THEN** after migration it has `version: 0.2`, `kits: {java: {}, node: {shadow-node-modules: true}}`, and the old keys are removed

#### Scenario: v1 global config with packages
- **WHEN** global config has `packages: {apt: [ffmpeg], npm: [turbo], pip: [ansible], run: ["curl ..."]}`
- **THEN** after migration `kits.apt.packages: [ffmpeg]`, `kits.node.packages: [turbo]`, `kits.python.packages: [ansible]`, `kits.shell.build: ["curl ..."]`

#### Scenario: v1 global config with agents list
- **WHEN** global config has `agents: [claude, gemini]`
- **THEN** after migration `agents: {claude: {}, gemini: {}}`

#### Scenario: Already v2 config
- **WHEN** global config has `version: 0.2`
- **THEN** no migration is performed

### Requirement: Project config migration
When loading `.asylum` or `.asylum.local`, the system SHALL detect v1 format by the presence of the `features` key and migrate to v2 format in place.

#### Scenario: Project config with features key
- **WHEN** `.asylum` contains `features: {onboarding: false}` and `packages: {apt: [jq]}`
- **THEN** after migration it has `kits: {node: {onboarding: false}, apt: {packages: [jq]}}` and the old keys are removed

#### Scenario: Project config without features key
- **WHEN** `.asylum` does not contain a `features` key
- **THEN** no migration is performed (assumed to be v2 or compatible)

### Requirement: Backup before migration
Before rewriting a config file during migration, the system SHALL create a backup copy with a `.backup` suffix.

#### Scenario: Backup created
- **WHEN** `~/.asylum/config.yaml` is migrated
- **THEN** `~/.asylum/config.yaml.backup` contains the original content

### Requirement: Migration field mapping
The migration SHALL map v1 fields to v2 structure as specified in the design's migration mapping table.

#### Scenario: tab-title migrated
- **WHEN** v1 config has `tab-title: "🤖 {project}"`
- **THEN** v2 config has `kits: {title: {tab-title: "🤖 {project}"}}`

#### Scenario: versions migrated
- **WHEN** v1 config has `versions: {java: "17"}`
- **THEN** v2 config has `kits: {java: {default-version: "17"}}`

#### Scenario: onboarding migrated
- **WHEN** v1 config has `onboarding: {npm: false}`
- **THEN** v2 config has `kits: {node: {onboarding: false}}`
