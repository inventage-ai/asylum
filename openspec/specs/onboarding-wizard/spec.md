### Requirement: Unified onboarding flow
Before starting a container, the system SHALL collect all pending onboarding steps (configuration that is not yet set but has a prompt defined) and present them as a single wizard flow. If no steps are pending, the wizard SHALL be skipped entirely.

#### Scenario: Multiple pending steps
- **WHEN** both config isolation and kit credentials are unconfigured
- **THEN** the system SHALL present a wizard with both steps in sequence

#### Scenario: Single pending step
- **WHEN** only config isolation is unconfigured
- **THEN** the system SHALL present a wizard with one step

#### Scenario: No pending steps
- **WHEN** all onboarding options are already configured
- **THEN** the system SHALL skip the wizard and proceed to container start

#### Scenario: Non-interactive mode
- **WHEN** stdin is not a TTY
- **THEN** the system SHALL skip the wizard and use defaults for unconfigured options

### Requirement: Onboarding step detection
The system SHALL detect pending onboarding steps using "not configured" checks, not first-run detection. Each step fires when its config value is absent, regardless of whether the user is new or migrating.

#### Scenario: Isolation not configured
- **WHEN** `agents.<agent>.config` is not set in the loaded config
- **THEN** the isolation step SHALL be included in the wizard

#### Scenario: Credentials not configured for active kit
- **WHEN** an active kit has a non-nil CredentialFunc and its parent kit has no `credentials` config
- **THEN** the credentials step SHALL be included in the wizard

#### Scenario: V1 migration user
- **WHEN** the user migrated from v1 (has `~/.asylum/agents/` but no credential config)
- **THEN** the credential step SHALL appear in the wizard

### Requirement: Onboarding result application
After the wizard completes, the system SHALL apply all completed step results to `~/.asylum/config.yaml` and update the in-memory config. Steps that were not reached (due to cancel) SHALL not be applied.

#### Scenario: All steps completed
- **WHEN** the user completes all wizard steps
- **THEN** all results SHALL be written to config and applied in-memory

#### Scenario: Cancelled mid-wizard
- **WHEN** the user cancels at step 2 of 3
- **THEN** step 1's result SHALL be written to config, steps 2 and 3 SHALL not be applied

#### Scenario: Cancelled at first step
- **WHEN** the user cancels at step 1
- **THEN** no config SHALL be written
