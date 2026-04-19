## ADDED Requirements

### Requirement: Agent interface
The agent package SHALL define an Agent interface with methods: Name, Binary, NativeConfigDir, ContainerConfigDir, AsylumConfigDir, EnvVars, HasSession, Command — matching PLAN.md section 8. The `Command` method SHALL accept an options parameter that carries context from the container layer (e.g. the shared kit-skills directory path), so agent implementations can tailor their launch command accordingly.

#### Scenario: Interface methods
- **WHEN** an Agent implementation is used
- **THEN** all interface methods return correct values per the agent profile table in PLAN.md section 3.1

#### Scenario: Command receives kit-skills context
- **WHEN** `Command` is called with an options value whose `KitSkillsDir` field is `/opt/asylum-skills`
- **THEN** agent implementations that care about skills (Claude) act on it, and agents that do not (Gemini, Codex, Opencode, Echo) ignore it

### Requirement: Agent registry
The agent package SHALL provide a Get function that returns an Agent by name ("claude", "gemini", "codex") or an error for unknown names.

#### Scenario: Valid agent lookup
- **WHEN** `Get("gemini")` is called
- **THEN** it returns the Gemini agent implementation with no error

#### Scenario: Invalid agent lookup
- **WHEN** `Get("unknown")` is called
- **THEN** it returns an error

### Requirement: Claude command generation
Claude SHALL generate commands per PLAN.md section 3.2 with `--dangerously-skip-permissions` and conditional `--continue`. When the container layer signals that at least one active kit declares `ProvidesSkills: true`, the generated command SHALL also include `--add-dir /opt/asylum-skills`.

#### Scenario: Default with session
- **WHEN** Command is called with resume=true, no extra args, and no skill kits active
- **THEN** command includes `--dangerously-skip-permissions --continue` without `--add-dir`

#### Scenario: Default without session
- **WHEN** Command is called with resume=false, no extra args, and no skill kits active
- **THEN** command includes `--dangerously-skip-permissions` without `--continue` or `--add-dir`

#### Scenario: With extra args and resume
- **WHEN** Command is called with resume=true, extra args `["fix the bug"]`, and no skill kits active
- **THEN** command includes `--dangerously-skip-permissions --continue fix the bug` without `--add-dir`

#### Scenario: With skill-providing kit active
- **WHEN** Command is called with resume=false, no extra args, and the container layer signals a skill kit is active
- **THEN** command includes `--dangerously-skip-permissions --add-dir /opt/asylum-skills`

#### Scenario: With skill-providing kit active and resume
- **WHEN** Command is called with resume=true, no extra args, and the container layer signals a skill kit is active
- **THEN** command includes `--dangerously-skip-permissions --continue --add-dir /opt/asylum-skills`

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
