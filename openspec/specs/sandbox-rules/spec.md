### Requirement: Kit rules snippet field
The `Kit` struct SHALL have a `RulesSnippet string` field. Kits MAY populate it with a markdown fragment describing the tools and capabilities they provide to the sandbox.

#### Scenario: Kit with rules snippet
- **WHEN** a kit has a non-empty `RulesSnippet`
- **THEN** that snippet SHALL be included in the assembled rules output for containers where the kit is active

#### Scenario: Kit without rules snippet
- **WHEN** a kit has an empty `RulesSnippet`
- **THEN** the kit SHALL be silently skipped during rules assembly (no blank section emitted)

### Requirement: Rules snippet assembly
The kit package SHALL provide an `AssembleRulesSnippets` function that concatenates `RulesSnippet` values from a slice of kits, following the same pattern as `AssembleDockerSnippets` and `AssembleBannerLines`.

#### Scenario: Multiple kits with snippets
- **WHEN** three kits have non-empty rules snippets
- **THEN** the assembled output SHALL contain all three snippets separated by newlines, in the order the kits appear in the input slice

#### Scenario: No kits have snippets
- **WHEN** no kits have rules snippets
- **THEN** the assembled output SHALL be an empty string

### Requirement: Sandbox rules file generation
The system SHALL generate a markdown rules file at `~/.asylum/projects/<container-name>/sandbox-rules.md` each time a new container is started. The file SHALL contain a core section followed by assembled kit rules snippets.

#### Scenario: Container start with active kits
- **WHEN** a container is started with java and node kits active
- **THEN** the rules file SHALL contain the core sandbox context section AND the rules snippets from both kits

#### Scenario: Container start with no kits
- **WHEN** a container is started with an empty kit list
- **THEN** the rules file SHALL contain only the core sandbox context section

#### Scenario: Container restart with changed kits
- **WHEN** a container's config changes to add a new kit and the container is restarted
- **THEN** the regenerated rules file SHALL reflect the updated kit set

### Requirement: Kit tools field
The `Kit` struct SHALL have a `Tools []string` field. Kits that make commands available (but don't need prose to explain) SHALL populate it with command names. The kit package SHALL provide an `AggregateTools` function that collects tools from all kits into a deduplicated list of `"tool (kit-name)"` strings.

#### Scenario: Kit with tools
- **WHEN** kits have non-empty `Tools` fields
- **THEN** the rules file SHALL include a "Kit Tools" section listing all tools with their kit names

#### Scenario: No kit tools
- **WHEN** no kits have `Tools` entries
- **THEN** the rules file SHALL NOT include a "Kit Tools" section

### Requirement: Core rules content
The core section of the rules file SHALL describe: the sandbox identity (Asylum Docker container), the Asylum version, the container user (`claude`), host connectivity (`host.docker.internal`), and the base tools always available from the core Dockerfile (git, Docker CLI, curl, wget, jq, yq, ripgrep, fd, make, cmake, gcc, vim, nano, htop, zip/unzip, ssh). Kit-installed tools (gh, tmux, mvn, etc.) appear in the separate "Kit Tools" section.

#### Scenario: Core content present
- **WHEN** a rules file is generated
- **THEN** it SHALL mention that the environment is an Asylum Docker sandbox, the Asylum version, the user is `claude` with passwordless sudo, and `host.docker.internal` resolves to the host machine

### Requirement: Version included in rules file
`RunOpts` SHALL accept a `Version string` field. The rules file header SHALL include the Asylum version.

#### Scenario: Version displayed
- **WHEN** a rules file is generated with version "1.2.3"
- **THEN** the file header SHALL contain "v1.2.3"

### Requirement: Reference document
The system SHALL embed a detailed Asylum reference document (`assets/asylum-reference.md`) via `go:embed`. On container start, the reference doc SHALL be written alongside the rules file and mounted read-only at `<project-dir>/.claude/asylum-reference.md`. The rules file SHALL reference it for troubleshooting and config details. The reference doc SHALL include a link to the changelog on GitHub.

#### Scenario: Reference doc accessible but not auto-loaded
- **WHEN** a container is running
- **THEN** the reference doc SHALL be readable at `.claude/asylum-reference.md` but SHALL NOT be in `.claude/rules/` (so it is not auto-loaded by Claude Code)

#### Scenario: Reference doc content
- **WHEN** the reference doc is read
- **THEN** it SHALL describe the container lifecycle, layered config system, available kits, volume mounting, self-update mechanism, and troubleshooting steps

### Requirement: Rules file mounted into container
The generated rules file SHALL be mounted read-only into the container at `<project-dir>/.claude/rules/asylum-sandbox.md`. The reference doc SHALL be mounted at `<project-dir>/.claude/asylum-reference.md`. The mounts SHALL NOT create any files on the host filesystem inside the project directory.

#### Scenario: Mount does not pollute host
- **WHEN** a container is running with the rules file mounted
- **THEN** the host project directory SHALL NOT contain a `.claude/rules/asylum-sandbox.md` file

#### Scenario: User's existing rules preserved
- **WHEN** the project already has `.claude/rules/my-rules.md` on the host
- **THEN** that file SHALL remain accessible inside the container alongside the injected `asylum-sandbox.md`

### Requirement: Resolved kits passed to container assembly
`RunOpts` SHALL accept a `Kits []*kit.Kit` field so that the container assembly has access to the resolved kit list for rules generation.

#### Scenario: Kits available in RunOpts
- **WHEN** `RunArgs` is called with a populated `Kits` field
- **THEN** the rules file SHALL be generated using those kits' rules snippets

### Requirement: Kit-provided rules snippets
Each kit that installs tools or provides capabilities MUST populate `RulesSnippet` with a concise markdown description. At minimum, the docker, java, python, and node kits SHALL have rules snippets.

#### Scenario: Docker kit snippet
- **WHEN** the docker kit is active
- **THEN** its rules snippet SHALL mention that full Docker engine is available (not just CLI) and the container runs in privileged mode

#### Scenario: Java kit snippet
- **WHEN** the java kit is active
- **THEN** its rules snippet SHALL mention the available JDK versions and that `mise` manages version switching

#### Scenario: Node kit snippet
- **WHEN** the node kit is active
- **THEN** its rules snippet SHALL mention the globally installed packages (typescript, eslint, prettier, etc.) and that `fnm` manages Node versions

#### Scenario: Python kit snippet
- **WHEN** the python kit is active
- **THEN** its rules snippet SHALL mention the available tools (black, ruff, mypy, pytest, poetry, etc.) and that `uv` manages packages
