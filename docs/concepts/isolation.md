# Config Isolation

Asylum controls how each agent's config directory (e.g. `~/.claude`) is managed inside the container. Three isolation modes determine where config lives on the host and how it's mounted.

## Modes

### `shared` (default)

The host config directory is mounted directly into the container. Changes sync both ways — anything the agent writes inside the container is immediately visible on the host, and vice versa.

| | Path |
|---|---|
| **Host** | `~/.claude` |
| **Container** | `~/.claude` |

**Use when:** you want identical config inside and outside the sandbox, or use the same API keys, MCP servers, and settings everywhere.

**Trade-off:** no isolation. A misbehaving agent can modify your host config.

### `isolated`

Agent config is stored in Asylum's own directory, separate from the host. On first run, Asylum seeds this by copying from the host config. After that, the two are independent.

| | Path |
|---|---|
| **Host** | `~/.asylum/agents/claude/` |
| **Container** | `~/.claude` |

**Use when:** you want a stable sandbox config that doesn't drift with host changes, but is shared across all projects.

**Trade-off:** changes to your host config (e.g. new MCP servers) won't appear in the container. You need to update the asylum copy separately, or delete it to re-trigger seeding.

### `project`

Each project gets its own independent agent config directory. On first run, Asylum seeds it from the host config, just like `isolated` mode.

| | Path |
|---|---|
| **Host** | `~/.asylum/projects/<container>/claude-config/` |
| **Container** | `~/.claude` |

**Use when:** you need completely separate agent state per project — different settings, different conversation history, no cross-contamination.

**Trade-off:** more directories to manage. Config changes must be made per-project.

## Configuration

Set the isolation mode in your config:

```yaml
# ~/.asylum/config.yaml (or .asylum / .asylum.local)
agents:
  claude:
    config: shared   # shared (default) | isolated | project
```

### First-Run Wizard

The [first-run wizard](first-run.md) prompts for Claude's isolation mode if one isn't configured. `shared` is pre-selected as recommended; the choice is saved to `~/.asylum/config.yaml`. In non-interactive environments, `shared` is used as the default.

Other agents (Gemini, Codex, Copilot, Pi) follow the same `shared`-by-default behavior unless explicitly configured.

## Config Seeding

In `isolated` and `project` modes, Asylum performs a one-time copy of your host agent config into the target directory on first run. This seeds API keys, settings, and other config so you don't have to set them up again.

After seeding, the copy is independent — changes to the host config don't propagate. To re-seed, delete the asylum copy and restart.

Seeding is skipped in `shared` mode since the host directory is used directly.

### First-Run Session Behavior

When config is freshly seeded, Asylum forces a new agent session (`-n` behavior) regardless of what the seeded data contains. The seeded files may include session state from the host, but they don't represent a container session.
