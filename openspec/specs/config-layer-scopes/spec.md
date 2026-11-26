# config-layer-scopes Specification

## Purpose

Defines which config files accept which options, so a project's committed `.asylum` cannot give its container access outside the project directory. Covers how out-of-scope options are dropped and how the user is told about them at startup.

## Requirements

### Requirement: Option scopes
Every config option SHALL have one of three scopes. Options with scope `any` SHALL be accepted in `~/.asylum/config.yaml`, the local layer (`.asylum.local` / `.asylum.local.yaml`) and the project layer (`.asylum` / `.asylum.yaml`). Options with scope `personal` SHALL be accepted in the global config and the local layer. Options with scope `global` SHALL be accepted in the global config only. CLI flags SHALL NOT be restricted by scopes.

#### Scenario: Any-scope option in the project layer
- **WHEN** `.asylum` contains `kits: {apt: {packages: [golang]}}`
- **THEN** `golang` is installed in the project image

#### Scenario: Personal option in the local layer
- **WHEN** `.asylum.local` contains `volumes: ["~/data:/data"]`
- **THEN** `~/data` is mounted at `/data`

#### Scenario: Personal option in the project layer
- **WHEN** `.asylum` contains `volumes: ["~/data:/data"]`
- **THEN** `~/data` is not mounted

#### Scenario: Global option in the local layer
- **WHEN** `.asylum.local` contains `release-channel: dev` and the global config does not set it
- **THEN** self-update uses the `stable` channel

#### Scenario: CLI flag
- **WHEN** the user runs `asylum -v ~/data:/data` and no config file sets `volumes`
- **THEN** `~/data` is mounted at `/data`

### Requirement: Personal options
The following options SHALL have scope `personal`: `volumes`, `env-from-host`, `default-resume`, `agents.<agent>.config`, `agents.<agent>.companions`, `agents.<agent>.seed`, and for every kit `credentials`, `isolation`, `allow-agent-terminal-title` and `schemes`. All options not listed as `personal` or `global` SHALL have scope `any`.

#### Scenario: SSH isolation in the project layer
- **WHEN** the global config does not set `kits.ssh.isolation` and `.asylum` contains `kits: {ssh: {isolation: shared}}`
- **THEN** the host `~/.ssh` is not mounted and the SSH kit uses `isolated` mode

#### Scenario: Agent config mode in the project layer
- **WHEN** the global config sets `agents: {claude: {config: isolated}}` and `.asylum` sets `agents: {claude: {config: shared}}`
- **THEN** Claude runs with `isolated` config

#### Scenario: Credentials in the project layer
- **WHEN** `.asylum` contains `kits: {java: {credentials: auto}}` and no other layer sets `credentials` for `java`
- **THEN** no Maven credentials are mounted

### Requirement: Global-only options
The options `release-channel` and `version-check-interval` SHALL have scope `global`.

#### Scenario: Version check interval in the project layer
- **WHEN** `.asylum` contains `version-check-interval: 1m`
- **THEN** the interval from the global config, or the 24-hour default, applies

### Requirement: Personal kits
Enabling a kit whose use reaches outside the project directory or adds container privileges SHALL have scope `personal`. This SHALL apply to the `docker`, `dropshare`, `iterm`, `cx` and `rtk` kits. A kit SHALL also be able to declare the scope of its kit-specific options. Kit options a kit does not declare SHALL use the general scope of that option.

#### Scenario: Docker kit enabled in the project layer
- **WHEN** neither the global config nor `.asylum.local` enables `docker` and `.asylum` contains `kits: {docker: {}}`
- **THEN** the docker kit is not active and the container is not privileged

#### Scenario: Docker kit enabled in the local layer
- **WHEN** `.asylum.local` contains `kits: {docker: {}}`
- **THEN** the docker kit is active

#### Scenario: Plain kit enabled in the project layer
- **WHEN** `.asylum` contains `kits: {python: {}}`
- **THEN** the python kit is active

### Requirement: Disabling is accepted in every layer
Setting `disabled: true` on a kit SHALL be accepted in every layer, including for personal kits, because it only narrows the sandbox. A project-layer entry for a personal kit that contains anything other than `disabled: true` SHALL be rejected as a whole.

#### Scenario: Project disables a personal kit
- **WHEN** the global config enables `docker` and `.asylum` contains `kits: {docker: {disabled: true}}`
- **THEN** the docker kit is not active and nothing is reported as rejected

### Requirement: Rejected options never take effect
An option set in a file whose layer its scope does not accept SHALL be removed from that layer before layers are merged. This SHALL apply to every command that loads config, so the option has no effect anywhere. Other options in the same file SHALL still apply.

#### Scenario: Mixed file
- **WHEN** `.asylum` contains `volumes: ["~/data:/data"]` and `env: {DEBUG: "1"}`
- **THEN** `DEBUG=1` is set in the container and `~/data` is not mounted

### Requirement: Startup prompt for rejected options
When a session start finds rejected options and a terminal is attached, asylum SHALL show a blocking prompt before it starts or attaches to the container. The prompt SHALL list each rejected option with its file, line, path and value, and SHALL name the files that accept it. It SHALL offer to start without these options or to abort.

#### Scenario: Prompt lists the rejected option
- **WHEN** a terminal is attached and line 5 of `.asylum` sets `volumes: ["~/.m2/settings.xml:/home/claude/.m2/settings.xml"]`
- **THEN** the prompt shows `.asylum:5`, `volumes`, the volume entry, and names `.asylum.local` and `~/.asylum/config.yaml`

#### Scenario: User starts without the options
- **WHEN** the user chooses to start without the options
- **THEN** the session starts and the rejected options have no effect

#### Scenario: User aborts
- **WHEN** the user chooses to abort
- **THEN** asylum exits with a non-zero status without creating or attaching to a container

### Requirement: Decision remembered per project
When the user chooses to start without the rejected options, asylum SHALL remember the decision for the project, keyed by the set of rejected entries (file, path and value). It SHALL NOT prompt again while that set stays the same, and SHALL prompt again when it changes.

#### Scenario: Unchanged rejections
- **WHEN** the user chose to start without the rejected options and starts asylum again with the same config files
- **THEN** no prompt is shown and a single warning line says that options in project config are ignored

#### Scenario: New rejected entry
- **WHEN** the user chose to start without the rejected options and `.asylum` later adds another volume
- **THEN** the prompt is shown again and lists all rejected entries

### Requirement: Non-interactive starts
When rejected options are found and no terminal is attached, asylum SHALL log a warning for each rejected option with the same details as the prompt and SHALL start without the options. Commands that do not start a session SHALL log the same warnings and SHALL NOT prompt.

#### Scenario: No terminal
- **WHEN** asylum starts a session without a terminal and `.asylum` enables the `docker` kit
- **THEN** a warning names `.asylum`, the line and `kits.docker`, and the session starts without the docker kit

### Requirement: Config files are not rewritten for scopes
Asylum SHALL NOT move, remove or rewrite options in any config file because of their scope.

#### Scenario: File unchanged after starting without the options
- **WHEN** the user starts without the rejected options from `.asylum`
- **THEN** `.asylum` and `.asylum.local` are byte-for-byte unchanged
