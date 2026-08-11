# update-command Specification

## Purpose
TBD - created by archiving change on-demand-update. Update Purpose after archive.
## Requirements
### Requirement: On-demand update command

Asylum SHALL provide an `asylum update` subcommand that refreshes agent versions on demand and rebuilds images when needed, without starting or exec'ing into a container. The command SHALL force a fetch of all agent versions regardless of the staleness of `versions.json`, write the successfully fetched versions, and then ensure the base and project images are up to date. When no agent version changed, the image ensure step SHALL be a no-op (cached images reused); when a version changed, the base image (and dependent project image) SHALL be rebuilt.

#### Scenario: Force fetch bypasses staleness

- **WHEN** `asylum update` is run and `versions.json` was updated less than the configured interval ago
- **THEN** all agent versions are fetched anyway (the staleness check is not consulted) and the file is rewritten with the fetched versions

#### Scenario: New versions trigger a rebuild

- **WHEN** `asylum update` fetches versions that differ from the ones baked into the current base image
- **THEN** the assembled Dockerfile hash changes and the base image (and its dependent project image) is rebuilt

#### Scenario: Unchanged versions skip the rebuild

- **WHEN** `asylum update` fetches versions identical to those already baked into the current images
- **THEN** the image hashes match, the cached images are reused, and no rebuild occurs

#### Scenario: Command does not start a container

- **WHEN** `asylum update` completes
- **THEN** asylum exits after the image ensure step without starting a container or exec'ing an agent

#### Scenario: All fetches fail

- **WHEN** `asylum update` is run and every version fetch fails
- **THEN** the existing `versions.json` is left unchanged and the command exits without rebuilding

### Requirement: Update command is distinct from self-update

The `asylum update` command SHALL update agent CLI versions and container images only. It SHALL NOT update the asylum binary itself; binary updates remain the responsibility of `asylum self-update`.

#### Scenario: Binary is not modified

- **WHEN** `asylum update` is run
- **THEN** the asylum executable on disk is not replaced and no release download for the asylum binary is attempted

