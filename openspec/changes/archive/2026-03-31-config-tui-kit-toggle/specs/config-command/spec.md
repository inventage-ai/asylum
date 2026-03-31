## MODIFIED Requirements

### Requirement: Kits tab shows multiselect of all kits
The Kits tab SHALL display a multiselect list of all registered kits (excluding always-on kits). Kits that are currently active (present in config and not disabled) SHALL be pre-selected.

#### Scenario: Kit list population
- **WHEN** the Kits tab is displayed
- **THEN** all registered kits with tier Default or OptIn are listed, with currently active (not disabled) kits checked

#### Scenario: Toggle kit selection
- **WHEN** the user presses space on a kit entry
- **THEN** the kit's selection state is toggled (checked/unchecked)

### Requirement: Confirm applies all changes
When the user presses Enter, all changes across all tabs SHALL be applied to `~/.asylum/config.yaml`.

#### Scenario: Apply kit activation
- **WHEN** the user selects a previously inactive kit and presses Enter
- **THEN** the kit is activated using the appropriate strategy: if the kit has never been enabled (exists only as a comment block), the comment is removed and a clean YAML entry is inserted; if the kit was previously enabled and has `disabled: true`, the `disabled` field is removed

#### Scenario: Apply kit deactivation
- **WHEN** the user deselects a previously active kit and presses Enter
- **THEN** `disabled: true` is added as the first property of the kit's existing YAML entry

#### Scenario: Apply credential change
- **WHEN** the user toggles a credential kit and presses Enter
- **THEN** the kit's `credentials` value is updated to `auto` (if selected) or `false` (if deselected)

#### Scenario: Apply isolation change
- **WHEN** the user selects a different isolation level and presses Enter
- **THEN** the agent's config isolation is updated in the config file

#### Scenario: Cancel discards changes
- **WHEN** the user presses Escape
- **THEN** no changes are written and the command exits

### Requirement: Kit comment removal on activation
When activating a never-enabled kit, the system SHALL detect and remove any existing commented-out config block for that kit before inserting the active config entry.

#### Scenario: Commented kit with options
- **WHEN** a kit has a commented-out block like `# python:` followed by commented option lines (e.g. `#   versions:`, `#     - 3.14`)
- **THEN** all commented lines belonging to that kit block are removed before the active entry is inserted

#### Scenario: No commented version exists
- **WHEN** a kit has no commented-out config in the file
- **THEN** the active entry is inserted normally without any removal step

### Requirement: Kit entry removal on deactivation
When deactivating a kit, the system SHALL NOT remove the kit's YAML entry. Instead, it SHALL add `disabled: true` to the existing entry.

#### Scenario: Active kit with nested config
- **WHEN** a kit `java:` has nested lines like `versions:`, `default-version:` and is deactivated
- **THEN** `disabled: true` is added as the first property under `java:`, and all other config is preserved
