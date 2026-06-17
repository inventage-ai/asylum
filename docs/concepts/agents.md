# Agents

Asylum is primarily built and tested with Claude Code. Gemini CLI and Codex support is experimental — basic functionality works, but these agents receive less testing and may have rough edges.

All agents run in YOLO mode (auto-approve all actions) by default.

## Supported Agents

| Agent | Binary | Default Mode | Config Dir |
|-------|--------|-------------|------------|
| [Claude Code](https://claude.ai) | `claude` | `--dangerously-skip-permissions` | `~/.claude` |
| [Gemini CLI](https://github.com/google-gemini/gemini-cli) | `gemini` | `--yolo` | `~/.gemini` |
| [Codex](https://github.com/openai/codex) | `codex` | `--yolo` | `~/.codex` |
| [Pi](https://github.com/mariozechner/pi-coding-agent) | `pi` | (none) | `~/.pi` |

## Selecting an Agent

```sh
asylum                # Claude Code (default)
asylum -a gemini      # Gemini CLI
asylum -a codex       # Codex
asylum -a pi          # Pi
```

Or set it in your config:

```yaml
agent: gemini
```

## Config Isolation

Asylum controls how each agent's config directory is managed inside the container. Three modes are available: `shared` (host config mounted directly), `isolated` (default — separate copy in `~/.asylum/agents/<agent>/`), and `project` (per-project copy).

On first run with Claude, Asylum prompts you to choose a mode. See [Config Isolation](isolation.md) for full details.

## Passing Extra Args

Use `--` to pass flags to the agent:

```sh
asylum -- --verbose
asylum -a gemini -- --sandboxed
```

## Resume Behavior

Each agent resumes its previous session by default. Use `-n` to start fresh. See [Sessions](sessions.md) for details on how each agent detects previous sessions.

## Installing Multiple Agents

By default, only Claude Code is installed in the base image. To install additional agents:

```yaml
agents:
  claude: {}
  gemini: {}
  codex: {}
```

Or via CLI:

```sh
asylum --agents claude,gemini
```

Agent installation requires their kit dependencies (Gemini, Codex, and Pi need the `node` kit).

## Companions

A primary agent can declare other installed agents as **companions**. When the primary launches, each companion's config dir is also mounted (writable) and its env vars are merged into the container, but the companion itself is not launched. This enables Claude Code plugins that shell out to other agent CLIs — for example, the codex Claude Code plugin invokes `codex` from inside a Claude session and needs `~/.codex` to be present.

```yaml
agents:
  claude:
    config: shared
    companions: [codex]
  codex:
    config: shared
```

Notes:

- Companions are **one-directional**: declaring `codex` as a companion of `claude` does **not** make `claude` a companion of `codex`. Running `asylum codex` ignores Claude's companion list.
- Each companion's own `config` (shared/isolated/project) is used to resolve where its config dir comes from. So a companion in `shared` mode mounts the host's `~/.codex`; in `isolated` mode it mounts `~/.asylum/agents/codex`.
- Companions must be in the installed agent set. If a listed companion isn't installed, asylum errors before starting the container.
- On env var collisions between primary and companion, the primary wins and a warning is logged.
- The list merges last-wins across config layers. Set `companions: []` in `.asylum` or `.asylum.local` to clear a list inherited from `~/.asylum/config.yaml`.
