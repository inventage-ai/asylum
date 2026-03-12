## ADDED Requirements

### Requirement: Agent interface
The agent package SHALL define an Agent interface with methods: Name, Binary, NativeConfigDir, ContainerConfigDir, AsylumConfigDir, EnvVars, HasSession, Command — matching PLAN.md section 8.

#### Scenario: Interface methods
- **WHEN** an Agent implementation is used
- **THEN** all interface methods return correct values per the agent profile table in PLAN.md section 3.1

### Requirement: Agent registry
The agent package SHALL provide a Get function that returns an Agent by name ("claude", "gemini", "codex") or an error for unknown names.

#### Scenario: Valid agent lookup
- **WHEN** `Get("gemini")` is called
- **THEN** it returns the Gemini agent implementation with no error

#### Scenario: Invalid agent lookup
- **WHEN** `Get("unknown")` is called
- **THEN** it returns an error

### Requirement: Claude command generation
Claude SHALL generate commands per PLAN.md section 3.2 with `--dangerously-skip-permissions` and conditional `--continue`.

#### Scenario: Default with session
- **WHEN** Command is called with resume=true and no extra args
- **THEN** command includes `--dangerously-skip-permissions --continue`

#### Scenario: Default without session
- **WHEN** Command is called with resume=false and no extra args
- **THEN** command includes `--dangerously-skip-permissions` without `--continue`

#### Scenario: With extra args and resume
- **WHEN** Command is called with resume=true and extra args `["fix the bug"]`
- **THEN** command includes `--dangerously-skip-permissions --continue fix the bug`

### Requirement: Gemini command generation
Gemini SHALL generate commands per PLAN.md section 3.2 with `--yolo` and conditional `--resume`.

#### Scenario: Default with session
- **WHEN** Command is called with resume=true and no extra args
- **THEN** command includes `--yolo --resume`

#### Scenario: Default without session
- **WHEN** Command is called with resume=false and no extra args
- **THEN** command includes `--yolo` without `--resume`

### Requirement: Codex command generation
Codex SHALL generate commands per PLAN.md section 3.2. Resume uses `resume --last` subcommand. Extra args skip resume.

#### Scenario: Default with session
- **WHEN** Command is called with resume=true and no extra args
- **THEN** command uses `codex resume --last --yolo`

#### Scenario: Default without session
- **WHEN** Command is called with resume=false and no extra args
- **THEN** command uses `codex --yolo`

#### Scenario: With extra args (resume ignored)
- **WHEN** Command is called with resume=true and extra args
- **THEN** command uses `codex --yolo <args>` without resume

### Requirement: Session detection
Each agent SHALL detect existing sessions by checking agent-specific filesystem markers per PLAN.md section 3.2.

#### Scenario: Claude session exists
- **WHEN** `~/.asylum/agents/claude/projects/` contains subdirectories
- **THEN** HasSession returns true

#### Scenario: No Claude session
- **WHEN** `~/.asylum/agents/claude/projects/` is empty or missing
- **THEN** HasSession returns false

#### Scenario: Codex session exists
- **WHEN** `~/.asylum/agents/codex/history.jsonl` exists and is non-empty
- **THEN** HasSession returns true
