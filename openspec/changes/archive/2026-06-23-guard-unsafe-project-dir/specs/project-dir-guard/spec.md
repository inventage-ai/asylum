## ADDED Requirements

### Requirement: Unsafe project directory detection

On the container-run path, asylum SHALL classify the resolved project directory as unsafe to sandbox when it is exactly the user's home directory or the filesystem root `/`. All other directories SHALL be treated as safe and used unchanged.

#### Scenario: Home directory is unsafe
- **WHEN** asylum is launched and the resolved project directory equals the user's home directory
- **THEN** the directory is classified as unsafe and the redirect to a fresh workspace is triggered

#### Scenario: Filesystem root is unsafe
- **WHEN** asylum is launched and the resolved project directory equals the filesystem root `/`
- **THEN** the directory is classified as unsafe and the redirect to a fresh workspace is triggered

#### Scenario: A normal project directory is safe
- **WHEN** asylum is launched in a directory that is neither the home directory nor the filesystem root
- **THEN** the directory is used as-is and no workspace is created

#### Scenario: A subdirectory of home is safe
- **WHEN** asylum is launched in a subdirectory of the home directory (e.g. `~/projects/foo`)
- **THEN** the directory is used as-is and no workspace is created

### Requirement: Redirect to a fresh dated workspace

When the project directory is unsafe, asylum SHALL create a new directory under `~/asylum-workspace/<YYYY-MM-DD>-<three-random-words>/`, then use that directory as the project directory for the remainder of the run. The directory SHALL be created empty (no repository initialization). A new workspace SHALL be created on every unsafe launch, with no reuse of previously created workspaces and no cross-launch session resume.

#### Scenario: Workspace is created and used
- **WHEN** the project directory is unsafe
- **THEN** asylum creates `~/asylum-workspace/<today>-<three-random-words>/` and uses it as the project directory for container assembly and the working directory

#### Scenario: Workspace is left empty
- **WHEN** a workspace directory is created
- **THEN** it contains no initialized git repository or seeded files

#### Scenario: Every unsafe launch is fresh
- **WHEN** asylum is launched from an unsafe directory more than once
- **THEN** each launch creates a distinct new workspace directory and does not reuse a prior one

#### Scenario: Name collision is avoided
- **WHEN** a generated workspace path already exists
- **THEN** asylum generates a different name so an existing directory is never reused

### Requirement: Workspace name generation

The three words in a workspace name SHALL be drawn from a wordlist embedded in the binary via `go:embed`, and the date prefix SHALL be the current date formatted as `YYYY-MM-DD`.

#### Scenario: Name format
- **WHEN** a workspace name is generated
- **THEN** it has the form `<YYYY-MM-DD>-<word>-<word>-<word>` using words from the embedded wordlist

### Requirement: Announce the redirect

When asylum redirects to a workspace, it SHALL print a clearly visible warning that names the created workspace path, so the user knows their work is located there rather than in the original directory.

#### Scenario: Redirect is announced
- **WHEN** asylum redirects an unsafe directory to a workspace
- **THEN** it emits a warning line that includes the absolute path of the created workspace

### Requirement: Guard is scoped to the run path

The unsafe-directory guard SHALL run only on the container-run path. Other subcommands, in particular `cleanup`, SHALL NOT create a workspace or redirect the project directory.

#### Scenario: Cleanup does not redirect
- **WHEN** `asylum cleanup` is run from the home directory or filesystem root
- **THEN** no workspace is created and the project directory is not redirected
