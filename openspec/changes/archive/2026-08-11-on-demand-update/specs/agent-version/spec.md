## MODIFIED Requirements

### Requirement: Background refresh on subsequent runs

When the version file exists and is valid, asylum SHALL load it from disk (instantly) and proceed with the build. A background goroutine SHALL refresh the file when it is considered stale. The file is stale when it is older than the configured staleness interval (default 24 hours), OR when it is missing a version entry for one or more tracked agents (e.g. an agent whose fetch failed during a previous partial fetch). When stale, the goroutine fetches all agent versions and updates the file.

#### Scenario: Background fetch is skipped
- **WHEN** the version file was updated less than the configured interval ago and contains an entry for every tracked agent
- **THEN** the background goroutine does nothing and the build proceeds with cached versions

#### Scenario: Background fetch succeeds
- **WHEN** the version file is stale and all fetches succeed
- **THEN** the file is updated with new versions and no error is reported to the user

#### Scenario: Partial fetch is retried before the interval
- **WHEN** the version file is younger than the configured interval but is missing a version for a tracked agent
- **THEN** it is considered stale and the next run attempts to fetch the missing agent again

#### Scenario: Background fetch fails
- **WHEN** the version file is stale and fetches fail
- **THEN** the failure is silently ignored and the cached versions remain valid

#### Scenario: Background fetch is fire-and-forget
- **WHEN** a background fetch is in progress
- **THEN** it does not block the current run; the next run picks up any new versions

## ADDED Requirements

### Requirement: Configurable staleness interval

The staleness interval used to decide whether the background refresh runs SHALL be configurable via the top-level `version-check-interval` config field. The value SHALL be a Go duration string (e.g. `24h`, `1h`, `168h`). When the field is unset, empty, or unparseable, asylum SHALL fall back to a default of 24 hours. The resolved interval SHALL be subject to the normal layered config merge, so a project-level `.asylum` file can override the global value.

#### Scenario: Default interval when unset
- **WHEN** no `version-check-interval` is configured
- **THEN** the staleness interval used for the background refresh is 24 hours

#### Scenario: Configured interval is honoured
- **WHEN** `version-check-interval: 1h` is configured
- **THEN** the background refresh treats `versions.json` as stale once it is older than 1 hour

#### Scenario: Invalid interval falls back to default
- **WHEN** `version-check-interval` is set to a value that is not a valid Go duration
- **THEN** asylum uses the 24-hour default and does not fail the run

#### Scenario: Project layer overrides global interval
- **WHEN** the global config sets `version-check-interval: 168h` and a project `.asylum` sets `version-check-interval: 1h`
- **THEN** the resolved interval for that project is 1 hour
