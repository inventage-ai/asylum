## ADDED Requirements

### Requirement: Flag parsing
The CLI SHALL parse `-a/--agent`, `-p`, `-v`, `-e`, `--java`, `-n/--new`, `--skip-onboarding`, `--rebuild`, and `-h/--help` flags. The `--version` and `--cleanup` flags SHALL be accepted as aliases for the `version` and `cleanup` commands respectively. Unknown flags SHALL produce an error.

#### Scenario: Known flags consumed
- **WHEN** `asylum -a gemini -p 3000` is run
- **THEN** agent is set to gemini, port 3000 is forwarded, no passthrough args

#### Scenario: Skip onboarding flag
- **WHEN** `asylum --skip-onboarding` is run
- **THEN** the onboarding system is not invoked for this session

### Requirement: Command dispatch
The CLI SHALL dispatch to version, cleanup, agent mode (default), shell mode, ssh-init, self-update, or arbitrary command based on subcommands and flags. The CLI SHALL accept `selfupdate` as an alias for `self-update`, and `--version`/`--cleanup` as flag aliases for the `version`/`cleanup` commands.

#### Scenario: Version command
- **WHEN** `asylum version` is run
- **THEN** the CLI prints `asylum <version>` to stdout and exits with code 0

#### Scenario: Version flag alias
- **WHEN** `asylum --version` is run
- **THEN** behavior is identical to `asylum version`

#### Scenario: Cleanup command
- **WHEN** `asylum cleanup` is run
- **THEN** the cleanup logic executes and the process exits before any container setup

#### Scenario: Cleanup flag alias
- **WHEN** `asylum --cleanup` is run
- **THEN** behavior is identical to `asylum cleanup`

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

#### Scenario: Self-update with version argument
- **WHEN** `asylum self-update 0.4.0` is run
- **THEN** the self-update targets the specified version

#### Scenario: Selfupdate alias
- **WHEN** `asylum selfupdate` is run
- **THEN** the self-update logic executes identically to `asylum self-update`

#### Scenario: Selfupdate alias with arguments
- **WHEN** `asylum selfupdate --dev` is run
- **THEN** the self-update targets the dev channel, same as `asylum self-update --dev`

### Requirement: Process replacement
The CLI SHALL use `syscall.Exec` to replace itself with the docker process.

#### Scenario: Exec into docker
- **WHEN** the docker run args are assembled
- **THEN** `syscall.Exec` is called with the docker binary path and args
