## ADDED Requirements

### Requirement: Read Java version from .tool-versions
The config system SHALL read `.tool-versions` from the project directory and use the Java version as `versions.java` when not already set by asylum config or CLI flags.

#### Scenario: .tool-versions provides Java version
- **WHEN** `.tool-versions` contains `java 21.0.2` and no asylum config sets `versions.java`
- **THEN** the loaded config has `versions.java` set to `"21.0.2"`

#### Scenario: Asylum config overrides .tool-versions
- **WHEN** `.tool-versions` contains `java 21.0.2` and `.asylum` sets `versions.java: "17"`
- **THEN** the loaded config has `versions.java` set to `"17"`

#### Scenario: CLI flag overrides .tool-versions
- **WHEN** `.tool-versions` contains `java 21.0.2` and `--java 25` is passed
- **THEN** the loaded config has `versions.java` set to `"25"`

#### Scenario: No .tool-versions file
- **WHEN** no `.tool-versions` file exists in the project directory
- **THEN** config loading is unaffected

#### Scenario: .tool-versions with no Java line
- **WHEN** `.tool-versions` exists but has no `java` entry
- **THEN** `versions.java` is unaffected
