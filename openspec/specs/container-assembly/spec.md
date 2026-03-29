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
The container SHALL include all common mounts: project dir at real path, gitconfig, ssh, caches (as named Docker volumes), history, custom volumes, and direnv.

#### Scenario: All common mounts present
- **WHEN** gitconfig exists, ssh dir exists, and project has .envrc
- **THEN** all conditional mounts are included in the args

#### Scenario: Missing optional paths
- **WHEN** gitconfig and ssh dir do not exist
- **THEN** those mounts are omitted, all others remain

#### Scenario: Cache directories use named volumes
- **WHEN** the container is started
- **THEN** cache directories (npm, pip, maven, gradle) are mounted as named Docker volumes using `--mount type=volume,src=<container-name>-cache-<tool>,dst=<path>`

#### Scenario: No host cache directory created
- **WHEN** the container is started
- **THEN** no `~/.asylum/cache/` directory is created on the host

### Requirement: Agent-specific mounts and env vars
The container SHALL mount the agent's asylum config dir and set agent-specific env vars.

#### Scenario: Claude agent
- **WHEN** agent is claude
- **THEN** `~/.asylum/agents/claude/` is mounted at `/home/claude/.claude` and `CLAUDE_CONFIG_DIR` is set

### Requirement: Port forwarding
Ports from config SHALL be mapped in docker run args.

#### Scenario: Simple port
- **WHEN** port is `3000`
- **THEN** `-p 3000:3000` is added to args

#### Scenario: Mapped port
- **WHEN** port is `8080:80`
- **THEN** `-p 8080:80` is added to args

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
