### Requirement: Hidden field on Kit struct
The Kit struct SHALL have a `Hidden bool` field. When set to true, the kit SHALL be excluded from all interactive selection surfaces.

#### Scenario: Hidden kit excluded from config TUI
- **WHEN** a kit has `Hidden: true`
- **THEN** it SHALL NOT appear in the Kits tab of `asylum config`

#### Scenario: Hidden kit excluded from new-kit sync prompt
- **WHEN** a kit has `Hidden: true` and is detected as a new kit during `SyncNewKits`
- **THEN** it SHALL NOT be included in the interactive prompt shown to the user
- **AND** it SHALL be silently added as a comment in the config file (same as non-interactive TierOptIn behavior)

#### Scenario: Hidden kit excluded from sandbox rules disabled list
- **WHEN** a kit has `Hidden: true` and is not active
- **THEN** it SHALL NOT appear in the "Disabled Kits" section of the sandbox rules file

#### Scenario: Hidden kit remains fully functional
- **WHEN** a kit has `Hidden: true` and is configured in `.asylum`
- **THEN** it SHALL activate, build, and contribute to the image exactly like any visible kit

### Requirement: apt kit is hidden
The `apt` kit SHALL have `Hidden: true` set.

#### Scenario: apt kit not shown in config TUI
- **WHEN** the user runs `asylum config`
- **THEN** the `apt` kit SHALL NOT appear in the Kits tab

#### Scenario: apt kit activates when configured
- **WHEN** the user adds `apt:` with packages to their `.asylum` config
- **THEN** the apt kit SHALL activate and install the specified packages
