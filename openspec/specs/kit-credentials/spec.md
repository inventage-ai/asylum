### Requirement: Kit credential function
The Kit struct SHALL support an optional `CredentialFunc` field of type `func(CredentialOpts) ([]CredentialMount, error)`. Kits that do not handle credentials SHALL leave this field nil.

#### Scenario: Kit with credential support
- **WHEN** a kit defines a non-nil CredentialFunc
- **THEN** the system SHALL call it during container launch if credentials are enabled for that kit

#### Scenario: Kit without credential support
- **WHEN** a kit has a nil CredentialFunc
- **THEN** the system SHALL skip credential processing for that kit

### Requirement: Credential configuration
KitConfig SHALL support a `credentials` field that accepts three forms: the string `"auto"`, a list of strings (kit-specific identifiers), or `false`/absent (off). The default when not specified SHALL be off.

#### Scenario: Credentials set to auto
- **WHEN** `credentials: auto` is set for a kit
- **THEN** the kit's CredentialFunc SHALL be called with mode `auto`

#### Scenario: Credentials set to explicit list
- **WHEN** `credentials` is set to a list of strings (e.g. `[nexus-releases, nexus-snapshots]`)
- **THEN** the kit's CredentialFunc SHALL be called with mode `explicit` and the list passed as identifiers

#### Scenario: Credentials off by default
- **WHEN** `credentials` is not set or set to `false`
- **THEN** the kit's CredentialFunc SHALL NOT be called

### Requirement: Credential mount generation
When a kit's CredentialFunc returns CredentialMount entries, the system SHALL write each mount's content to `~/.asylum/projects/<container-name>/credentials/` and bind-mount it read-only into the container at the specified destination path.

#### Scenario: Credential file written and mounted
- **WHEN** CredentialFunc returns a mount with content and destination `~/.m2/settings.xml`
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

### Requirement: Maven auto credential discovery
The java/maven sub-kit SHALL provide a CredentialFunc that, in auto mode, reads the root `pom.xml` of the project directory and extracts server IDs from `<repositories>`, `<pluginRepositories>`, `<distributionManagement>`, and the same elements within `<profiles>`.

#### Scenario: pom.xml with repositories
- **WHEN** `pom.xml` contains `<repository>` entries with `<id>` elements
- **THEN** the CredentialFunc SHALL extract those IDs as requested server IDs

#### Scenario: pom.xml with plugin repositories
- **WHEN** `pom.xml` contains `<pluginRepository>` entries with `<id>` elements
- **THEN** the CredentialFunc SHALL extract those IDs as requested server IDs

#### Scenario: pom.xml with distribution management
- **WHEN** `pom.xml` contains `<distributionManagement>` with `<repository>` or `<snapshotRepository>` entries
- **THEN** the CredentialFunc SHALL extract their `<id>` elements as requested server IDs

#### Scenario: pom.xml with profiles
- **WHEN** `pom.xml` contains `<profiles>` with repository definitions inside individual profiles
- **THEN** the CredentialFunc SHALL extract server IDs from those profile-scoped repositories

#### Scenario: No pom.xml
- **WHEN** the project directory does not contain a `pom.xml`
- **THEN** the CredentialFunc SHALL return an empty result without error

### Requirement: Maven settings.xml filtering
The java/maven CredentialFunc SHALL read `~/.m2/settings.xml` from the host and generate a minimal settings.xml containing only `<server>` entries whose `<id>` matches a requested server ID.

#### Scenario: Matching servers found
- **WHEN** `~/.m2/settings.xml` contains `<server>` entries matching requested IDs
- **THEN** the generated settings.xml SHALL include only those matching server entries

#### Scenario: Server ID not found in settings.xml
- **WHEN** a requested server ID is not found in `~/.m2/settings.xml`
- **THEN** the generated settings.xml SHALL include an XML comment: `<!-- server "<id>" referenced in pom.xml but not found in ~/.m2/settings.xml -->`

#### Scenario: No settings.xml on host
- **WHEN** `~/.m2/settings.xml` does not exist on the host
- **THEN** the CredentialFunc SHALL return an empty result without error

#### Scenario: No matching servers
- **WHEN** no requested server IDs match any entries in `~/.m2/settings.xml`
- **THEN** the CredentialFunc SHALL return an empty result without error

### Requirement: Maven explicit credential mode
In explicit mode, the java/maven CredentialFunc SHALL skip pom.xml parsing and use the provided list of strings directly as server IDs to look up in `~/.m2/settings.xml`.

#### Scenario: Explicit server IDs
- **WHEN** credentials are configured as a list `[nexus-releases, nexus-snapshots]`
- **THEN** the CredentialFunc SHALL filter `~/.m2/settings.xml` for those IDs without reading pom.xml
