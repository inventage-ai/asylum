# kit-state-tracking Specification

## Purpose

Remembers which kits an installation has already seen, in `~/.asylum/state.json`, so asylum can tell a genuinely new kit from one the user already declined. Without that record every upgrade would re-offer every optional kit. The list is updated once the sync flow completes.

## Requirements

### Requirement: Persistent kit state
The system SHALL maintain a state file at `~/.asylum/state.json` that tracks which kits the installation has previously seen.

#### Scenario: First run with no state file
- **WHEN** asylum starts and `~/.asylum/state.json` does not exist
- **THEN** the file is created with `known_kits` set to the full list of currently registered kit names

#### Scenario: State file exists
- **WHEN** asylum starts and `~/.asylum/state.json` exists
- **THEN** the file is loaded and `known_kits` is compared against the current registry

#### Scenario: State file deleted
- **WHEN** the user deletes `~/.asylum/state.json` and restarts asylum
- **THEN** all kits are treated as newly seen and the sync flow runs for each

### Requirement: New kit detection
The system SHALL detect kits that are registered but not present in the `known_kits` list.

#### Scenario: New kit added in update
- **WHEN** a new kit `rust` is registered and `known_kits` does not contain `rust`
- **THEN** `rust` is identified as a new kit and passed to the config sync flow

#### Scenario: All kits known
- **WHEN** every registered kit name is present in `known_kits`
- **THEN** no sync action is taken

#### Scenario: Kit removed from registry
- **WHEN** `known_kits` contains `legacy` but no kit named `legacy` is registered
- **THEN** the extra name is ignored (no cleanup, no error)

### Requirement: State update after sync
After the config sync flow completes, the system SHALL update `known_kits` to include all currently registered kit names and write the state file.

#### Scenario: State updated after sync
- **WHEN** new kits are detected and the sync flow completes (regardless of user prompt answers)
- **THEN** `state.json` is written with `known_kits` containing all registered kit names
