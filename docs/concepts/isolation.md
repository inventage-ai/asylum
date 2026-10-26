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

On macOS, Claude keeps its login in the Keychain rather than in `~/.claude`, and the container cannot read the Keychain. So even in `shared` mode you log in once inside the container. The container stores that login in `~/.claude/.credentials.json`, and the shared mount keeps that file on the host.

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
    seed: full       # full (default) | settings | none — see Config Seeding
```

### First-Run Wizard

The [first-run wizard](first-run.md) prompts for Claude's isolation mode if one isn't configured. `shared` is pre-selected as recommended; the choice is saved to `~/.asylum/config.yaml`. In non-interactive environments, `shared` is used as the default.

Other agents (Gemini, Codex, Copilot, Pi) follow the same `shared`-by-default behavior unless explicitly configured.

## Config Seeding

In `isolated` and `project` modes, Asylum performs a one-time copy of your host agent config into the target directory on first run. This seeds API keys, settings, and other config so you don't have to set them up again.

By default the whole host directory is copied, including the transcripts, history and todos of every project. `seed` controls what is copied:

| Value | Copies |
|-------|--------|
| `full` (default) | The entire host config directory |
| `settings` | Only settings, global instructions, custom agents/commands/skills, and credential files stored in the config directory. Sessions, history, caches and logs stay on the host. |
| `none` | Nothing; the directory starts empty |

```yaml
# ~/.asylum/config.yaml or .asylum.local
agents:
  claude:
    config: project
    seed: settings
```

For Claude, `settings` copies `settings.json`, `CLAUDE.md`, `keybindings.json`, `agents/`, `commands/`, `skills/`, `rules/`, `output-styles/`, `hooks/` and `.credentials.json`. Other files that `settings.json` points to, such as a statusline script directly in `~/.claude`, are not copied; keep them under `hooks/` or use `full`. `plugins/` is left out because its files contain absolute host paths, so plugins enabled in `settings.json` need to be installed again in the container. On macOS Claude keeps its login in the Keychain rather than in `.credentials.json`, so you log in once per seeded directory. Claude's `.claude.json`, which holds user-scope MCP servers and onboarding state, is not seeded either, so those MCP servers must be added again in the container and the first-run prompts appear once. Each agent has its own list; OpenCode's `~/.opencode` holds no settings, so `settings` behaves like `none` there.

`seed` is only honored in `~/.asylum/config.yaml` and `.asylum.local`. In a project's `.asylum`, which is usually committed, it is ignored with a warning, so a repository cannot override your `seed` choice.

After seeding, the copy is independent — changes to the host config don't propagate. To re-seed, delete the asylum copy and restart.

Seeding is skipped in `shared` mode since the host directory is used directly.

### First-Run Session Behavior

When config is freshly seeded, Asylum forces a new agent session (`-n` behavior) regardless of what the seeded data contains. The seeded files may include session state from the host, but they don't represent a container session.
