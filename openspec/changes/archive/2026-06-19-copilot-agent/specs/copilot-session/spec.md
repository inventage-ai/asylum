## ADDED Requirements

### Requirement: Project-scoped copilot session detection
Copilot's `session-state/` directory is global to the config dir, and `copilot --resume` lists every recent session regardless of which project they came from. Asylum SHALL gate the `--resume` flag on an Asylum-owned per-project marker stored at `<configDir>/asylum-projects/<encoded-project-path>/.has_session`. The marker SHALL be written by `WriteMarker` after a successful first launch of copilot in a given project. A project SHALL NOT inherit another project's marker, even when both share the same copilot config dir.

#### Scenario: Fresh project — no resume
- **WHEN** copilot is launched in a project that has no marker yet under `<configDir>/asylum-projects/`
- **THEN** `HasSession` returns false and the launch command does NOT include `--resume`

#### Scenario: Returning project — resume passed
- **WHEN** copilot has previously been launched in this project (its marker exists)
- **THEN** `HasSession` returns true and the launch command includes `--resume`

#### Scenario: Cross-project isolation
- **WHEN** project A has a marker but project B does not, and both share the same copilot config dir
- **THEN** `HasSession` returns true for project A and false for project B

#### Scenario: Marker persists across runs
- **WHEN** `WriteMarker` is called twice for the same project
- **THEN** the second call succeeds and `HasSession` continues to return true
