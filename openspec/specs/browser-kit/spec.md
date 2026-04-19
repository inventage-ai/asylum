# browser-kit Specification

## Purpose
Provides AI-native browser automation inside the sandbox via agent-browser.

## Requirements

### Requirement: agent-browser kit registration
The system SHALL register an `agent-browser` kit via `init()` in `internal/kit/agent_browser.go` with name `"agent-browser"`, TierOptIn, and a dependency on the `node` kit.

#### Scenario: Kit is registered at startup
- **WHEN** the application starts
- **THEN** the kit registry contains an `"agent-browser"` entry with Tier set to TierOptIn and Deps containing `"node"`

### Requirement: Backward-compatible alias
The system SHALL resolve the old kit name `"browser"` to `"agent-browser"` so existing configs continue to work.

#### Scenario: Old config key still works
- **WHEN** a user config contains `browser:` under `kits:`
- **THEN** `kit.Resolve` treats it as `"agent-browser"` without error

### Requirement: agent-browser and Chrome installation
The kit SHALL provide a DockerSnippet that installs the agent-browser npm package and then runs `agent-browser install --with-deps` (as root) to install Chrome with all required system libraries.

#### Scenario: agent-browser installed in image
- **WHEN** the agent-browser kit is active and the Docker image is built
- **THEN** the `agent-browser` CLI is available and Chrome is installed

### Requirement: Claude Code skill generation
The kit SHALL generate the agent-browser Claude Code skill at image-build time using `npx skills add vercel-labs/agent-browser` and stage it under the shared kit-skills root at `/opt/asylum-skills/.claude/skills/agent-browser/`. The kit SHALL set `ProvidesSkills: true` so the Claude agent launcher passes `--add-dir /opt/asylum-skills`. The kit SHALL NOT emit an entrypoint snippet that creates directories under `$HOME/.claude/skills/` or that bind-mounts the staged skill into `$HOME/.claude/skills/agent-browser/`.

#### Scenario: Skill staged under shared root at build time
- **WHEN** the Docker image is built with the agent-browser kit active
- **THEN** `/opt/asylum-skills/.claude/skills/agent-browser/SKILL.md` exists in the image and contains the upstream skill content

#### Scenario: Skill discoverable via --add-dir at runtime
- **WHEN** the container starts with the agent-browser kit active and the configured agent is Claude
- **THEN** the Claude launch command includes `--add-dir /opt/asylum-skills` and the skill is discoverable by Claude

#### Scenario: Host ~/.claude/skills/agent-browser not created in shared mode
- **WHEN** the container runs in shared agent-config mode and `~/.claude/skills/agent-browser/` does not exist on the host before the run
- **THEN** after container exit, `~/.claude/skills/agent-browser/` still does not exist on the host

### Requirement: agent-browser config snippet
The kit SHALL provide a ConfigSnippet and ConfigNodes so that kit sync can add an `agent-browser` entry to the user's config file.

#### Scenario: Config entry added during kit sync
- **WHEN** kit sync detects agent-browser as a new kit
- **THEN** an `agent-browser:` entry with a descriptive comment is added to the kits section of `config.yaml`

### Requirement: agent-browser tools metadata
The kit SHALL declare `Tools: []string{"agent-browser"}` so the tool is listed in aggregated tool output.

#### Scenario: Tool listed in aggregated tools
- **WHEN** `AggregateTools` is called with active kits including agent-browser
- **THEN** the result contains `"agent-browser (agent-browser)"`

### Requirement: agent-browser banner line
The kit SHALL provide a BannerLines entry that prints the agent-browser version in the welcome banner.

#### Scenario: Version shown in banner
- **WHEN** the container starts with agent-browser kit active
- **THEN** the welcome banner includes a line showing the agent-browser version

### Requirement: agent-browser rules snippet
The kit SHALL provide a RulesSnippet describing the agent-browser workflow: open, snapshot, interact using refs, re-snapshot.

#### Scenario: Rules file contains agent-browser section
- **WHEN** sandbox rules are assembled with agent-browser kit active
- **THEN** the rules file contains a section describing the agent-browser core workflow
