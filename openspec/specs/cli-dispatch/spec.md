## ADDED Requirements

### Requirement: Flag parsing
The CLI SHALL parse `-a/--agent`, `-p`, `-v`, `-e`, `--java`, `-n/--new`, `--skip-onboarding`, `--cleanup`, `--rebuild`, `--version`, and `-h/--help` flags. Unknown flags SHALL be passed through to the agent.

#### Scenario: Known flags consumed
- **WHEN** `asylum -a gemini -p 3000` is run
- **THEN** agent is set to gemini, port 3000 is forwarded, no passthrough args

#### Scenario: Unknown flags passed through
- **WHEN** `asylum -a gemini -p "fix the bug"` is run
- **THEN** `-p "fix the bug"` is passed to the agent as extra args

#### Scenario: Version flag
- **WHEN** `asylum --version` is run
- **THEN** the CLI prints `asylum <version>` to stdout and exits with code 0

#### Scenario: Skip onboarding flag
- **WHEN** `asylum --skip-onboarding` is run
- **THEN** the onboarding system is not invoked for this session

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

### Requirement: Process replacement
The CLI SHALL use `syscall.Exec` to replace itself with the docker process.

#### Scenario: Exec into docker
- **WHEN** the docker run args are assembled
- **THEN** `syscall.Exec` is called with the docker binary path and args
