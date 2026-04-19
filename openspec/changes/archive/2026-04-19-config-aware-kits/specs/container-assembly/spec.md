## ADDED Requirements

### Requirement: Kit-contributed environment variables in container
The container assembly SHALL collect environment variables from all active kits that provide an `EnvFunc`. These SHALL be represented as RunArgs with source `kit` and priority 1, and SHALL NOT be hardcoded per-kit in the container assembly code.

#### Scenario: Java kit contributes ASYLUM_JAVA_VERSION
- **WHEN** the java kit is active with `default-version: 21`
- **THEN** the container run args SHALL include `-e ASYLUM_JAVA_VERSION=21` with source `kit`

#### Scenario: Kit returns no env vars
- **WHEN** a kit's `EnvFunc` returns an empty map
- **THEN** no env args SHALL be added for that kit

#### Scenario: No hardcoded kit env vars
- **WHEN** the container is assembled
- **THEN** the container assembly code SHALL NOT contain any kit-specific env var logic (e.g., no `if java` checks)

## REMOVED Requirements

### Requirement: Agent-specific mounts and env vars
**Reason**: The java-specific `ASYLUM_JAVA_VERSION` env var portion of this requirement is replaced by the generic kit `EnvFunc` mechanism. Agent config mounts and agent-specific env vars remain unchanged — only the kit env var portion is removed.
**Migration**: Java env var is now contributed by the java kit's `EnvFunc` instead of hardcoded in container assembly.
