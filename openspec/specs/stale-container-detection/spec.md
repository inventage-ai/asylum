## ADDED Requirements

### Requirement: Image staleness detection on running container
When a container is already running, asylum SHALL compare the container's image ID against the expected image tag's ID before exec'ing into it.

#### Scenario: Container image matches expected tag
- **WHEN** the running container's image ID matches the image ID of the expected tag from `EnsureProject`
- **THEN** asylum SHALL exec into the container without prompting or restarting

#### Scenario: Container image is stale with no active sessions
- **WHEN** the running container's image ID does not match the expected tag's image ID
- **AND** the container has no active exec sessions (only the sleep process)
- **THEN** asylum SHALL log "config changed, restarting container..." and kill the container, then start a new one with the correct image

#### Scenario: Container image is stale with active sessions
- **WHEN** the running container's image ID does not match the expected tag's image ID
- **AND** the container has active exec sessions
- **THEN** asylum SHALL prompt the user: "Image has changed. Restart container?"
- **AND** if the user confirms, asylum SHALL kill the container and start a new one
- **AND** if the user declines, asylum SHALL exec into the existing (stale) container

### Requirement: Active session detection
Asylum SHALL detect whether a running container has active exec sessions by counting processes via `docker top`.

#### Scenario: Only sleep process running
- **WHEN** the container has exactly one process (the sleep infinity entrypoint)
- **THEN** `HasActiveSessions` SHALL return false

#### Scenario: Additional processes running
- **WHEN** the container has processes beyond the sleep entrypoint (agent sessions, shells, background tasks)
- **THEN** `HasActiveSessions` SHALL return true

#### Scenario: docker top fails
- **WHEN** `docker top` fails (container stopping, Docker daemon issue)
- **THEN** `HasActiveSessions` SHALL return true (assume sessions exist, prefer prompting over silent kill)

### Requirement: Config drift detection
Asylum SHALL detect when non-image runtime config (volumes, env vars, ports) has changed relative to the running container's config.

#### Scenario: Config hash matches
- **WHEN** the running container's `asylum.config.hash` label matches the current config hash
- **THEN** no warning SHALL be displayed

#### Scenario: Config hash differs (image is up to date)
- **WHEN** the running container's image is up to date but the `asylum.config.hash` label does not match the current config hash
- **THEN** asylum SHALL warn: "config changed (volumes/env/ports) -- restart with --rebuild to apply"

#### Scenario: Config hash differs (image is also stale)
- **WHEN** both the image and config hash are stale
- **THEN** the image staleness handling SHALL take precedence (restart handles both)

#### Scenario: Container has no config hash label (legacy container)
- **WHEN** the running container has no `asylum.config.hash` label
- **THEN** asylum SHALL skip the config drift check (no warning)

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

### Requirement: Unconditional image freshness check
`EnsureBase` and `EnsureProject` SHALL be called on every `asylum` invocation, regardless of whether a container is currently running. When images are up to date, these calls SHALL return quickly (hash check only, no build).

#### Scenario: Container running, images up to date
- **WHEN** a container is running and the current config produces the same image hashes
- **THEN** `EnsureBase` and `EnsureProject` SHALL return without building, and the running container SHALL not be disturbed

#### Scenario: Container running, base image changed
- **WHEN** a container is running and `EnsureBase` detects a hash mismatch
- **THEN** `EnsureBase` SHALL rebuild the base image, `EnsureProject` SHALL rebuild the project image (due to `baseRebuilt` flag), and the running container SHALL be detected as stale

#### Scenario: EnsureBase inspect failure with running container
- **WHEN** `docker inspect` fails during `EnsureBase` and a container is running
- **THEN** asylum SHALL treat the images as up to date and exec into the running container rather than erroring out

### Requirement: Docker helper functions
New docker package functions SHALL support container and image introspection.

#### Scenario: ContainerImageID returns image digest
- **WHEN** `ContainerImageID` is called with a running container name
- **THEN** it SHALL return the SHA256 image ID of the image the container was started with

#### Scenario: ImageID returns tag digest
- **WHEN** `ImageID` is called with a valid image tag
- **THEN** it SHALL return the SHA256 image ID that the tag currently points to

#### Scenario: ContainerLabel returns label value
- **WHEN** `ContainerLabel` is called with a container name and label key
- **THEN** it SHALL return the label value, or empty string if the label does not exist
