## MODIFIED Requirements

### Requirement: Credential file detection
The system SHALL present a TUI multiselect prompt listing all kits that provide credential support (non-nil CredentialFunc), regardless of whether credential files exist on the host. The prompt SHALL allow the user to select which kits should have credential support enabled.

#### Scenario: Kits with credential support available
- **WHEN** the user runs asylum for the first time and there are kits with CredentialFunc defined
- **THEN** the system SHALL display a TUI multiselect with each credential-capable kit as an option

#### Scenario: No kits with credential support
- **WHEN** no registered kits have a CredentialFunc
- **THEN** the system SHALL skip the credential prompt entirely

#### Scenario: Non-interactive first run
- **WHEN** asylum starts non-interactively (stdin is not a TTY)
- **THEN** the system SHALL skip the credential prompt and leave credentials off (default)

### Requirement: Interactive credential mount prompt
When the user selects kits in the credential multiselect, the system SHALL write `credentials: auto` for each selected kit into `~/.asylum/config.yaml`.

#### Scenario: User selects kits
- **WHEN** the user selects Java/Maven in the multiselect and confirms
- **THEN** `~/.asylum/config.yaml` SHALL be updated with `kits: { java: { credentials: auto } }`

#### Scenario: User selects no kits
- **WHEN** the user confirms the multiselect with no kits selected
- **THEN** the system SHALL not write any credential config

#### Scenario: User cancels
- **WHEN** the user cancels the multiselect (esc)
- **THEN** the system SHALL not write any credential config

### Requirement: Config file generation
When the user accepts credential support for kits, the system SHALL write `credentials: auto` under each selected kit in `~/.asylum/config.yaml`, using yaml.Node manipulation to preserve existing config formatting and comments.

#### Scenario: Config updated for selected kits
- **WHEN** the user selects Java/Maven for credential support
- **THEN** `~/.asylum/config.yaml` SHALL contain `credentials: auto` under the java kit entry
