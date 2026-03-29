## ADDED Requirements

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
