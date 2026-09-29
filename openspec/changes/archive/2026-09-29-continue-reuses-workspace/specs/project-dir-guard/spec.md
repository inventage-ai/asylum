# Spec Delta

## MODIFIED Requirements

### Requirement: Redirect to a fresh dated workspace

When the project directory is unsafe and no resume flag is given, asylum SHALL create a new directory under `~/asylum-workspace/<YYYY-MM-DD>-<three-random-words>/`, then use that directory as the project directory for the remainder of the run. The directory SHALL be created empty (no repository initialization). A new workspace SHALL be created on every unsafe launch without a resume flag.

#### Scenario: Workspace is created and used
- **WHEN** the project directory is unsafe and no resume flag is given
- **THEN** asylum creates `~/asylum-workspace/<today>-<three-random-words>/` and uses it as the project directory for container assembly and the working directory

#### Scenario: Workspace is left empty
- **WHEN** a workspace directory is created
- **THEN** it contains no initialized git repository or seeded files

#### Scenario: Every unsafe launch is fresh
- **WHEN** asylum is launched from an unsafe directory more than once without `--continue` or `--resume`
- **THEN** each launch creates a distinct new workspace directory and does not reuse a prior one

#### Scenario: Name collision is avoided
- **WHEN** a generated workspace path already exists
- **THEN** asylum generates a different name so an existing directory is never reused for a fresh workspace

### Requirement: Announce the redirect

When asylum redirects to a workspace, it SHALL print a clearly visible warning that names the workspace path, so the user knows their work is located there rather than in the original directory. The warning SHALL say whether the workspace is fresh or reused.

#### Scenario: Redirect is announced
- **WHEN** asylum redirects an unsafe directory to a fresh workspace
- **THEN** it emits a warning line that includes the absolute path of the created workspace

#### Scenario: Reuse is announced
- **WHEN** asylum redirects an unsafe directory to an existing workspace because of a resume flag
- **THEN** it emits a warning that says the workspace is being continued and includes its absolute path

## ADDED Requirements

### Requirement: Resume flags reuse the newest workspace

When the project directory is unsafe and the arguments contain `--continue` or `--resume`, asylum SHALL use the newest existing workspace under `~/asylum-workspace/` as the project directory instead of creating one. The newest workspace is the directory with the most recent modification time among those whose names match `<YYYY-MM-DD>-<word>-<word>-<word>`. Other entries in `~/asylum-workspace/` SHALL be ignored. When no matching workspace exists, asylum SHALL create a fresh workspace as it does without a resume flag.

#### Scenario: Continue reuses the newest workspace
- **WHEN** `~/asylum-workspace/` holds workspaces A and B, B was modified more recently, and `asylum --continue` is run from the home directory
- **THEN** B is used as the project directory and no new workspace is created

#### Scenario: Resume behaves the same
- **WHEN** `asylum --resume` is run from the home directory and a workspace exists
- **THEN** the newest workspace is used as the project directory

#### Scenario: Unrelated directories are ignored
- **WHEN** `~/asylum-workspace/` contains a directory named `notes` that was modified after every generated workspace
- **THEN** `notes` is not chosen, and the newest generated workspace is used

#### Scenario: No workspace exists yet
- **WHEN** `asylum --continue` is run from the home directory and `~/asylum-workspace/` has no matching workspace or does not exist
- **THEN** a fresh workspace is created as without the flag

#### Scenario: Safe directories are unaffected
- **WHEN** `asylum --continue` is run in a safe project directory
- **THEN** that directory is used unchanged
