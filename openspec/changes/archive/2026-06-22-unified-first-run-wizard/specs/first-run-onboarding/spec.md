## MODIFIED Requirements

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

## ADDED Requirements

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
