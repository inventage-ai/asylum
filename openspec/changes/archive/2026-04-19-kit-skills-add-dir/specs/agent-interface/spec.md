## MODIFIED Requirements

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

### Requirement: Agent interface
The agent package SHALL define an Agent interface with methods: Name, Binary, NativeConfigDir, ContainerConfigDir, AsylumConfigDir, EnvVars, HasSession, Command — matching PLAN.md section 8. The `Command` method SHALL accept an options parameter that carries context from the container layer (e.g. the shared kit-skills directory path), so agent implementations can tailor their launch command accordingly.

#### Scenario: Interface methods
- **WHEN** an Agent implementation is used
- **THEN** all interface methods return correct values per the agent profile table in PLAN.md section 3.1

#### Scenario: Command receives kit-skills context
- **WHEN** `Command` is called with an options value whose `KitSkillsDir` field is `/opt/asylum-skills`
- **THEN** agent implementations that care about skills (Claude) act on it, and agents that do not (Gemini, Codex, Opencode, Echo) ignore it
