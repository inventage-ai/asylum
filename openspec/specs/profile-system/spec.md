## ADDED Requirements

### Requirement: Profile struct
A profile SHALL be a Go struct with fields: Name, Description, DockerSnippet, EntrypointSnippet, CacheDirs, Config, OnboardingTasks, and SubProfiles.

#### Scenario: Profile with sub-profiles
- **WHEN** a profile is defined with SubProfiles containing child profiles
- **THEN** each child profile is a complete Profile struct accessible by name

### Requirement: Built-in profile registry
The system SHALL provide a registry of built-in profiles: `java` (with sub-profiles `maven`, `gradle`), `python` (with sub-profile `uv`), and `node` (with sub-profiles `npm`, `pnpm`, `yarn`).

#### Scenario: Registry lookup by name
- **WHEN** a profile is looked up by name (e.g., "java")
- **THEN** the corresponding Profile struct is returned

#### Scenario: Unknown profile name
- **WHEN** a profile is looked up by a name not in the registry
- **THEN** an error is returned

### Requirement: Hierarchical activation
Activating a top-level profile SHALL activate it and all its sub-profiles. Activating a sub-profile via `parent/child` syntax SHALL activate only the parent and that specific child.

#### Scenario: Activate top-level profile
- **WHEN** the profile list contains "java"
- **THEN** the resolved list contains the java profile, maven sub-profile, and gradle sub-profile

#### Scenario: Activate specific sub-profile
- **WHEN** the profile list contains "java/maven"
- **THEN** the resolved list contains the java profile and maven sub-profile, but not gradle

#### Scenario: Multiple sub-profiles equivalent to parent
- **WHEN** the profile list contains "java/maven" and "java/gradle"
- **THEN** the resolved list is equivalent to activating "java"

#### Scenario: Parent appears before children
- **WHEN** profiles are resolved
- **THEN** the returned list has each parent profile before its children, in deterministic order

### Requirement: Default activation
When no `profiles` key is specified in any config layer, the system SHALL default to activating all built-in top-level profiles with all their sub-profiles.

#### Scenario: No profiles key in config
- **WHEN** no config layer specifies `profiles`
- **THEN** all built-in profiles (java, python, node) with all sub-profiles are active

#### Scenario: Explicit empty profiles
- **WHEN** any config layer specifies `profiles: []`
- **THEN** no language profiles are active (core only)

### Requirement: Deduplication
A profile activated through multiple paths SHALL appear only once in the resolved list.

#### Scenario: Profile activated by name and by sub-profile
- **WHEN** the profile list contains both "java" and "java/maven"
- **THEN** the java profile and maven sub-profile each appear exactly once
