## ADDED Requirements

### Requirement: One-time upgrade dialog for existing users
On the first invocation of `asylum` after upgrading to the version that ships the default-new-session behaviour change, asylum SHALL present a one-time interactive TUI dialog informing the user that the default session behaviour has changed and offering to restore the previous auto-resume behaviour via a config flag. The dialog SHALL be shown at most once per user.

#### Scenario: Existing installation, dialog not yet shown
- **WHEN** `~/.asylum/` exists from a prior asylum run AND `state.json` does not record that the resume-migration prompt has been shown AND the user runs `asylum` (agent mode)
- **THEN** asylum displays a TUI dialog explaining the change before launching the agent
- **AND** asylum records the prompt as shown in `state.json` regardless of which option the user picks

#### Scenario: Dialog already shown
- **WHEN** `state.json` records that the resume-migration prompt has been shown
- **THEN** asylum SHALL NOT display the dialog again, on any subsequent invocation

#### Scenario: User picks "keep new default"
- **WHEN** the dialog is shown and the user selects the option to keep the new default
- **THEN** no config file is written, the prompt-shown flag is set, and asylum proceeds with the new-session default

#### Scenario: User picks "restore previous behaviour"
- **WHEN** the dialog is shown and the user selects the option to restore auto-resume
- **THEN** `~/.asylum/config.yaml` is updated to set `default-resume: true` (creating the file if absent, preserving other keys if present)
- **AND** the prompt-shown flag is set in `state.json`
- **AND** asylum proceeds with the current invocation honouring the new `default-resume: true` value

#### Scenario: Opt-in write failure re-prompts on next run
- **WHEN** the user selects "restore previous behaviour" but writing `default-resume: true` to `~/.asylum/config.yaml` fails (read-only filesystem, full disk, unparseable existing file, etc.)
- **THEN** the prompt-shown flag SHALL NOT be set in `state.json`
- **AND** asylum SHALL log the error to the user
- **AND** asylum SHALL continue the current invocation with the new-session default (the user's preference is lost for this run but will be asked again next interactive run, rather than silently committing them to the new default)

### Requirement: New installations skip the dialog
A new asylum installation (no pre-existing `~/.asylum/` directory before the current run) SHALL NOT be shown the migration dialog. New users see the new default behaviour silently.

#### Scenario: First-ever asylum run
- **WHEN** `~/.asylum/` does not exist at the start of the asylum invocation
- **THEN** asylum SHALL NOT show the migration dialog
- **AND** asylum SHALL mark the prompt as already shown in the newly created `state.json` so it is never shown later

#### Scenario: Detection is unaffected by initialisation order
- **WHEN** asylum determines whether the user is new vs. existing
- **THEN** the detection SHALL probe a location that is not created by the eager default-config write, the kit-sync step, or any other initialisation that runs on every invocation — concretely, `<home>/.asylum/agents/`, which is created only when an agent's config is first materialised (`container.EnsureAgentConfig`)

### Requirement: Dialog is skipped in non-interactive contexts
The migration dialog SHALL NOT be shown when stdin/stdout is not a terminal (e.g. piped invocation, CI, or `asylum run`/`asylum cleanup`/`asylum version`/`asylum self-update` subcommands). In that case the prompt-shown flag SHALL NOT be set, so the next interactive invocation will surface the dialog.

#### Scenario: Non-interactive shell
- **WHEN** asylum is invoked in a context without a TTY
- **THEN** the dialog is suppressed and the prompt-shown flag is left untouched

#### Scenario: Non-agent subcommands
- **WHEN** the user runs `asylum version`, `asylum cleanup`, `asylum config`, `asylum self-update`, `asylum shell`, or `asylum run`
- **THEN** the dialog is suppressed and the prompt-shown flag is left untouched

### Requirement: Prompt-shown state persistence
The asylum `state.json` SHALL include a boolean field recording that the resume-migration prompt has been shown for this installation.

#### Scenario: Field round-trips
- **WHEN** asylum sets the prompt-shown flag and writes `state.json`
- **THEN** subsequent loads of `state.json` return the field as `true`

#### Scenario: Default value
- **WHEN** `state.json` is missing or omits the field
- **THEN** the loaded value is `false`
