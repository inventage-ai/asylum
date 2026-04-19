# ast-grep-kit Specification

## Purpose
TBD - created by archiving change add-plugin-kits. Update Purpose after archive.
## Requirements
### Requirement: ast-grep kit registration
The system SHALL register an `ast-grep` kit via `init()` in `internal/kit/astgrep.go` with name `"ast-grep"`, TierOptIn, and a dependency on the `node` kit.

#### Scenario: Kit is registered at startup
- **WHEN** the application starts
- **THEN** the kit registry contains an `"ast-grep"` entry with Tier set to TierOptIn and Deps containing `"node"`

### Requirement: ast-grep installation via Docker snippet
The kit SHALL provide a DockerSnippet that installs `@ast-grep/cli` globally via npm, making the `sg` command available in the container.

#### Scenario: ast-grep installed in image
- **WHEN** the ast-grep kit is active and the Docker image is built
- **THEN** the `sg` command is available on PATH inside the container

#### Scenario: npm environment used for install
- **WHEN** the DockerSnippet executes
- **THEN** it sources the fnm environment before running `npm install -g @ast-grep/cli`

### Requirement: ast-grep config snippet
The kit SHALL provide a ConfigSnippet and ConfigNodes so that kit sync can add an `ast-grep` entry to the user's config file.

#### Scenario: Config entry added during kit sync
- **WHEN** kit sync detects ast-grep as a new kit
- **THEN** an `ast-grep:` entry with a descriptive comment is added to the kits section of `config.yaml`

### Requirement: ast-grep tools metadata
The kit SHALL declare `Tools: []string{"sg"}` so the tool is listed in aggregated tool output.

#### Scenario: Tool listed in aggregated tools
- **WHEN** `AggregateTools` is called with active kits including ast-grep
- **THEN** the result contains `"sg (ast-grep)"`

### Requirement: ast-grep banner line
The kit SHALL provide a BannerLines entry that prints the ast-grep version in the welcome banner.

#### Scenario: Version shown in banner
- **WHEN** the container starts with ast-grep kit active
- **THEN** the welcome banner includes a line showing the ast-grep/sg version

### Requirement: ast-grep rules snippet
The kit SHALL provide a RulesSnippet describing ast-grep's availability and usage for agents.

#### Scenario: Rules file contains ast-grep section
- **WHEN** sandbox rules are assembled with ast-grep kit active
- **THEN** the rules file contains a section describing ast-grep and the `sg` command

### Requirement: ast-grep Claude Code skill delivery
The kit SHALL generate the ast-grep Claude Code skill at image-build time using `npx skills add ast-grep/agent-skill` and stage it under the shared kit-skills root at `/opt/asylum-skills/.claude/skills/ast-grep/`. The kit SHALL set `ProvidesSkills: true` so the Claude agent launcher passes `--add-dir /opt/asylum-skills`. The kit SHALL NOT emit an entrypoint snippet that creates directories under `$HOME/.claude/skills/` or that bind-mounts the staged skill into `$HOME/.claude/skills/ast-grep/`.

#### Scenario: Skill staged under shared root at build time
- **WHEN** the Docker image is built with the ast-grep kit active
- **THEN** `/opt/asylum-skills/.claude/skills/ast-grep/SKILL.md` exists in the image and contains the upstream skill content

#### Scenario: Skill discoverable via --add-dir at runtime
- **WHEN** the container starts with the ast-grep kit active and the configured agent is Claude
- **THEN** the Claude launch command includes `--add-dir /opt/asylum-skills` and the skill is discoverable by Claude

#### Scenario: Host ~/.claude/skills/ast-grep not created in shared mode
- **WHEN** the container runs in shared agent-config mode and `~/.claude/skills/ast-grep/` does not exist on the host before the run
- **THEN** after container exit, `~/.claude/skills/ast-grep/` still does not exist on the host

