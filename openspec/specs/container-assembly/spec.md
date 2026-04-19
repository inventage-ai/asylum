## ADDED Requirements

### Requirement: Container naming
Container name SHALL be `asylum-<sha256(project_dir)[:12]>-<sanitized_basename>` and hostname SHALL be `asylum-<sanitized_basename>`. On first run, old-format project directories (`asylum-<hash>` without suffix) SHALL be migrated to the new format.

#### Scenario: Naming from project path
- **WHEN** the project directory is `/home/user/code/myapp`
- **THEN** the container name is `asylum-<hash[:12]>-myapp` and hostname is `asylum-myapp`

#### Scenario: Migration of old project directory
- **WHEN** `~/.asylum/projects/asylum-<hash>` exists but `~/.asylum/projects/asylum-<hash>-<project>` does not
- **THEN** the old directory is renamed and port allocations are updated

### Requirement: Common volume mounts
The container SHALL include all common mounts: project dir at real path, gitconfig, caches (as named Docker volumes), history, custom volumes, and direnv. SSH mounts are handled by the SSH kit's credential function. All mounts SHALL be represented as RunArgs with source `core` and priority 0, except user-configured volumes which SHALL have source `user config (volumes)` and priority 2. The docker subcommand (`run`) and mode flag (`-d`) SHALL NOT be emitted as RunArgs; they are prepended during flattening.

#### Scenario: All common mounts present
- **WHEN** gitconfig exists and project has .envrc
- **THEN** all conditional mounts are included as RunArgs with source `core`

#### Scenario: Missing optional paths
- **WHEN** gitconfig does not exist
- **THEN** that mount is omitted, all others remain

#### Scenario: Cache directories use named volumes
- **WHEN** the container is started
- **THEN** cache directories are mounted as named Docker volumes via `--mount` RunArgs with source `core`

#### Scenario: User volume conflicts with core mount
- **WHEN** a user-configured volume mounts to the same container path as a core mount
- **THEN** the user volume (priority 2) SHALL override the core mount (priority 0)

#### Scenario: No run or -d in RunArgs
- **WHEN** `RunArgs()` is called
- **THEN** the returned resolved args SHALL NOT contain entries with Flag `run` or `-d`

### Requirement: Agent-specific mounts and env vars
The container SHALL mount the agent's asylum config dir and set agent-specific env vars. These SHALL be represented as RunArgs with source `core`.

#### Scenario: Claude agent
- **WHEN** agent is claude
- **THEN** RunArgs for the config dir mount and env vars SHALL have source `core`

### Requirement: Kit-contributed environment variables in container
The container assembly SHALL collect environment variables from all active kits that provide an `EnvFunc`. These SHALL be represented as RunArgs with source `kit` and priority 1, and SHALL NOT be hardcoded per-kit in the container assembly code.

#### Scenario: Java kit contributes ASYLUM_JAVA_VERSION
- **WHEN** the java kit is active with `default-version: 21`
- **THEN** the container run args SHALL include `-e ASYLUM_JAVA_VERSION=21` with source `kit`

#### Scenario: Kit returns no env vars
- **WHEN** a kit's `EnvFunc` returns an empty map
- **THEN** no env args SHALL be added for that kit

#### Scenario: No hardcoded kit env vars
- **WHEN** the container is assembled
- **THEN** the container assembly code SHALL NOT contain any kit-specific env var logic (e.g., no `if java` checks)

### Requirement: Port forwarding
Ports from user config and kit-allocated ports SHALL both be represented as RunArgs in the unified pipeline. User-configured ports SHALL have priority 2 (config), kit-allocated ports SHALL have priority 1 (kit). Deduplication on container port SHALL prevent conflicts.

#### Scenario: Simple port from config
- **WHEN** user config has port `3000`
- **THEN** a RunArg with Flag=`-p`, Value=`3000:3000`, Source=`user config (ports)`, Priority=2 SHALL be added

#### Scenario: Mapped port from config
- **WHEN** user config has port `8080:80`
- **THEN** a RunArg with Flag=`-p`, Value=`8080:80`, Source=`user config (ports)`, Priority=2 SHALL be added

#### Scenario: Kit-allocated ports
- **WHEN** the ports kit allocates ports 10000-10004
- **THEN** five RunArgs with Flag=`-p`, Source=`ports kit`, Priority=1 SHALL be added

#### Scenario: Config port overrides kit port on same container port
- **WHEN** user config has port `3000:3000` and the ports kit also allocated port 3000
- **THEN** only the user config mapping SHALL appear in the final args

### Requirement: Shell mode selection
The container command SHALL vary based on mode: agent (default), shell, admin shell, or arbitrary command.

#### Scenario: Agent mode
- **WHEN** mode is agent
- **THEN** the agent's Command() output is used

#### Scenario: Shell mode
- **WHEN** mode is shell
- **THEN** the container command is `/bin/zsh`

#### Scenario: Admin shell mode
- **WHEN** mode is admin shell
- **THEN** the container command includes sudo notice and `/bin/zsh`

#### Scenario: Arbitrary command
- **WHEN** mode is command with args
- **THEN** those args are used as the container command

### Requirement: Mount git worktree directories
When the project directory is a git worktree, the volume assembly SHALL mount both the worktree-specific gitdir and the main repo's `.git` directory into the container.

#### Scenario: Project is a git worktree
- **WHEN** the project directory's `.git` is a file containing `gitdir: /repo/.git/worktrees/feature`
- **THEN** both `/repo/.git/worktrees/feature` and `/repo/.git` are mounted at their real host paths

#### Scenario: Project is a regular repo
- **WHEN** the project directory's `.git` is a directory
- **THEN** no additional git-related volumes are added (`.git` is already inside the mounted project dir)

#### Scenario: Project has no .git
- **WHEN** the project directory has no `.git` file or directory
- **THEN** no additional git-related volumes are added

### Requirement: Agent config seeding
On first run for an agent, the system SHALL copy the agent's native host config to the asylum agents directory.

#### Scenario: First run with existing native config
- **WHEN** `~/.asylum/agents/claude/` does not exist but `~/.claude/` does
- **THEN** contents of `~/.claude/` are copied to `~/.asylum/agents/claude/`

#### Scenario: First run without native config
- **WHEN** neither `~/.asylum/agents/claude/` nor `~/.claude/` exists
- **THEN** `~/.asylum/agents/claude/` is created empty

### Requirement: Shadow node_modules volumes
During volume assembly, the system SHALL detect `node_modules` directories in the project and shadow them with named Docker volumes so host-built native binaries are not visible inside the container.

#### Scenario: Node.js project with node_modules
- **WHEN** the project has a `package.json` and a `node_modules` directory
- **THEN** `--mount type=volume,src=<named-volume>,dst=<node_modules_path>` is added after the project directory mount

#### Scenario: Non-Node.js project
- **WHEN** the project has no `package.json`
- **THEN** no shadow volume mounts are added

#### Scenario: Feature disabled via config
- **WHEN** config has `features: { shadow-node-modules: false }`
- **THEN** no shadow volume mounts are added regardless of project contents
