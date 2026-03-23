## MODIFIED Requirements

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
