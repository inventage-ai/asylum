## ADDED Requirements

### Requirement: Entrypoint decomposition
The monolithic entrypoint.sh SHALL be split into an embedded core fragment (PATH setup, git config, SSH, direnv, Docker daemon) and an embedded tail fragment (welcome banner, exec).

#### Scenario: Core fragment content
- **WHEN** the core entrypoint is examined
- **THEN** it contains base PATH exports, host gitconfig handling, SSH setup, direnv approval translation, and optional Docker daemon startup

#### Scenario: Tail fragment content
- **WHEN** the tail entrypoint is examined
- **THEN** it contains the welcome banner display and the final `exec "$@"`

### Requirement: Entrypoint assembly from active profiles
The entrypoint SHALL be assembled at image build time by concatenating: core fragment + EntrypointSnippets from active profiles (in deterministic order) + tail fragment.

#### Scenario: Node profile entrypoint snippet
- **WHEN** the node profile is active
- **THEN** its EntrypointSnippet (fnm env setup, PATH additions) is included in the assembled entrypoint

#### Scenario: Python/uv profile entrypoint snippet
- **WHEN** the python/uv profile is active
- **THEN** its EntrypointSnippet (venv auto-creation for Python projects) is included in the assembled entrypoint

#### Scenario: Java profile entrypoint snippet
- **WHEN** the java profile is active
- **THEN** its EntrypointSnippet (mise activation, ASYLUM_JAVA_VERSION handling) is included in the assembled entrypoint

### Requirement: Welcome banner reflects active profiles
The welcome banner in the entrypoint tail SHALL only display version information for tools installed by active profiles.

#### Scenario: Only java profile active
- **WHEN** the base image was built with only the java profile
- **THEN** the welcome banner shows Java version but not Python or Node.js versions

#### Scenario: All profiles active
- **WHEN** all profiles are active (default)
- **THEN** the welcome banner shows Java, Python, and Node.js versions (same as current behavior)
