## Why

Agent CLIs (Claude Code, Gemini CLI, Codex) are hardcoded in `Dockerfile.core` and always installed, regardless of which agent the user actually uses. This wastes image build time and disk space, and makes it impossible to add new agent CLIs without changing the core image. Extracting agent installation into a configurable system — analogous to profiles for language toolchains — lets users select exactly which agents are present in the base image.

## What Changes

- New agent installation registry: each agent CLI has an install snippet (Dockerfile RUN instructions), optional profile dependency (e.g., Gemini/Codex depend on node for npm), and a banner line
- New `agents` field in config YAML (all layers + CLI flag `--agents`); nil (unspecified) defaults to `["claude"]`, empty means none
- Agent install snippets are inserted into the assembled Dockerfile between profile snippets and the tail
- Agent install snippets are removed from `Dockerfile.core` — core only installs managers (fnm, mise, uv) and system tools
- Base image hash includes active agent install snippets
- Welcome banner shows version lines only for installed agents
- New opencode agent: installed via `go install`, no profile dependencies
- fnm + Node.js LTS remains in core (needed by node-dependent agents and is lightweight infrastructure)

## Capabilities

### New Capabilities
- `agent-install`: Agent CLI installation registry, config-driven agent selection, Dockerfile snippet assembly, and dependency resolution against active profiles

### Modified Capabilities
- `profile-image-build`: Base image assembly now includes agent install snippets between profile snippets and tail
- `container-image`: Welcome banner agent version lines are dynamic based on installed agents

## Impact

- **assets/Dockerfile.core**: Remove Claude Code, Gemini CLI, and Codex install lines
- **internal/agent/install.go** (new): Agent install registry with install snippets, profile dependencies, banner lines
- **internal/agent/opencode.go** (new): Opencode agent runtime implementation + install definition
- **internal/config/config.go**: New `Agents *[]string` field, merge semantics (last-wins, same as profiles)
- **cmd/asylum/main.go**: Resolve agents from config, validate profile dependencies, pass to image builder, `--agents` CLI flag
- **internal/image/image.go**: `assembleDockerfile` and `baseHash` include agent install snippets
- **assets/entrypoint.tail**: Agent banner lines become dynamic (assembled at build time like profile banner lines)
