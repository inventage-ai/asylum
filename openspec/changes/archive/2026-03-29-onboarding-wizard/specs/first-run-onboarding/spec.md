## MODIFIED Requirements

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

### Requirement: First-run detection
The system SHALL detect a first-run condition by checking whether `~/.asylum/agents/` directory exists. If it does not exist, the system SHALL trigger the first-run onboarding flow before loading config. The `agents/` directory is created by `EnsureAgentConfig` on the first actual run, making it a reliable signal that distinguishes fresh installs from existing users (since the installer only creates `~/.asylum/bin/`).

#### Scenario: First run — agents directory does not exist
- **WHEN** the user runs `asylum` and `~/.asylum/agents/` does not exist
- **THEN** the system SHALL run the first-run onboarding flow before proceeding

#### Scenario: Subsequent run — agents directory exists
- **WHEN** the user runs `asylum` and `~/.asylum/agents/` already exists
- **THEN** the system SHALL skip first-run onboarding and proceed normally
