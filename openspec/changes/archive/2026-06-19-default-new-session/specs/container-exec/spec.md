## MODIFIED Requirements

### Requirement: Exec agent into running container
When a container is already running for the current project and the user runs `asylum` (agent mode), asylum SHALL exec the agent into the running container. By default, asylum SHALL start a new agent session — it SHALL NOT auto-resume from local session markers. Resume happens only when the user explicitly passes `--continue` or `--resume` (which asylum forwards to the agent), or when the resolved config has `default-resume: true`.

#### Scenario: Agent exec with running container
- **WHEN** the user runs `asylum` and a container is running for the project
- **THEN** asylum execs the agent command into the running container via `docker exec -it`

#### Scenario: Default starts a new session
- **WHEN** the user runs `asylum` with no resume-related flags and `default-resume` is unset (or `false`)
- **THEN** the exec'd agent starts a fresh session — no `--continue`/`--resume` is injected, regardless of whether a local session marker exists

#### Scenario: --continue passthrough
- **WHEN** the user runs `asylum --continue`
- **THEN** `--continue` is included verbatim in the agent's argv, and asylum does NOT additionally inject its own resume flag

#### Scenario: --resume passthrough
- **WHEN** the user runs `asylum --resume`
- **THEN** `--resume` is included verbatim in the agent's argv

#### Scenario: default-resume restores previous behaviour
- **WHEN** the resolved config has `default-resume: true` and a local session marker indicates a prior session exists
- **THEN** asylum injects the agent's native resume flag (e.g. `--continue` for Claude, `--resume` for Gemini/Copilot, `resume --last` for Codex) as it did before this change

#### Scenario: default-resume with no prior session
- **WHEN** `default-resume: true` is set but `HasSession` returns false
- **THEN** asylum starts a new session — there is nothing to resume

#### Scenario: -n/--new no-op with default-resume on
- **WHEN** `default-resume: true` is set AND the user runs `asylum -n`
- **THEN** `-n` is ignored (no-op) and asylum still resumes per `default-resume`
