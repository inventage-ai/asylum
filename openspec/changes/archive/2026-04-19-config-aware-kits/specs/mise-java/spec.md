## MODIFIED Requirements

### Requirement: Java versions managed by mise
The java kit's `DockerSnippetFunc` SHALL generate the mise install command from the configured `Versions` list (defaulting to `[17, 21, 25]` when no config is provided). The default version SHALL be set from `DefaultVersion` config (defaulting to `21`).

#### Scenario: Default versions (no config)
- **WHEN** the java kit is active with no user version config
- **THEN** the generated DockerSnippet SHALL install Java 17, 21, and 25 with 21 as default

#### Scenario: Custom versions configured
- **WHEN** the java kit has `versions: [21]` and `default-version: 21`
- **THEN** the generated DockerSnippet SHALL install only Java 21

#### Scenario: Multiple custom versions
- **WHEN** the java kit has `versions: [17, 25]` and `default-version: 25`
- **THEN** the generated DockerSnippet SHALL install Java 17 and 25 with 25 as default

#### Scenario: Default version not in versions list
- **WHEN** `default-version` specifies a version not in `versions`
- **THEN** the kit's `ProjectSnippetFunc` SHALL return a snippet that installs the missing version via mise

### Requirement: Java version selection via ASYLUM_JAVA_VERSION
The java kit's `EnvFunc` SHALL return `{"ASYLUM_JAVA_VERSION": defaultVersion}` when a default version is configured. The entrypoint snippet is unchanged.

#### Scenario: Default version configured
- **WHEN** the java kit has `default-version: 25`
- **THEN** `EnvFunc` SHALL return `{"ASYLUM_JAVA_VERSION": "25"}`

#### Scenario: No user config
- **WHEN** the java kit has no user config (`*SnippetConfig` is nil)
- **THEN** `EnvFunc` SHALL return `{"ASYLUM_JAVA_VERSION": "21"}` (the built-in default)

### Requirement: Non-pre-installed Java in project Dockerfile
The java kit's `ProjectSnippetFunc` SHALL return a mise install command when `DefaultVersion` is set but not present in `Versions`. When `DefaultVersion` is in `Versions` (or not set), the func SHALL return an empty string.

#### Scenario: Custom Java version not in base
- **WHEN** `default-version` is `11` and `versions` is `[17, 21, 25]`
- **THEN** `ProjectSnippetFunc` SHALL return `RUN ~/.local/bin/mise install java@11 && mise use --global java@11`

#### Scenario: Default version already in base
- **WHEN** `default-version` is `21` and `versions` includes `21`
- **THEN** `ProjectSnippetFunc` SHALL return an empty string

#### Scenario: No default version
- **WHEN** no `default-version` is configured
- **THEN** `ProjectSnippetFunc` SHALL return an empty string
