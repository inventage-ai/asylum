# config-defaults

## Purpose

Define the default `~/.asylum/config.yaml` written on first run: its content, the commented examples that make options discoverable, and how kit-provided snippets are assembled into it.
## Requirements
### Requirement: Default config on first run
When `~/.asylum/config.yaml` does not exist, the system SHALL write a default config file with all default values and descriptive comments.

#### Scenario: First run creates config
- **WHEN** asylum starts and `~/.asylum/config.yaml` does not exist
- **THEN** the file is created with version 0.2, default agent (claude), default kits (java, python, node), and commented-out sections for additional options

#### Scenario: Config already exists
- **WHEN** asylum starts and `~/.asylum/config.yaml` already exists
- **THEN** no default config is written (existing config is loaded as-is or migrated)

### Requirement: Default config content
The default config SHALL be assembled from a header template, kit-provided ConfigSnippets (for active kits) and comments (for opt-in kits), and a footer template. The header SHALL document the `version-check-interval` field as a commented example, showing the default of `24h`, so users can discover and adjust the background version-refresh window.

#### Scenario: Default kits present
- **WHEN** the default config is examined
- **THEN** it contains active kits for java (versions 17, 21, 25; default 21), python, and node with their default settings, assembled from each kit's ConfigSnippet

#### Scenario: Optional sections commented out
- **WHEN** the default config is examined
- **THEN** optional agents (gemini, codex, opencode), optional kits (apt, shell, title), ports, volumes, and env sections are present but commented out with explanatory comments

#### Scenario: Version check interval documented
- **WHEN** the default config is examined
- **THEN** a commented `version-check-interval` entry is present with an explanatory comment and the `24h` default

### Requirement: Kit config sync
When new kits are detected (registered but not in state.json), the system SHALL insert their config entries into the existing config.yaml using yaml.Node tree manipulation, preserving comments and user edits.

#### Scenario: New kit added to existing config
- **WHEN** a new TierDefault kit is activated and config.yaml has a kits mapping
- **THEN** the kit's ConfigNodes are appended to the kits mapping, preserving existing content

#### Scenario: New opt-in kit added as comment
- **WHEN** a new TierOptIn kit is detected
- **THEN** it is added as a foot comment on the kits mapping node

