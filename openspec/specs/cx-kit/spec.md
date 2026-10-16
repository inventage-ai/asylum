# cx-kit Specification

## Purpose

Installs cx, a semantic code navigation tool for AI agents, as an opt-in kit, with the language grammars chosen in the kit's `packages` list. Agents learn how to use it from a rules file generated at image build time.

## Requirements
### Requirement: cx kit registration
The system SHALL register a `cx` kit via `init()` in `internal/kit/cx.go` with name `"cx"` and TierOptIn. The kit SHALL have no kit dependencies.

#### Scenario: Kit is registered at startup
- **WHEN** the application starts
- **THEN** the kit registry contains a `"cx"` entry with Tier set to TierOptIn and no Deps

### Requirement: cx installation via install script
The kit SHALL provide a DockerSnippet that installs the cx CLI by downloading and running the install script from the cx repository. After installation, the snippet SHALL also run `cx skill` to generate the rules file at `/tmp/asylum-kit-rules/cx.md`. The installed binary SHALL be available on PATH.

#### Scenario: cx installed and rules generated in image
- **WHEN** the cx kit is active and the Docker image is built
- **THEN** the `cx` command is available on PATH inside the container
- **AND** `/tmp/asylum-kit-rules/cx.md` contains the output of `cx skill`

### Requirement: cx config snippet with languages
The kit SHALL provide a ConfigSnippet and ConfigNodes so that kit sync can add a `cx` entry to the user's config file. The config snippet SHALL include a commented-out `packages` list showing example language grammars (e.g., python, typescript, go).

#### Scenario: Config entry added during kit sync
- **WHEN** kit sync detects cx as a new kit
- **THEN** a `cx:` entry with a descriptive comment and example languages is added to the kits section of `config.yaml`

### Requirement: cx language installation during Docker build
When languages are configured via the `packages` field in the cx kit config, the system SHALL install those tree-sitter language grammars by running `cx lang add <language>` for each entry during the project image build.

#### Scenario: Languages installed in project image
- **WHEN** the cx kit config contains `packages: [python, typescript, go]`
- **THEN** the project image build runs `cx lang add python`, `cx lang add typescript`, `cx lang add go`

#### Scenario: No languages configured
- **WHEN** the cx kit config has no `packages` field
- **THEN** no `cx lang add` commands are run and cx is available with no pre-installed grammars

### Requirement: cx tools metadata
The kit SHALL declare `Tools: []string{"cx"}` so the tool is listed in aggregated tool output.

#### Scenario: Tool listed in aggregated tools
- **WHEN** `AggregateTools` is called with active kits including cx
- **THEN** the result contains `"cx (cx)"`

### Requirement: cx banner line
The kit SHALL provide a BannerLines entry that prints the cx version in the welcome banner.

#### Scenario: Version shown in banner
- **WHEN** the container starts with cx kit active
- **THEN** the welcome banner includes a line showing the cx version

### Requirement: cx rules snippet
The cx kit SHALL NOT populate the `RulesSnippet` field. Instead, the cx kit SHALL generate a standalone rules file via `cx skill` during the Docker build and place it in the agent's rules directory via the entrypoint. The kit SHALL retain its `Tools` field so that `cx` continues to appear in the aggregated "Kit Tools" list.

#### Scenario: Rules file contains cx section
- **WHEN** sandbox rules are assembled with cx kit active
- **THEN** the assembled `asylum-sandbox.md` SHALL NOT contain a cx-specific rules snippet
- **AND** the cx tool SHALL still appear in the "Kit Tools" section

#### Scenario: Standalone cx rules file present
- **WHEN** the container starts with cx kit active and `cx skill` succeeded during build
- **THEN** a standalone `cx.md` rules file SHALL be bind-mounted into the agent's rules directory

