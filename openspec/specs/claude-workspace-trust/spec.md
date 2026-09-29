# claude-workspace-trust Specification

## Purpose
Lets Claude Code in an asylum container start without the workspace trust dialog, because the container is the trust boundary, while letting users restore the dialog through config.

## Requirements

### Requirement: Claude containers skip the workspace trust dialog
The Claude agent SHALL contribute `CLAUDE_CODE_SANDBOXED=1` to the container environment. The variable reaches the container whether Claude is the primary agent or a companion. Asylum SHALL NOT write trust entries into Claude's config file, because under `shared` isolation that file is the host's and host sessions would inherit the trust.

#### Scenario: Claude as primary agent
- **WHEN** a container is started with `claude` as the agent
- **THEN** the container environment contains `CLAUDE_CODE_SANDBOXED=1`

#### Scenario: Claude as companion
- **WHEN** a container is started with another primary agent and `claude` as a companion
- **THEN** the container environment contains `CLAUDE_CODE_SANDBOXED=1`

#### Scenario: Other agents are unaffected
- **WHEN** a container is started without Claude as primary agent or companion
- **THEN** no `CLAUDE_CODE_SANDBOXED` variable is contributed

#### Scenario: No dialog in the home directory
- **WHEN** Claude Code starts in a container whose project directory is the user's home directory
- **THEN** no workspace trust dialog is shown

### Requirement: Users can restore the trust dialog
A user-configured `env:` entry for `CLAUDE_CODE_SANDBOXED` SHALL override the agent's value. The documentation SHALL name the empty string as the way to turn the behaviour off, because Claude Code treats any non-empty value, including `"0"`, as on.

#### Scenario: Empty value in config
- **WHEN** the config contains `env: {CLAUDE_CODE_SANDBOXED: ""}`
- **THEN** the container environment contains `CLAUDE_CODE_SANDBOXED=` with an empty value and the agent's `1` is not emitted
