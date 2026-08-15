# tui-prompts Specification

## Purpose

Provides the interactive building blocks the wizard and `asylum config` are built from: single-select, multi-select with checkboxes, options carrying a label and optional description, and a wizard that runs several steps in one session. Also requires the tabbed UI to keep selections when the user moves between tabs, so nothing is silently lost before confirming.

## Requirements

### Requirement: Single-choice prompt
The TUI package SHALL provide a `Select` function that displays a list of options and returns the selected index.

#### Scenario: User selects an option
- **WHEN** `Select` is called with 3 options and defaultIdx 1
- **THEN** the prompt displays all options with option 1 pre-selected, and returns the user's final selection

#### Scenario: User cancels
- **WHEN** the user presses Escape or Ctrl+C during a Select prompt
- **THEN** the function returns -1 and an error

#### Scenario: Non-interactive mode
- **WHEN** `Select` is called but stdin is not a TTY
- **THEN** the function returns the default index with no prompt

### Requirement: Multi-choice prompt
The TUI package SHALL provide a `MultiSelect` function that displays a list of options with checkboxes and returns the selected indices.

#### Scenario: User selects multiple options
- **WHEN** `MultiSelect` is called with 4 options and 2 pre-selected
- **THEN** the prompt displays all options with the pre-selected ones checked, and returns the final selection

#### Scenario: User cancels multi-select
- **WHEN** the user presses Escape during a MultiSelect prompt
- **THEN** the function returns nil and an error

### Requirement: Option descriptions
Each option SHALL support a `Label` and an optional `Description` displayed below the label.

#### Scenario: Option with description
- **WHEN** an option has both Label and Description set
- **THEN** the label is displayed prominently and the description in dimmer text below it

### Requirement: Wizard prompt
The TUI package SHALL provide a `Wizard` function that runs multiple Select and MultiSelect steps within a single bubbletea program, with a tab bar showing step progress.

#### Scenario: Multi-step wizard
- **WHEN** `Wizard` is called with 2 steps (a Select and a MultiSelect)
- **THEN** the prompt displays a tab bar with both step titles, renders the first step, and advances to the second step on confirm

#### Scenario: Tab bar shows progress
- **WHEN** the user completes step 1 and is on step 2
- **THEN** step 1 shows a ✓ in the tab bar and step 2 is highlighted as current

#### Scenario: Cancel during wizard
- **WHEN** the user presses Escape during any step
- **THEN** the wizard returns results for all completed steps and marks the current step as cancelled

#### Scenario: Non-interactive wizard
- **WHEN** `Wizard` is called but stdin is not a TTY
- **THEN** the function returns default selections for all steps without prompting

#### Scenario: Single-step wizard
- **WHEN** `Wizard` is called with only 1 step
- **THEN** the wizard renders the tab bar with one tab and behaves identically to running the step alone

### Requirement: Tab state preservation across tab switches
The tabbed TUI SHALL preserve user selections when switching between tabs. When the user navigates away from a tab and later returns, the tab SHALL display the selections as the user left them, not the original defaults.

#### Scenario: Multiselect selections preserved after round-trip
- **WHEN** the user toggles selections on a multiselect tab, switches to another tab, then switches back
- **THEN** the multiselect tab SHALL display the user's modified selections, not the original defaults

#### Scenario: Select cursor preserved after round-trip
- **WHEN** the user moves the cursor on a select tab, switches to another tab, then switches back
- **THEN** the select tab SHALL display the cursor at the user's last position, not the default

#### Scenario: First visit uses defaults
- **WHEN** the user navigates to a tab for the first time
- **THEN** the tab SHALL display the default selections (`DefaultSel` for multiselect, `DefaultIdx` for select)

#### Scenario: Empty selection preserved
- **WHEN** the user deselects all options on a multiselect tab, switches away, then switches back
- **THEN** the tab SHALL show no options selected (not revert to defaults)
