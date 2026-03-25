## MODIFIED Requirements

### Requirement: Dockerfile decomposition
The monolithic Dockerfile SHALL be split into an embedded core fragment (OS, shell, build tools, Docker, language managers) and an embedded tail fragment (oh-my-zsh, shell config, entrypoint COPY, final USER/WORKDIR). Agent CLI installations are NOT part of the core fragment — they are assembled from agent install snippets.

#### Scenario: Core fragment content
- **WHEN** the core Dockerfile is examined
- **THEN** it contains OS package installation, Docker installation, GitHub/GitLab CLIs, user creation, mise/fnm/uv manager installation — but no agent CLI installations and no language-specific installations

#### Scenario: Tail fragment content
- **WHEN** the tail Dockerfile is examined
- **THEN** it contains oh-my-zsh setup, shell configuration, git config, tmux config, entrypoint COPY, and final USER/WORKDIR directives

### Requirement: Base image assembly from global profiles
The base image Dockerfile SHALL be assembled by concatenating: core fragment + DockerSnippets from globally-active profiles (in deterministic order) + DockerSnippets from active agent installs (in deterministic order) + tail fragment.

#### Scenario: Default config (no profiles or agents specified)
- **WHEN** no profiles or agents are specified
- **THEN** the assembled Dockerfile includes profile snippets for java, python, and node, followed by the claude agent install snippet only, followed by the tail

#### Scenario: Subset of agents active
- **WHEN** config specifies `agents: [claude]`
- **THEN** the assembled base Dockerfile includes only the Claude install snippet, not Gemini or Codex

### Requirement: Hash-based cache invalidation includes profiles
The image hash for cache invalidation SHALL include the content of active profile DockerSnippets, EntrypointSnippets, and active agent install DockerSnippets in addition to the embedded Dockerfile and entrypoint fragments.

#### Scenario: Agent selection changes trigger rebuild
- **WHEN** the set of active agents changes from [claude, gemini, codex] to [claude]
- **THEN** the base image hash changes and a rebuild is triggered
