# Spec Delta

## Purpose

Lets users pass host environment variables, typically secrets, into the container under a different name without writing their values to any config file.

## ADDED Requirements

### Requirement: Map host variables into the container
The config key `env-from-host` SHALL map a container variable name to a host variable name. When the container is created, the system SHALL set each container variable to the current value of its host variable. The value SHALL NOT be written to any file by asylum.

#### Scenario: Host variable is set
- **WHEN** the config contains `env-from-host: {GH_TOKEN: ASYLUM_GH_TOKEN}` and the host has `ASYLUM_GH_TOKEN=abc`
- **THEN** the container is created with `GH_TOKEN=abc`

#### Scenario: Host variable is unset or empty
- **WHEN** the config contains `env-from-host: {GH_TOKEN: ASYLUM_GH_TOKEN}` and `ASYLUM_GH_TOKEN` is unset or empty on the host
- **THEN** `GH_TOKEN` is not set in the container and container creation succeeds

### Requirement: Only user-maintained configs may map host variables
`env-from-host` SHALL be honored in `~/.asylum/config.yaml` and in the local layer (`.asylum.local` / `.asylum.local.yaml`). In the project layer (`.asylum` / `.asylum.yaml`) it SHALL be ignored and a warning SHALL be logged.

#### Scenario: Mapping in the local layer
- **WHEN** `.asylum.local` contains `env-from-host: {GH_TOKEN: ASYLUM_GH_TOKEN}` and the host variable is set
- **THEN** `GH_TOKEN` is set in the container

#### Scenario: Mapping in the project layer
- **WHEN** `.asylum` contains `env-from-host: {AWS_SECRET_ACCESS_KEY: AWS_SECRET_ACCESS_KEY}`
- **THEN** the mapping is ignored, the variable is not set in the container, and a warning names the file and points to `.asylum.local` or `~/.asylum/config.yaml`

### Requirement: Layer merge and precedence
`env-from-host` maps SHALL be merged per key across the global and local layers, last value wins. If a container variable name is also set in `env` (including `-e` on the command line), the `env` value SHALL win.

#### Scenario: Local layer overrides global mapping
- **WHEN** the global config maps `GH_TOKEN: GH_TOKEN_WORK` and `.asylum.local` maps `GH_TOKEN: GH_TOKEN_PRIVATE`
- **THEN** the container's `GH_TOKEN` is read from `GH_TOKEN_PRIVATE`

#### Scenario: Literal env wins
- **WHEN** the config maps `GH_TOKEN: ASYLUM_GH_TOKEN` and the user runs `asylum -e GH_TOKEN=other`
- **THEN** the container's `GH_TOKEN` is `other`

### Requirement: Same restrictions as literal env vars
Variables from `env-from-host` SHALL be subject to the same restrictions as `env`: the reserved active-agent variable SHALL be skipped with a warning, and a value containing a newline SHALL abort container creation with an error naming the variable.

#### Scenario: Value with newline
- **WHEN** the mapped host variable contains a newline
- **THEN** container creation fails with an error naming the container variable, without printing the value
