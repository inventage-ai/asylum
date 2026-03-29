## MODIFIED Requirements

### Requirement: Credential mount generation
When a kit's CredentialFunc returns CredentialMount entries, the system SHALL process each mount based on its type: if `HostPath` is set, the system SHALL bind-mount the host path directly into the container read-only at the specified `Destination`; if `Content` is set, the system SHALL write the content to `~/.asylum/projects/<container-name>/credentials/` and bind-mount it read-only at the specified `Destination`.

#### Scenario: Host path mount
- **WHEN** CredentialFunc returns a mount with `HostPath` set and `Destination` set to `~/.config/gh/`
- **THEN** the system SHALL bind-mount the host path directly at `~/.config/gh/` with mode `ro`

#### Scenario: Content-based mount (existing behavior)
- **WHEN** CredentialFunc returns a mount with `Content` set and `Destination` set to `~/.m2/settings.xml`
- **THEN** the system SHALL write the content to `~/.asylum/projects/<cname>/credentials/settings.xml` and bind-mount it at `~/.m2/settings.xml` with mode `ro`

#### Scenario: Credential mount ordering
- **WHEN** the kit also declares CacheDirs that overlap with the credential destination (e.g. `~/.m2`)
- **THEN** the credential bind mount SHALL be added after the cache volume mount

#### Scenario: CredentialFunc returns empty
- **WHEN** CredentialFunc returns an empty slice and no error
- **THEN** the system SHALL not create any credential mounts for that kit

#### Scenario: CredentialFunc returns error
- **WHEN** CredentialFunc returns an error
- **THEN** the system SHALL log a warning and continue container launch without credentials for that kit
