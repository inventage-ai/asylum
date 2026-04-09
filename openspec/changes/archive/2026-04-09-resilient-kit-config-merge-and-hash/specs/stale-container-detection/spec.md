## MODIFIED Requirements

### Requirement: Config hash stored on container
The config hash SHALL be stored as a Docker label (`asylum.config.hash`) on the container at creation time. The hash SHALL be computed by YAML-serializing the full `Config` struct (after zeroing non-runtime fields: `Version`, `Agent`, `ReleaseChannel`, `Agents`) and taking its SHA256 digest. Order-insensitive lists (`Volumes`, `Ports`) SHALL be sorted before serialization. New config fields SHALL be included in the hash automatically without code changes.

#### Scenario: New container creation
- **WHEN** a new container is started via `docker run`
- **THEN** the container SHALL have an `asylum.config.hash` label with the current config hash

#### Scenario: Deterministic hash
- **WHEN** the same config values are provided in different map iteration orders
- **THEN** the computed hash SHALL be identical

#### Scenario: Credential change detected
- **WHEN** a kit's credential config changes (e.g. from `auto` to an explicit list, or explicit IDs are added)
- **THEN** the config hash SHALL differ from the previously stored label, triggering the stale config warning

#### Scenario: New field automatically included
- **WHEN** a new field is added to `Config` or `KitConfig`
- **AND** that field is set to a non-zero value
- **THEN** the config hash SHALL differ from a hash where that field is absent, without any changes to the hash function

#### Scenario: Non-runtime fields excluded
- **WHEN** only `Version`, `Agent`, `ReleaseChannel`, or `Agents` config changes
- **THEN** the config hash SHALL NOT change

#### Scenario: Volume/port order independent
- **WHEN** the same volumes or ports are listed in a different order
- **THEN** the computed hash SHALL be identical
