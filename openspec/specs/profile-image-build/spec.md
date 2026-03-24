## ADDED Requirements

### Requirement: Dockerfile decomposition
The monolithic Dockerfile SHALL be split into an embedded core fragment (OS, shell, build tools, Docker, agents, language managers) and an embedded tail fragment (oh-my-zsh, shell config, entrypoint COPY, final USER/WORKDIR).

#### Scenario: Core fragment content
- **WHEN** the core Dockerfile is examined
- **THEN** it contains OS package installation, Docker installation, GitHub/GitLab CLIs, user creation, mise/fnm/uv manager installation, and agent CLI installation — but no language-specific installations

#### Scenario: Tail fragment content
- **WHEN** the tail Dockerfile is examined
- **THEN** it contains oh-my-zsh setup, shell configuration, git config, tmux config, entrypoint COPY, and final USER/WORKDIR directives

### Requirement: Base image assembly from global profiles
The base image Dockerfile SHALL be assembled by concatenating: core fragment + DockerSnippets from globally-active profiles (in deterministic order) + tail fragment.

#### Scenario: All profiles active (default)
- **WHEN** no profiles are specified (default: all active)
- **THEN** the assembled Dockerfile includes Docker snippets for java, python, and node profiles and their sub-profiles

#### Scenario: Subset of profiles active globally
- **WHEN** global config specifies `profiles: [java]`
- **THEN** the assembled base Dockerfile includes only java, maven, and gradle Docker snippets

### Requirement: Project image assembly from project profiles
The project image Dockerfile SHALL start FROM the base image and append: DockerSnippets from project-level profiles + user `packages` instructions.

#### Scenario: Project adds a profile not in global
- **WHEN** global config has `profiles: [java]` and project config has `profiles: [python]`
- **THEN** the project image installs python profile snippets on top of the base image

#### Scenario: Project adds packages alongside profiles
- **WHEN** project config has both `profiles: [node]` and `packages: { apt: [imagemagick] }`
- **THEN** the project image includes both the node profile Docker snippets and the apt package installation

### Requirement: Hash-based cache invalidation includes profiles
The image hash for cache invalidation SHALL include the content of active profile DockerSnippets and EntrypointSnippets in addition to the embedded Dockerfile and entrypoint fragments.

#### Scenario: Profile content changes trigger rebuild
- **WHEN** a profile's DockerSnippet is modified (new binary version)
- **THEN** the base image hash changes and a rebuild is triggered

#### Scenario: Profile selection changes trigger rebuild
- **WHEN** the set of active global profiles changes from [java, python, node] to [java]
- **THEN** the base image hash changes and a rebuild is triggered
