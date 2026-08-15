# kit-activation-prompt Specification

## Purpose

Tells an existing installation about kits added in a newer asylum release, instead of silently enabling them or leaving them undiscovered. Default and opt-in kits are offered in an interactive multiselect; always-on kits are announced but not offered, since declining them is not an option; and utility subcommands that never build a container skip the prompt entirely.

## Requirements

### Requirement: Interactive prompt for new kits
When new `TierDefault` or `TierOptIn` kits are detected in an interactive terminal session, the system SHALL present a single `tui.MultiSelect` prompt listing all new promptable kits. `TierDefault` kits SHALL be pre-selected; `TierOptIn` kits SHALL be unselected. The user can toggle any kit. Selected kits are activated in `config.yaml`; deselected kits are added as comments.

#### Scenario: Multiple new kits presented in one prompt
- **WHEN** kits `rust` (TierDefault) and `apt` (TierOptIn) are both newly detected
- **THEN** the system displays a single multiselect prompt containing both kits, each with its description, with `rust` pre-selected and `apt` unselected

#### Scenario: User accepts all defaults
- **WHEN** the multiselect prompt shows `rust` pre-selected and `apt` unselected and the user presses Enter without changes
- **THEN** `rust` is inserted as an active entry in `config.yaml`
- **AND** `apt` is inserted as a commented-out entry in `config.yaml`

#### Scenario: User selects an opt-in kit
- **WHEN** the multiselect prompt shows `apt` (TierOptIn) and the user selects it
- **THEN** `apt` is inserted as an active entry in `config.yaml`

#### Scenario: User deselects a default kit
- **WHEN** the multiselect prompt shows `rust` (TierDefault) pre-selected and the user deselects it
- **THEN** `rust` is inserted as a commented-out entry in `config.yaml`

#### Scenario: User cancels the prompt
- **WHEN** the user presses Escape or Ctrl+C during the multiselect prompt
- **THEN** all new kits are inserted as commented-out entries in `config.yaml` (same as declining all)

#### Scenario: Kit descriptions shown
- **WHEN** the multiselect prompt is displayed
- **THEN** each kit option shows the kit name as the label and the kit's `Description` field as the description text

#### Scenario: Non-interactive mode
- **WHEN** new kits are detected and the session is non-interactive (stdin is not a TTY)
- **THEN** all new `TierDefault` kits are activated without prompting and `TierOptIn` kits are added as comments

### Requirement: Info message for always-on kits
When new `TierAlwaysOn` kits are detected, the system SHALL display an informational message but not prompt, since these kits activate regardless of config.

#### Scenario: Always-on kit announced
- **WHEN** a new `TierAlwaysOn` kit `shell` is detected
- **THEN** an info message is displayed (e.g., "New kit: shell (always active)")
- **AND** no prompt is shown and no config modification is made

### Requirement: Opt-in kits included in prompt
When new `TierOptIn` kits are detected in an interactive session, they SHALL be included in the multiselect prompt (unselected by default) instead of showing a standalone info message.

#### Scenario: Opt-in kit in multiselect
- **WHEN** a new `TierOptIn` kit `apt` is detected in an interactive session
- **THEN** `apt` appears in the multiselect prompt with its description, unselected by default

#### Scenario: Opt-in kit non-interactive
- **WHEN** a new `TierOptIn` kit `apt` is detected in a non-interactive session
- **THEN** `apt` is added as a commented-out entry in the config with an info message

### Requirement: Skip prompts for utility subcommands
Kit sync prompts SHALL NOT run for utility subcommands that don't involve container setup.

#### Scenario: Self-update skips sync
- **WHEN** asylum is invoked with `self-update`
- **THEN** kit sync does not run

#### Scenario: Version skips sync
- **WHEN** asylum is invoked with `version`
- **THEN** kit sync does not run
