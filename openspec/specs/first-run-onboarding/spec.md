# first-run-onboarding Specification

## Purpose

Handles the very first `asylum` invocation on a host, detected by the absence of `~/.asylum/config.yaml`. It offers credential support for whichever active kits provide it, writes the resulting config, and keeps all of that logic in `internal/firstrun/` rather than spread through the CLI entry point.

## Requirements

### Requirement: First-run detection
The system SHALL detect a first-run condition by checking whether `~/.asylum/config.yaml` exists at startup, captured before `WriteDefaults` runs. If the file does not exist, the system SHALL treat the invocation as first-run and SHALL run the full first-run wizard (agents + kits, in addition to isolation and credentials) before loading the resolved config used by `ensureImages`. The `~/.asylum/agents/` directory is no longer used as the first-run signal — it remains in use by the resume-migration prompt under different semantics.

#### Scenario: First run — config file does not exist
- **WHEN** the user runs `asylum` and `~/.asylum/config.yaml` does not exist
- **THEN** the system SHALL flag the invocation as first-run before `WriteDefaults` runs and SHALL trigger the full first-run wizard

#### Scenario: Subsequent run — config file exists
- **WHEN** the user runs `asylum` and `~/.asylum/config.yaml` already exists
- **THEN** the system SHALL skip the first-run agents and kits wizard steps and only run the per-step "is this value unconfigured" prompts (isolation, credentials)

#### Scenario: Non-interactive mode
- **WHEN** asylum starts non-interactively (stdin is not a TTY) on what would be a first-run invocation
- **THEN** the system SHALL skip the wizard entirely and SHALL apply today's silent defaults (claude only, TierDefault kits, isolated config, no credentials)

### Requirement: Credential file detection
The system SHALL include a credentials step in the onboarding wizard when any active kit has credential support (non-nil CredentialFunc) but no `credentials` config is set. The step SHALL be a multiselect listing all credential-capable kits. Detection is based on "not configured" status, not first-run detection.

#### Scenario: Kits with credential support available
- **WHEN** an active kit has CredentialFunc and no credentials config
- **THEN** the system SHALL include a credential multiselect step in the onboarding wizard

#### Scenario: No kits with credential support
- **WHEN** no active kits have a CredentialFunc
- **THEN** the system SHALL not include a credential step in the wizard

#### Scenario: Credentials already configured
- **WHEN** all credential-capable kits already have `credentials` set (auto, explicit list, or false)
- **THEN** the system SHALL not include a credential step in the wizard

#### Scenario: Non-interactive mode
- **WHEN** asylum starts non-interactively (stdin is not a TTY)
- **THEN** the system SHALL skip the credential step and leave credentials off (default)

### Requirement: Interactive credential mount prompt
When the user selects kits in the credential wizard step, the system SHALL write `credentials: auto` for each selected kit into `~/.asylum/config.yaml`.

#### Scenario: User selects kits
- **WHEN** the user selects Java/Maven in the credential step and completes the wizard
- **THEN** `~/.asylum/config.yaml` SHALL be updated with `kits: { java: { credentials: auto } }`

#### Scenario: User selects no kits
- **WHEN** the user confirms the credential step with no kits selected
- **THEN** the system SHALL not write any credential config

#### Scenario: User cancels before reaching credential step
- **WHEN** the user cancels the wizard before the credential step
- **THEN** the system SHALL not write any credential config

### Requirement: Config file generation
When the user accepts credential support for kits, the system SHALL write `credentials: auto` under each selected kit in `~/.asylum/config.yaml`, using yaml.Node manipulation to preserve existing config formatting and comments.

#### Scenario: Config updated for selected kits
- **WHEN** the user selects Java/Maven for credential support
- **THEN** `~/.asylum/config.yaml` SHALL contain `credentials: auto` under the java kit entry

### Requirement: First-run wizard ownership
The `internal/firstrun/` package SHALL own the first-run wizard build, presentation, and result persistence. `cmd/asylum/main.go` SHALL call `firstrun.Run(...)` once, after `config.Load` and before `ensureImages`, and SHALL re-invoke `config.Load` when the wizard wrote any image-shaping settings (agents, kits) so the rebuilt config drives image generation.

#### Scenario: Wizard runs before image build
- **WHEN** the first-run wizard is triggered
- **THEN** it SHALL complete before `EnsureBase`/`EnsureProject` are called

#### Scenario: Config reloaded when wizard wrote image-shaping settings
- **WHEN** the wizard writes agent or kit selections to `~/.asylum/config.yaml`
- **THEN** `config.Load` SHALL be invoked again before `ensureImages` so the merged config reflects the new layer

#### Scenario: Config not reloaded when wizard wrote only runtime settings
- **WHEN** the wizard only wrote isolation or credentials (no agents/kits changes)
- **THEN** the in-memory mutations from the wizard's appliers SHALL be sufficient and no extra `config.Load` is required
