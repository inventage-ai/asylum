## MODIFIED Requirements

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
