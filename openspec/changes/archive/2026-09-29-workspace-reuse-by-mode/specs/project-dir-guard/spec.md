# Spec Delta

## MODIFIED Requirements

### Requirement: Redirect to a fresh dated workspace

When the project directory is unsafe and no reuse rule selects an existing workspace, asylum SHALL create a new directory under `~/asylum-workspace/<YYYY-MM-DD>-<three-random-words>/`, then use that directory as the project directory for the remainder of the run. The directory SHALL be created empty (no repository initialization). A plain agent launch, without `--continue` or `--resume`, SHALL always create a new workspace.

#### Scenario: Workspace is created and used
- **WHEN** the project directory is unsafe and no reuse rule selects an existing workspace
- **THEN** asylum creates `~/asylum-workspace/<today>-<three-random-words>/` and uses it as the project directory for container assembly and the working directory

#### Scenario: Workspace is left empty
- **WHEN** a workspace directory is created
- **THEN** it contains no initialized git repository or seeded files

#### Scenario: Every unsafe launch is fresh
- **WHEN** `asylum` is launched in agent mode from an unsafe directory more than once without `--continue` or `--resume`, while earlier workspaces exist
- **THEN** each launch creates a distinct new workspace directory and does not reuse a prior one

#### Scenario: Name collision is avoided
- **WHEN** a generated workspace path already exists
- **THEN** asylum generates a different name so an existing directory is never reused for a fresh workspace

### Requirement: Announce the redirect

When asylum redirects to a workspace, it SHALL print a clearly visible warning that names the workspace path, so the user knows their work is located there rather than in the original directory. The warning SHALL say whether the workspace is fresh, continued because of a resume flag, or attached because its container is running.

#### Scenario: Redirect is announced
- **WHEN** asylum redirects an unsafe directory to a fresh workspace
- **THEN** it emits a warning line that includes the absolute path of the created workspace

#### Scenario: Reuse is announced
- **WHEN** asylum redirects an unsafe directory to an existing workspace because of a resume flag
- **THEN** it emits a warning that says the workspace is being continued and includes its absolute path

#### Scenario: Attach is announced
- **WHEN** asylum redirects `shell` or `run` to a workspace whose container is running
- **THEN** it emits a warning that says it is attaching to the running workspace and includes its absolute path

### Requirement: Guard is scoped to the run path

The unsafe-directory guard SHALL run only on the container-run path. Other subcommands, in particular `cleanup`, `self-update` and `update`, SHALL NOT create a workspace or redirect the project directory.

#### Scenario: Cleanup does not redirect
- **WHEN** `asylum cleanup` is run from the home directory or filesystem root
- **THEN** no workspace is created and the project directory is not redirected

#### Scenario: Self-update does not redirect
- **WHEN** `asylum self-update` is run from the home directory
- **THEN** no workspace is created

#### Scenario: Update does not redirect
- **WHEN** `asylum update` is run from the home directory
- **THEN** no workspace is created

### Requirement: Resume flags reuse the newest workspace

When the project directory is unsafe, asylum runs in agent mode, and the arguments contain `--continue` or `--resume`, asylum SHALL use the newest existing workspace in which the active agent has a session as the project directory. The newest workspace is the one with the most recent modification time among directories under `~/asylum-workspace/` whose names match `<YYYY-MM-DD>-<word>-<word>-<word>`. Other entries SHALL be ignored. The active agent and its config isolation SHALL be resolved as for a normal launch. When no matching workspace holds a session for that agent, asylum SHALL create a fresh workspace.

#### Scenario: Continue reuses the newest workspace
- **WHEN** `~/asylum-workspace/` holds workspaces A and B, the agent has a session in both, B was modified more recently, and `asylum --continue` is run from the home directory
- **THEN** B is used as the project directory and no new workspace is created

#### Scenario: Resume behaves the same
- **WHEN** `asylum --resume` is run from the home directory and a workspace with a session for the agent exists
- **THEN** the newest such workspace is used as the project directory

#### Scenario: Unrelated directories are ignored
- **WHEN** `~/asylum-workspace/` contains a directory named `notes` that was modified after every generated workspace
- **THEN** `notes` is not chosen, and the newest generated workspace with a session is used

#### Scenario: Workspace without a session is skipped
- **WHEN** the newest workspace has no session for the agent and an older one does
- **THEN** `asylum --continue` uses the older workspace

#### Scenario: No workspace exists yet
- **WHEN** `asylum --continue` is run from the home directory and no matching workspace holds a session for the agent
- **THEN** a fresh workspace is created

#### Scenario: Safe directories are unaffected
- **WHEN** `asylum --continue` is run in a safe project directory
- **THEN** that directory is used unchanged

## ADDED Requirements

### Requirement: Shell and run attach to a running workspace

When the project directory is unsafe and asylum runs `shell` (including admin shell) or `run`, asylum SHALL use the newest generated workspace whose primary container is running as the project directory. When no such workspace exists, asylum SHALL create a fresh workspace. Resume flags among the arguments SHALL NOT change this rule.

#### Scenario: Shell attaches to the running workspace
- **WHEN** a Claude session is running in workspace A and `asylum shell` is run from the home directory
- **THEN** A is used as the project directory and the shell attaches to A's container

#### Scenario: Run ignores a literal resume flag
- **WHEN** `asylum run git rebase --continue` is run from the home directory and no workspace container is running
- **THEN** a fresh workspace is created, and no workspace is chosen because of the `--continue` argument

#### Scenario: Nothing running
- **WHEN** `asylum shell` is run from the home directory and no workspace container is running
- **THEN** a fresh workspace is created
