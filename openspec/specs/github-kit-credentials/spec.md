### Requirement: GitHub credential provider
The GitHub kit SHALL provide a `CredentialFunc` that mounts the host's `~/.config/gh/` directory read-only into the container at `~/.config/gh/`.

#### Scenario: gh config directory exists on host
- **WHEN** `credentials: auto` is set for the GitHub kit and `~/.config/gh/` exists on the host
- **THEN** the credential func SHALL return a `CredentialMount` with `HostPath` pointing to the host directory and `Destination` set to `~/.config/gh/`

#### Scenario: gh config directory does not exist
- **WHEN** `credentials: auto` is set but `~/.config/gh/` does not exist on the host
- **THEN** the credential func SHALL return an empty result without error

#### Scenario: Credentials not enabled
- **WHEN** `credentials` is not set or set to `false` for the GitHub kit
- **THEN** the credential func SHALL NOT be called

### Requirement: GitHub credential label
The GitHub kit SHALL define `CredentialLabel` as `"GitHub"` for display in the onboarding wizard.

#### Scenario: Onboarding display
- **WHEN** the onboarding wizard lists credential-capable kits
- **THEN** the GitHub kit SHALL appear with label `"GitHub"`
