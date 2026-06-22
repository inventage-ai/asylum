## ADDED Requirements

### Requirement: Agent selection step
On first-run invocations, the wizard SHALL include a multi-select step listing all registered agents except `echo` (a test stub), with `claude` pre-checked. Selection SHALL write `agent:` and `agents:` entries to `~/.asylum/config.yaml`. The step SHALL be skipped when not first-run, regardless of how many agents are configured.

#### Scenario: First run with defaults accepted
- **WHEN** the user runs `asylum` on first run and presses enter without changing the agent selection
- **THEN** `agent: claude` SHALL be written to `~/.asylum/config.yaml` and the `agents:` map SHALL contain `claude: {}` only

#### Scenario: First run with multiple agents picked
- **WHEN** the user selects claude and gemini in the agent step
- **THEN** the wizard SHALL present a follow-up single-select default-agent step

#### Scenario: First run with one non-default agent picked
- **WHEN** the user deselects claude and selects gemini only
- **THEN** `agent: gemini` SHALL be written and the default-agent step SHALL be skipped

#### Scenario: Echo hidden from picker
- **WHEN** the agent multi-select is presented
- **THEN** `echo` SHALL NOT appear as an option

#### Scenario: Subsequent run
- **WHEN** the user runs `asylum` on a non-first-run invocation
- **THEN** the agent step SHALL NOT be included in the wizard regardless of `cfg.Agent` value

### Requirement: Default-agent follow-up step
When the agent multi-select produced more than one selection, the wizard SHALL present a single-select step picking the default agent. If `claude` is among the selections, it SHALL be pre-selected; otherwise the first selected agent SHALL be pre-selected.

#### Scenario: Claude among multi-selected
- **WHEN** the user picked claude and gemini
- **THEN** the default-agent step SHALL pre-select claude

#### Scenario: Claude not picked
- **WHEN** the user picked gemini and codex (no claude)
- **THEN** the default-agent step SHALL pre-select gemini (the first picked)

#### Scenario: Single selection
- **WHEN** the user picked exactly one agent
- **THEN** the default-agent step SHALL be skipped

### Requirement: Kit selection step
On first-run invocations, the wizard SHALL include a multi-select step listing top-level kits (registry entries with no `/` in `Name`) excluding kits with `Tier == TierAlwaysOn`. `TierDefault` kits SHALL be pre-checked; `TierAvailable` kits SHALL be unchecked. Selection SHALL write uncommented `kits:` entries for chosen kits and commented entries for unchosen kits, matching the existing comment-vs-active pattern used by `WriteDefaults`. The step SHALL be skipped when not first-run.

#### Scenario: First run with defaults accepted
- **WHEN** the user presses enter without changing the kit selection
- **THEN** all `TierDefault` top-level kits SHALL be written as active entries in `~/.asylum/config.yaml` and all `TierAvailable` top-level kits SHALL be written as commented entries — identical to today's `WriteDefaults` output

#### Scenario: First run with TierAvailable kit enabled
- **WHEN** the user selects the `java` kit (TierAvailable) in the kit step
- **THEN** `java:` SHALL be written as an active entry under `kits:` in `~/.asylum/config.yaml`

#### Scenario: First run with TierDefault kit deselected
- **WHEN** the user deselects the `node` kit (TierDefault)
- **THEN** `node:` SHALL be written as a commented entry under `kits:` in `~/.asylum/config.yaml`

#### Scenario: Always-on kits excluded
- **WHEN** the kit multi-select is presented
- **THEN** kits with `Tier == TierAlwaysOn` (e.g. `ssh`, `ports`, `shell`, `node`'s always-on parts) SHALL NOT appear in the options

#### Scenario: Sub-kits excluded
- **WHEN** the kit multi-select is presented
- **THEN** kits whose `Name` contains `/` (e.g. `java/maven`, `python/pip`) SHALL NOT appear as separate options

#### Scenario: Subsequent run
- **WHEN** the user runs `asylum` on a non-first-run invocation
- **THEN** the kit step SHALL NOT be included in the wizard

### Requirement: Welcome banner
On first-run invocations where the wizard will present at least one step, the wizard SHALL print a one-line welcome banner before the first step.

#### Scenario: First run with steps
- **WHEN** the wizard is presented on first run
- **THEN** a welcome line ("Welcome to asylum — let's set up your sandbox.") SHALL be shown before step 1

#### Scenario: Non-first run
- **WHEN** the wizard is presented for an existing user (only isolation or credentials)
- **THEN** the welcome banner SHALL NOT be shown

## MODIFIED Requirements

### Requirement: Unified onboarding flow
Before starting a container, the system SHALL collect all pending onboarding steps and present them as a single wizard flow. On first-run invocations, the flow SHALL include the agent selection step, the optional default-agent step, the kit selection step, the isolation step (if applicable), and the credentials step (if applicable). On non-first-run invocations, the flow SHALL include only the isolation and credentials steps when their values are unconfigured. The wizard SHALL run before `ensureImages` so image-shaping selections affect the build. If no steps are pending, the wizard SHALL be skipped entirely.

#### Scenario: First run with all steps
- **WHEN** the user runs `asylum` on a fresh install
- **THEN** the wizard SHALL present agents, kits, isolation, and credentials in sequence (with the default-agent step appearing when applicable), all before any image build

#### Scenario: Multiple pending steps for existing user
- **WHEN** both config isolation and kit credentials are unconfigured and the invocation is not first-run
- **THEN** the wizard SHALL present isolation and credentials in sequence (no agents/kits steps)

#### Scenario: Single pending step for existing user
- **WHEN** only config isolation is unconfigured and the invocation is not first-run
- **THEN** the wizard SHALL present one step

#### Scenario: No pending steps
- **WHEN** all onboarding options are already configured and the invocation is not first-run
- **THEN** the wizard SHALL be skipped and the system SHALL proceed to `ensureImages`

#### Scenario: Non-interactive mode
- **WHEN** stdin is not a TTY
- **THEN** the system SHALL skip the wizard and use defaults for unconfigured options

### Requirement: Onboarding step detection
The system SHALL detect pending onboarding steps using two signals: first-run state (for agents and kits) and "not configured" state (for isolation and credentials). Each step fires when its trigger applies.

#### Scenario: First-run state triggers agents and kits
- **WHEN** the invocation is first-run (no `~/.asylum/config.yaml` at startup)
- **THEN** the agents and kits steps SHALL be included

#### Scenario: Non-first-run skips agents and kits
- **WHEN** the invocation is not first-run
- **THEN** the agents and kits steps SHALL NOT be included regardless of current config values

#### Scenario: Isolation not configured
- **WHEN** `agents.<agent>.config` is not set in the loaded config
- **THEN** the isolation step SHALL be included in the wizard

#### Scenario: Credentials not configured for active kit
- **WHEN** an active kit has a non-nil CredentialFunc and its parent kit has no `credentials` config
- **THEN** the credentials step SHALL be included in the wizard

#### Scenario: V1 migration user
- **WHEN** the user migrated from v1 (has `~/.asylum/agents/` but no credential config) and `~/.asylum/config.yaml` exists
- **THEN** the credential step SHALL appear in the wizard, but the agents and kits steps SHALL NOT
