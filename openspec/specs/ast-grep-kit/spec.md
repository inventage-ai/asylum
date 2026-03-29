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

