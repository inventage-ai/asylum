## MODIFIED Requirements

### Requirement: Config hash stored on container
The config hash SHALL be stored as a Docker label (`asylum.config.hash`) on the container at creation time. The hash SHALL be computed from a deterministic serialization of runtime-relevant config values: sorted volumes, sorted env key-value pairs, sorted ports, and kit credential settings (sorted by kit name, with explicit ID lists also sorted).

#### Scenario: New container creation
- **WHEN** a new container is started via `docker run`
- **THEN** the container SHALL have an `asylum.config.hash` label with the current config hash

#### Scenario: Deterministic hash
- **WHEN** the same config values are provided in different map iteration orders
- **THEN** the computed hash SHALL be identical

#### Scenario: Credential change detected
- **WHEN** a kit's credential config changes (e.g. from `auto` to an explicit list, or explicit IDs are added)
- **THEN** the config hash SHALL differ from the previously stored label, triggering the stale config warning

#### Scenario: Absent credentials do not contribute
- **WHEN** a kit has no credentials configured (`credentials` is absent or `false`)
- **THEN** that kit SHALL contribute nothing to the config hash
