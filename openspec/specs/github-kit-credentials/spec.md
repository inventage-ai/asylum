### Requirement: GitHub credential provider
The GitHub kit SHALL provide a `CredentialFunc` that extracts the `gh` auth token from the host via `gh auth token` and generates a `hosts.yml` file for the container.

#### Scenario: gh authenticated on host
- **WHEN** `credentials: auto` is set for the GitHub kit and `gh auth token` returns a token
- **THEN** the credential func SHALL return a `CredentialMount` with `Content` containing a valid `hosts.yml` and `Destination` set to `~/.config/gh/hosts.yml`

#### Scenario: gh not authenticated on host
- **WHEN** `credentials: auto` is set but `gh auth token` fails or returns empty
- **THEN** the credential func SHALL return an empty result without error

#### Scenario: gh not installed on host
- **WHEN** `credentials: auto` is set but `gh` is not available on the host
- **THEN** the credential func SHALL return an empty result without error

#### Scenario: Credentials not enabled
- **WHEN** `credentials` is not set or set to `false` for the GitHub kit
- **THEN** the credential func SHALL NOT be called

### Requirement: GitHub credential label
The GitHub kit SHALL define `CredentialLabel` as `"GitHub"` for display in the onboarding wizard.

#### Scenario: Onboarding display
- **WHEN** the onboarding wizard lists credential-capable kits
- **THEN** the GitHub kit SHALL appear with label `"GitHub"`
