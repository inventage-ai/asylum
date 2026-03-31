## ADDED Requirements

### Requirement: Config subcommand launches tabbed TUI
The system SHALL provide an `asylum config` subcommand that launches an interactive tabbed TUI with three tabs: Kits, Credentials, and Isolation.

#### Scenario: Launch config command
- **WHEN** the user runs `asylum config`
- **THEN** a tabbed TUI is displayed with tabs "Kits", "Credentials", and "Isolation", with the "Kits" tab active by default

#### Scenario: Non-TTY environment
- **WHEN** `asylum config` is run without a TTY (e.g. piped)
- **THEN** the command SHALL exit with an error message indicating a terminal is required

### Requirement: Tab navigation with arrow keys
The user SHALL be able to switch between tabs using the left and right arrow keys (or h/l).

#### Scenario: Switch to next tab
- **WHEN** the user presses the right arrow key while on the "Kits" tab
- **THEN** the "Credentials" tab becomes active and its content is displayed

#### Scenario: Switch to previous tab
- **WHEN** the user presses the left arrow key while on the "Credentials" tab
- **THEN** the "Kits" tab becomes active and its content is displayed

#### Scenario: Wrap around at edges
- **WHEN** the user presses the right arrow key while on the "Isolation" tab (last tab)
- **THEN** the active tab SHALL NOT change (no wrap-around)

### Requirement: Kits tab shows multiselect of all kits
The Kits tab SHALL display a multiselect list of all registered kits (excluding always-on kits). Kits that are currently active (present in config and not disabled) SHALL be pre-selected.

#### Scenario: Kit list population
- **WHEN** the Kits tab is displayed
- **THEN** all registered kits with tier Default or OptIn are listed, with currently active (not disabled) kits checked

#### Scenario: Toggle kit selection
- **WHEN** the user presses space on a kit entry
- **THEN** the kit's selection state is toggled (checked/unchecked)

### Requirement: Credentials tab shows multiselect of credential-capable kits
The Credentials tab SHALL display a multiselect list of all kits that have credential providers. Kits with `credentials: auto` SHALL be pre-selected.

#### Scenario: Credentials list population
- **WHEN** the Credentials tab is displayed
- **THEN** all credential-capable kits are listed with their credential labels, pre-selected if currently configured as `auto`

### Requirement: Isolation tab shows single-select for config isolation
The Isolation tab SHALL display a single-select list with options: Shared, Isolated, and Project-isolated. The current isolation level SHALL be pre-selected.

#### Scenario: Isolation options display
- **WHEN** the Isolation tab is displayed
- **THEN** three options are shown: "Shared with host", "Isolated (recommended)", "Project-isolated", with the current setting highlighted

### Requirement: Confirm applies all changes
When the user presses Enter, all changes across all tabs SHALL be applied to `~/.asylum/config.yaml`.

#### Scenario: Apply kit activation (first time)
- **WHEN** the user selects a never-enabled kit (exists only as a comment block) and presses Enter
- **THEN** the comment block is removed and a clean YAML entry is inserted into the kits section

#### Scenario: Apply kit activation (re-enable)
- **WHEN** the user selects a previously-enabled kit that has `disabled: true` and presses Enter
- **THEN** the `disabled` field is removed from the kit's YAML entry

#### Scenario: Apply kit deactivation
- **WHEN** the user deselects a previously active kit and presses Enter
- **THEN** `disabled: true` is added as the first property of the kit's existing YAML entry, preserving all other configuration

#### Scenario: Apply credential change
- **WHEN** the user toggles a credential kit and presses Enter
- **THEN** the kit's `credentials` value is updated to `auto` (if selected) or `false` (if deselected)

#### Scenario: Apply isolation change
- **WHEN** the user selects a different isolation level and presses Enter
- **THEN** the agent's config isolation is updated in the config file

#### Scenario: Cancel discards changes
- **WHEN** the user presses Escape
- **THEN** no changes are written and the command exits

### Requirement: Kit comment removal on first activation
When activating a never-enabled kit, the system SHALL detect and remove any existing commented-out config block for that kit before inserting the active config entry.

#### Scenario: Commented kit with options
- **WHEN** a kit has a commented-out block like `# python:` followed by commented option lines (e.g. `#   versions:`, `#     - 3.14`)
- **THEN** all commented lines belonging to that kit block are removed before the active entry is inserted

#### Scenario: No commented version exists
- **WHEN** a kit has no commented-out config in the file
- **THEN** the active entry is inserted normally without any removal step
