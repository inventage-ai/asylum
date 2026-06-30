## MODIFIED Requirements

### Requirement: Blocking fetch on first run

When the version file does not exist or is corrupted, asylum SHALL perform a blocking fetch of all agent versions before proceeding with the image build. The image build SHALL NOT proceed until at least one valid version has been resolved. The fetches SHALL run concurrently so that the total blocking time is bounded by the slowest single source rather than the sum of all sources.

#### Scenario: Blocking fetch succeeds
- **WHEN** the version file is missing and all fetches succeed
- **THEN** the versions are written to `versions.json` and the image build proceeds

#### Scenario: Some fetches fail during blocking
- **WHEN** the version file is missing but some fetches fail
- **THEN** the successfully fetched versions are saved (with the missing ones omitted), and the build proceeds with available versions

#### Scenario: All fetches fail during blocking
- **WHEN** the version file is missing and all fetches fail
- **THEN** the build proceeds with an empty version map (no version pinning), same as current behavior

#### Scenario: Fetches run concurrently
- **WHEN** a blocking fetch resolves all six agent versions
- **THEN** the sources are queried concurrently and total blocking time is bounded by the slowest single source

### Requirement: Background refresh on subsequent runs

When the version file exists and is valid, asylum SHALL load it from disk (instantly) and proceed with the build. A background goroutine SHALL refresh the file when it is considered stale. The file is stale when it is older than 24 hours, OR when it is missing a version entry for one or more tracked agents (e.g. an agent whose fetch failed during a previous partial fetch). When stale, the goroutine fetches all agent versions and updates the file.

#### Scenario: Background fetch is skipped
- **WHEN** the version file was updated less than 24 hours ago and contains an entry for every tracked agent
- **THEN** the background goroutine does nothing and the build proceeds with cached versions

#### Scenario: Background fetch succeeds
- **WHEN** the version file is stale and all fetches succeed
- **THEN** the file is updated with new versions and no error is reported to the user

#### Scenario: Partial fetch is retried before the interval
- **WHEN** the version file is younger than 24 hours but is missing a version for a tracked agent
- **THEN** it is considered stale and the next run attempts to fetch the missing agent again

#### Scenario: Background fetch fails
- **WHEN** the version file is stale and fetches fail
- **THEN** the failure is silently ignored and the cached versions remain valid

#### Scenario: Background fetch is fire-and-forget
- **WHEN** a background fetch is in progress
- **THEN** it does not block the current run; the next run picks up any new versions

## ADDED Requirements

### Requirement: Concurrency-safe version file writes

Writing `versions.json` SHALL be safe under concurrent invocation. When two `asylum` processes (e.g. for different projects, sharing one `~/.asylum/versions.json`) write the file at the same time, neither SHALL observe a truncated or interleaved file. Writes SHALL use a per-write unique temporary file followed by an atomic rename onto the target path.

#### Scenario: Concurrent writes do not corrupt the file
- **WHEN** two asylum processes write `versions.json` concurrently
- **THEN** each writes to its own unique temporary file and atomically renames it into place, and the resulting file is always a complete, valid JSON object

#### Scenario: Failed write leaves the previous file intact
- **WHEN** a version file write fails before the rename
- **THEN** the existing `versions.json` is left unchanged and no partial temporary file replaces it
