## ADDED Requirements

### Requirement: Dynamic cache directories
Cache directory volume mounts SHALL be aggregated from active profiles' CacheDirs fields instead of a hardcoded constant.

#### Scenario: Java profile cache dirs
- **WHEN** the java/maven profile is active
- **THEN** the `.m2` cache directory is mounted as a named volume

#### Scenario: No profile with caches active
- **WHEN** only profiles without CacheDirs are active
- **THEN** no cache directory volumes are mounted

#### Scenario: Multiple profiles with caches
- **WHEN** java/maven, java/gradle, node/npm, and python profiles are active
- **THEN** cache volumes for `.m2`, `.gradle`, `.npm`, and `.cache/pip` are all mounted

### Requirement: Profile-contributed environment variables
Environment variables from active profiles' Config.Env fields SHALL be included in the container's environment.

#### Scenario: Profile sets env var
- **WHEN** a profile's Config includes `env: { JAVA_TOOL_OPTIONS: "-Dfile.encoding=UTF-8" }`
- **THEN** that environment variable is set in the container

#### Scenario: User config overrides profile env var
- **WHEN** a profile sets `FOO=bar` and user config sets `FOO=baz`
- **THEN** the container has `FOO=baz` (user config wins)

### Requirement: Cache migration uses dynamic dirs
The cache migration logic (from old bind-mounted caches to named volumes) SHALL use the dynamic cache directory map from active profiles.

#### Scenario: Migration only for active profile caches
- **WHEN** old bind-mounted cache directories exist for maven and gradle, but only java/maven profile is active
- **THEN** only the maven cache is migrated
