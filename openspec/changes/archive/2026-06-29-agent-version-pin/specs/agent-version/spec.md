## ADDED Requirements

### Requirement: Version file location and format

Agent versions SHALL be stored in a local JSON file at `~/.asylum/versions.json`. The file SHALL use the format `{"agent_name": "version_string", ...}` where keys are agent names (e.g. `"claude"`, `"gemini"`) and values are version strings as returned by their respective sources (e.g. `"v2.1.195"` from GitHub tags, `"0.8.0"` from npm).

#### Scenario: File is read on startup
- **WHEN** asylum starts and `versions.json` exists
- **THEN** it is loaded as a JSON object mapping agent names to version strings

#### Scenario: File does not exist
- **WHEN** asylum starts and `versions.json` does not exist
- **THEN** an empty version map is returned, triggering a blocking fetch before the image build

#### Scenario: File is corrupted
- **WHEN** `versions.json` exists but contains invalid JSON
- **THEN** the file is treated as missing and a blocking fetch is triggered

### Requirement: Blocking fetch on first run

When the version file does not exist or is corrupted, asylum SHALL perform a blocking fetch of all agent versions before proceeding with the image build. The image build SHALL NOT proceed until at least one valid version has been resolved.

#### Scenario: Blocking fetch succeeds
- **WHEN** the version file is missing and all fetches succeed
- **THEN** the versions are written to `versions.json` and the image build proceeds

#### Scenario: Some fetches fail during blocking
- **WHEN** the version file is missing but some fetches fail
- **THEN** the successfully fetched versions are saved (with the missing ones omitted), and the build proceeds with available versions

#### Scenario: All fetches fail during blocking
- **WHEN** the version file is missing and all fetches fail
- **THEN** the build proceeds with an empty version map (no version pinning), same as current behavior

### Requirement: Background refresh on subsequent runs

When the version file exists and is valid, asylum SHALL load it from disk (instantly) and proceed with the build. A background goroutine SHALL check if the file is older than 24 hours, and if so, fetch all agent versions and update the file.

#### Scenario: Background fetch is skipped
- **WHEN** the version file was updated less than 24 hours ago
- **THEN** the background goroutine does nothing and the build proceeds with cached versions

#### Scenario: Background fetch succeeds
- **WHEN** the version file is older than 24 hours and all fetches succeed
- **THEN** the file is updated with new versions and no error is reported to the user

#### Scenario: Background fetch fails
- **WHEN** the version file is older than 24 hours and fetches fail
- **THEN** the failure is silently ignored and the cached versions remain valid

#### Scenario: Background fetch is fire-and-forget
- **WHEN** a background fetch is in progress
- **THEN** it does not block the current run; the next run picks up any new versions

### Requirement: Per-agent ARG injection

Each agent's install command in the base Dockerfile SHALL be preceded by an `ARG <AGENT>_VERSION=<value>` declaration. The ARGs SHALL be placed immediately before their respective RUN instructions, not at the top of the Dockerfile, so that Docker layer caching is preserved per-agent.

#### Scenario: Claude gets versioned ARG
- **WHEN** Claude is installed and versions.json contains `"claude": "v2.1.195"`
- **THEN** the Dockerfile includes `ARG CLAUDE_VERSION=v2.1.195` immediately before Claude's RUN instruction

#### Scenario: Gemini gets versioned ARG
- **WHEN** Gemini is installed and versions.json contains `"gemini": "0.8.0"`
- **THEN** the Dockerfile includes `ARG GEMINI_VERSION=0.8.0` immediately before Gemini's RUN instruction

#### Scenario: ARG scope is per-RUN
- **WHEN** the Dockerfile contains per-agent ARG declarations
- **THEN** each ARG only applies to the RUN instruction immediately following it, preserving layer cache isolation

### Requirement: Install command uses version ARG

Each agent's install command SHALL be modified to use its version ARG value:
- **npm agents** (Gemini, Codex, Pi): append `@${<AGENT>_VERSION}` to the package name
- **curl agents** (Claude, Copilot, Opencode): pass the version to the install script via the appropriate mechanism (Claude: `-- ${VERSION}`, Copilot: `VERSION=${VERSION}` env var, Opencode: `--version ${VERSION}` flag)

#### Scenario: npm agent uses @tag
- **WHEN** Gemini's version ARG is `GEMINI_VERSION=0.8.0`
- **THEN** the RUN instruction is `RUN npm install -g @google/gemini-cli@${GEMINI_VERSION}`

#### Scenario: curl agent uses argument
- **WHEN** Claude's version ARG is `CLAUDE_VERSION=v2.1.195`
- **THEN** the RUN instruction is `RUN curl -fsSL https://claude.ai/install.sh | bash -s -- ${CLAUDE_VERSION}`

#### Scenario: Copilot uses env var
- **WHEN** Copilot's version ARG is `COPILOT_VERSION=v1.0.65`
- **THEN** the RUN instruction is `RUN VERSION=${COPILOT_VERSION} curl -fsSL https://gh.io/copilot-install | bash`

#### Scenario: Opencode uses flag
- **WHEN** Opencode's version ARG is `OPENCODE_VERSION=v0.0.55`
- **THEN** the RUN instruction is `RUN curl -fsSL https://opencode.ai/install | bash -s -- --version ${OPENCODE_VERSION}`

### Requirement: Base image hash includes versions

The base image hash (computed in `baseHash`) SHALL incorporate the version map so that changes to agent versions trigger an appropriate rebuild. Because version ARGs are part of the assembled Dockerfile content, the existing Dockerfile hashing approach naturally covers this.

#### Scenario: Version change triggers rebuild
- **WHEN** the version for any agent changes in versions.json
- **THEN** the assembled Dockerfile differs from the previous build, the hash changes, and the base image is rebuilt

#### Scenario: No version change means no rebuild
- **WHEN** the version map is unchanged between runs
- **THEN** the assembled Dockerfile is identical, the hash matches, and the cached base image is reused

### Requirement: Six agents are versioned

The version pinning system SHALL track versions for all six registered agents: Claude, Gemini, Codex, Copilot, Opencode, and Pi. Each agent fetches its version from its designated source:

| Agent | Source | API endpoint |
|-------|--------|-------------|
| Claude | GitHub tags | `github.com/anthropics/claude-code` tags |
| Gemini | npm registry | `@google/gemini-cli` latest |
| Codex | npm registry | `@openai/codex` latest |
| Copilot | GitHub releases | `github/copilot-cli` latest release |
| Opencode | GitHub releases | `opencode-ai/opencode` latest release |
| Pi | npm registry | `@earendil-works/pi-coding-agent` latest |

#### Scenario: Claude fetches from GitHub tags
- **WHEN** the version fetcher resolves Claude's version
- **THEN** it queries the GitHub API for `anthropics/claude-code` tags and returns the first non-pre-release tag (without the `v` prefix)

#### Scenario: npm agents fetch from registry
- **WHEN** the version fetcher resolves a npm agent's version (Gemini, Codex, or Pi)
- **THEN** it queries the npm registry JSON API for the `latest` version and returns the `version` field

#### Scenario: GitHub agents fetch from releases
- **WHEN** the version fetcher resolves Copilot or Opencode's version
- **THEN** it queries the GitHub API for the latest release and returns the `tag_name` (with the `v` prefix removed)

#### Scenario: Missing agent has no entry in versions.json
- **WHEN** an agent's install is not active (not registered in the agent map)
- **THEN** no ARG is generated for that agent and no fetch is attempted
