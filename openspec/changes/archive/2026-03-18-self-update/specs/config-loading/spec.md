## ADDED Requirements

### Requirement: Release channel config field
The config system SHALL support an optional `release-channel` scalar field with values `stable` or `dev`. It follows scalar merge semantics (last value wins across layers).

#### Scenario: Release channel set in global config
- **WHEN** `~/.asylum/config.yaml` contains `release-channel: dev`
- **THEN** the loaded config has `ReleaseChannel` set to `"dev"`

#### Scenario: Project config overrides global
- **WHEN** global config has `release-channel: dev` and project config has `release-channel: stable`
- **THEN** the merged result has `ReleaseChannel` set to `"stable"`

#### Scenario: Not set defaults to empty
- **WHEN** no config file sets `release-channel`
- **THEN** the loaded config has `ReleaseChannel` set to `""` (callers treat empty as stable)
