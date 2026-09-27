# Spec Delta

## MODIFIED Requirements

### Requirement: Unconditional image freshness check
`EnsureBase` and `EnsureProject` SHALL be called on every `asylum` invocation, regardless of whether a container is currently running. When images are up to date, these calls SHALL return quickly (hash check only, no build).

#### Scenario: Container running, images up to date
- **WHEN** a container is running and the current config produces the same image hashes
- **THEN** `EnsureBase` and `EnsureProject` SHALL return without building, and the running container SHALL not be disturbed

#### Scenario: Container running, base image changed
- **WHEN** a container is running and `EnsureBase` detects a hash mismatch
- **THEN** `EnsureBase` SHALL rebuild the base image, `EnsureProject` SHALL rebuild the project image because it was built on a different base, and the running container SHALL be detected as stale

#### Scenario: Base image changed by another project's run
- **WHEN** a project's container is running on a project image, and the base image was rebuilt during another project's invocation
- **THEN** the next invocation for this project SHALL rebuild its project image, and the running container SHALL be detected as stale

#### Scenario: EnsureBase inspect failure with running container
- **WHEN** `docker inspect` fails during `EnsureBase` and a container is running
- **THEN** asylum SHALL treat the images as up to date and exec into the running container rather than erroring out
