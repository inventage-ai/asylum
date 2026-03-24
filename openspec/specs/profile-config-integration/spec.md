## ADDED Requirements

### Requirement: Profiles field in config
The Config struct SHALL include a `Profiles` field that accepts a list of profile identifiers (e.g., `["java", "python/uv", "node"]`).

#### Scenario: Profiles in YAML config
- **WHEN** a config file contains `profiles: [java, python/uv]`
- **THEN** the parsed Config has Profiles set to `["java", "python/uv"]`

#### Scenario: No profiles key
- **WHEN** no config file specifies `profiles`
- **THEN** the parsed Config has Profiles as nil (interpreted as "all" at resolution time)

#### Scenario: Empty profiles list
- **WHEN** a config file contains `profiles: []`
- **THEN** the parsed Config has Profiles as an empty non-nil slice

### Requirement: Profile config defaults in merge chain
Profile config defaults SHALL be injected into the merge chain after the config layer that activated them but before subsequent layers.

#### Scenario: Global profile defaults overridden by project config
- **WHEN** global config activates the java profile (which defaults to `versions.java: 21`) and project config sets `versions.java: 17`
- **THEN** the merged config has `versions.java: 17`

#### Scenario: Project profile defaults overridden by local config
- **WHEN** project config activates a profile with `env.FOO: bar` and local config sets `env.FOO: baz`
- **THEN** the merged config has `env.FOO: baz`

### Requirement: Profiles last-wins across config layers
The `profiles` field SHALL follow last-wins semantics: if a later config layer specifies `profiles`, it replaces the value from earlier layers entirely (no merging of profile lists).

#### Scenario: Project profiles replace global profiles
- **WHEN** global config has `profiles: [java, python, node]` and project config has `profiles: [java]`
- **THEN** the effective profile list is `[java]` (not a union)

#### Scenario: Local config overrides project profiles
- **WHEN** project config has `profiles: [java, python]` and local config has `profiles: [node]`
- **THEN** the effective profile list is `[node]`

### Requirement: CLI flag for profiles
A `--profiles` CLI flag SHALL allow overriding the profile list from the command line, following the same last-wins semantics.

#### Scenario: CLI overrides all config layers
- **WHEN** config has `profiles: [java]` and CLI passes `--profiles python,node`
- **THEN** the effective profile list is `[python, node]`

### Requirement: Two-tier profile resolution
Profiles activated in global config SHALL be resolved for the base image build. Profiles activated in project config (that are not already in the global set) SHALL be resolved for the project image build.

#### Scenario: Profile in global only
- **WHEN** global config has `profiles: [java]` and project config has no profiles key
- **THEN** java is built into the base image; no additional profiles in the project image

#### Scenario: Profile in project but not global
- **WHEN** global config has `profiles: [java]` and project config has `profiles: [python]`
- **THEN** java is in the base image; python is built into the project image

#### Scenario: Same profile in both tiers
- **WHEN** global config has `profiles: [java]` and project config also has `profiles: [java]`
- **THEN** java is only in the base image; the project image adds nothing for java
