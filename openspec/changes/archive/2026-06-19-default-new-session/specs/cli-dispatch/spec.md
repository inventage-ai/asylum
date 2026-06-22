## MODIFIED Requirements

### Requirement: Flag parsing
The CLI SHALL parse `-a/--agent`, `-p`, `-v`, `-e`, `--java`, `-n/--new`, `--continue`, `--resume`, `--skip-onboarding`, `--rebuild`, and `-h/--help` flags. The `--version` and `--cleanup` flags SHALL be accepted as aliases for the `version` and `cleanup` commands respectively. Unknown flags SHALL produce an error.

`-n/--new` SHALL be recognised but is a no-op: starting a new session is the default. It is retained for backwards compatibility with existing user scripts and aliases. Help text SHALL mark it deprecated.

`--continue` and `--resume` SHALL be recognised and forwarded verbatim to the underlying agent as passthrough args. Asylum SHALL NOT translate them or attempt to derive a session from local markers when these flags are present — resume is the agent's responsibility.

#### Scenario: Known flags consumed
- **WHEN** `asylum -a gemini -p 3000` is run
- **THEN** agent is set to gemini, port 3000 is forwarded, no passthrough args

#### Scenario: Skip onboarding flag
- **WHEN** `asylum --skip-onboarding` is run
- **THEN** the onboarding system is not invoked for this session

#### Scenario: --new is a recognised no-op
- **WHEN** `asylum --new` (or `asylum -n`) is run
- **THEN** parsing succeeds and the flag has no effect on session selection (a new session starts, same as `asylum` with no flags)

#### Scenario: --continue forwarded to agent
- **WHEN** `asylum --continue` is run
- **THEN** `--continue` appears verbatim in the args passed to the underlying agent process

#### Scenario: --resume forwarded to agent
- **WHEN** `asylum --resume` is run
- **THEN** `--resume` appears verbatim in the args passed to the underlying agent process

#### Scenario: Passthrough position preserved
- **WHEN** `asylum --continue "fix bug"` is run
- **THEN** the agent receives `--continue` followed by `"fix bug"` in order, alongside any agent-specific default args (e.g. `--dangerously-skip-permissions` for Claude)
