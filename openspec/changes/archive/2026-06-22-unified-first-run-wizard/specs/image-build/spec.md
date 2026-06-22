## ADDED Requirements

### Requirement: Pre-build context line
When any actual base or project image build is going to occur during a single asylum invocation, the system SHALL emit exactly one user-facing context line before the per-image `building...` log lines: `Building sandbox image — this takes a few minutes the first time, subsequent runs reuse the cache.` The line SHALL be suppressed entirely when both `EnsureBase` and `EnsureProject` are cache hits in the same invocation.

#### Scenario: First-run build
- **WHEN** neither the base nor the project image exists and asylum is invoked
- **THEN** the context line SHALL be printed exactly once before `building base image...` and SHALL NOT be repeated before `building project image...`

#### Scenario: Base image rebuild only
- **WHEN** the base image hash differs but the project image is up to date (or trivially derived)
- **THEN** the context line SHALL be printed exactly once before `building base image...`

#### Scenario: Project image rebuild only
- **WHEN** the base image is up to date but the project image needs rebuilding
- **THEN** the context line SHALL be printed exactly once before `building project image...`

#### Scenario: Both images cached
- **WHEN** both `EnsureBase` and `EnsureProject` are cache hits
- **THEN** no context line SHALL be printed

#### Scenario: Running container short-circuit
- **WHEN** a container is already running and image-check errors trigger the "using running container" fallthrough
- **THEN** no context line SHALL be printed (no actual build occurs)
