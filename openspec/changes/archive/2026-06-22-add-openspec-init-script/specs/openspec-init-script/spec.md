## ADDED Requirements

### Requirement: OpenSpec init script on PATH
The `openspec` kit SHALL install an executable `asylum-openspec-init` script on the container PATH. The script SHALL initialize OpenSpec in the current project non-interactively using the project's preferred settings, requiring no flags or arguments from the caller.

#### Scenario: Script is available when openspec kit is active
- **WHEN** the openspec kit is active in a container
- **THEN** `asylum-openspec-init` SHALL be present on PATH and executable

#### Scenario: Script runs without arguments
- **WHEN** `asylum-openspec-init` is invoked with no arguments from a project directory
- **THEN** it SHALL complete OpenSpec setup without prompting for interactive input

### Requirement: Active agent mapped to OpenSpec tools id
The script SHALL select the OpenSpec `--tools` value from the active agent reported by the `ASYLUM_AGENT` environment variable. The agent name SHALL be passed through unchanged except `copilot`, which SHALL map to `github-copilot`.

#### Scenario: Claude agent
- **WHEN** `ASYLUM_AGENT` is `claude`
- **THEN** the script SHALL initialize OpenSpec with `--tools claude`

#### Scenario: Copilot agent translation
- **WHEN** `ASYLUM_AGENT` is `copilot`
- **THEN** the script SHALL initialize OpenSpec with `--tools github-copilot`

### Requirement: Idempotent init versus update
The script SHALL detect whether OpenSpec is already initialized in the project by the presence of an `openspec/` directory. It SHALL run a fresh initialization when absent and refresh the existing setup when present, so that re-running the script is safe.

#### Scenario: Fresh project
- **WHEN** the script runs and no `openspec/` directory exists
- **THEN** it SHALL run `openspec init` with the resolved `--tools` value

#### Scenario: Already initialized project
- **WHEN** the script runs and an `openspec/` directory already exists
- **THEN** it SHALL run `openspec update --force` instead of re-initializing

### Requirement: Preferred workflow set materialized
With the preferred OpenSpec global config in place, the script's initialization SHALL materialize the `propose`, `explore`, `apply`, `verify`, and `archive` workflow command and skill files, and SHALL NOT materialize the `sync` workflow.

#### Scenario: Verify workflow present after init
- **WHEN** the script initializes a fresh project
- **THEN** the generated OpenSpec command/skill set SHALL include the `verify` workflow and SHALL omit the `sync` workflow
