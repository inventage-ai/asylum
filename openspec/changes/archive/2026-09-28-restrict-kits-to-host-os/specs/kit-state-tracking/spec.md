# Spec Delta

## MODIFIED Requirements

### Requirement: New kit detection
The system SHALL detect kits that are registered, available on the host, and not present in the `known_kits` list. Kits that are not available on the host SHALL NOT be detected as new.

#### Scenario: New kit added in update
- **WHEN** a new kit `rust` is registered and `known_kits` does not contain `rust`
- **THEN** `rust` is identified as a new kit and passed to the config sync flow

#### Scenario: All kits known
- **WHEN** every registered kit name is present in `known_kits`
- **THEN** no sync action is taken

#### Scenario: Kit removed from registry
- **WHEN** `known_kits` contains `legacy` but no kit named `legacy` is registered
- **THEN** the extra name is ignored (no cleanup, no error)

#### Scenario: Kit for another host OS
- **WHEN** a macOS-only kit is registered, `known_kits` does not contain it, and the host is Linux
- **THEN** it is not identified as a new kit

### Requirement: State update after sync
After the config sync flow completes, the system SHALL update `known_kits` to include all currently registered kit names that are available on the host, and write the state file. The same SHALL hold when all kits are marked known without a prompt, for example on first run or after a config migration. Kits not available on the host SHALL NOT be added to `known_kits`.

#### Scenario: State updated after sync
- **WHEN** new kits are detected and the sync flow completes (regardless of user prompt answers)
- **THEN** `state.json` is written with `known_kits` containing all registered kit names that are available on the host

#### Scenario: Kit for another host OS stays unknown
- **WHEN** the state is written on a Linux host
- **THEN** `known_kits` does not contain `iterm` or `dropshare`, so a macOS host sharing the same `~/.asylum` still offers them
