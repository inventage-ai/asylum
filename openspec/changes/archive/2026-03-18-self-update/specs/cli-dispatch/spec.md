## MODIFIED Requirements

### Requirement: Command dispatch
The CLI SHALL dispatch to version display, agent mode (default), shell mode, ssh-init, cleanup, self-update, or arbitrary command based on flags and positional args.

#### Scenario: Version display
- **WHEN** `--version` flag is set
- **THEN** the version string is printed and the process exits before any container setup

#### Scenario: Default invocation
- **WHEN** `asylum` is run with no positional args
- **THEN** the selected agent starts in YOLO mode

#### Scenario: Shell mode
- **WHEN** `asylum shell` is run
- **THEN** an interactive zsh shell starts

#### Scenario: Arbitrary command
- **WHEN** `asylum run ls -la` is run
- **THEN** `ls -la` runs in the container

#### Scenario: Self-update
- **WHEN** `asylum self-update` is run
- **THEN** the self-update logic executes and the process exits before any container setup

#### Scenario: Self-update with dev flag
- **WHEN** `asylum self-update --dev` is run
- **THEN** the self-update targets the dev channel
