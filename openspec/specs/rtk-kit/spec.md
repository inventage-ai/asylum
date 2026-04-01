## ADDED Requirements

### Requirement: RTK kit registration
The system SHALL register an `rtk` kit with tier `TierOptIn`, no dependencies, tool name `rtk`, and `NeedsMount: true`.

#### Scenario: Kit appears in registry
- **WHEN** the kit package is initialized
- **THEN** a kit named `rtk` SHALL be registered and retrievable via `kit.Get("rtk")`

#### Scenario: Kit is opt-in
- **WHEN** a user's config does not explicitly include `rtk` in their kits
- **THEN** the RTK kit SHALL NOT be activated

#### Scenario: Kit has no dependencies
- **WHEN** the RTK kit is activated
- **THEN** no other kits SHALL be auto-activated as dependencies

### Requirement: RTK installation and hook generation via Dockerfile snippet
The kit SHALL provide a `DockerSnippet` that installs the RTK binary using RTK's official install script and runs `rtk init -g` at build time to generate hook artifacts, saving them to `/tmp/asylum-kit-rtk/`.

#### Scenario: RTK binary available after image build
- **WHEN** the Docker image is built with the RTK kit active
- **THEN** the `rtk` command SHALL be available on the PATH inside the container

#### Scenario: Hook artifacts generated at build time
- **WHEN** the Docker image is built with the RTK kit active
- **THEN** the RTK hook script and RTK.md awareness doc SHALL be saved under `/tmp/asylum-kit-rtk/`

### Requirement: RTK hook mounting via entrypoint snippet
The kit SHALL provide an `EntrypointSnippet` that mounts the build-time-generated RTK hooks and RTK.md into the Claude Code config directory, and registers the PreToolUse hook in settings.json.

#### Scenario: Hook script mounted into Claude config
- **WHEN** the container starts with the RTK kit active and `~/.claude` exists
- **THEN** the RTK hook script SHALL be mounted into `~/.claude/hooks/`

#### Scenario: RTK.md mounted into Claude config
- **WHEN** the container starts with the RTK kit active and `~/.claude` exists
- **THEN** `RTK.md` SHALL be mounted into `~/.claude/`

#### Scenario: Hook registered in settings.json
- **WHEN** the container starts with the RTK kit active
- **THEN** the RTK PreToolUse hook SHALL be registered in `~/.claude/settings.json`

#### Scenario: No action when Claude config absent
- **WHEN** the container starts with the RTK kit active but `~/.claude` does not exist
- **THEN** no mounting or patching SHALL occur

### Requirement: RTK rules snippet
The kit SHALL provide a `RulesSnippet` documenting RTK's purpose and key commands (`rtk gain`, `rtk discover`) so agents are aware of the tool.

#### Scenario: Rules included when kit active
- **WHEN** the RTK kit is active
- **THEN** the sandbox rules SHALL include a section describing RTK and its commands

### Requirement: RTK banner line
The kit SHALL provide a `BannerLines` entry that displays the RTK version in the container welcome banner.

#### Scenario: Version shown in banner
- **WHEN** a user starts an interactive session with the RTK kit active
- **THEN** the welcome banner SHALL include a line showing the RTK version

### Requirement: RTK config snippet
The kit SHALL provide a `ConfigSnippet` and `ConfigComment` for the opt-in configuration entry, following the same pattern as other opt-in kits.

#### Scenario: Config snippet format
- **WHEN** a new config is generated
- **THEN** the RTK kit SHALL appear as a commented-out entry: `# rtk:` with a brief description

### Requirement: RTK documentation page
A documentation page SHALL exist at `docs/kits/rtk.md` describing the kit, its configuration, and usage.

#### Scenario: Documentation exists
- **WHEN** a user looks for RTK kit documentation
- **THEN** `docs/kits/rtk.md` SHALL describe activation, what's included, and basic usage examples
