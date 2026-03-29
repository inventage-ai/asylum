# Agents

Asylum supports three AI coding agents. Each runs in YOLO mode (auto-approve all actions) by default.

## Supported Agents

| Agent | Binary | Default Mode | Config Dir |
|-------|--------|-------------|------------|
| [Claude Code](https://claude.ai) | `claude` | `--dangerously-skip-permissions` | `/home/claude/.claude` |
| [Gemini CLI](https://github.com/google-gemini/gemini-cli) | `gemini` | `--yolo` | `/home/claude/.gemini` |
| [Codex](https://github.com/openai/codex) | `codex` | `--yolo` | `/home/claude/.codex` |

## Selecting an Agent

```sh
asylum                # Claude Code (default)
asylum -a gemini      # Gemini CLI
asylum -a codex       # Codex
```

Or set it in your config:

```yaml
agent: gemini
```

## Config Seeding

On first run, Asylum copies your host agent configuration into `~/.asylum/agents/<agent>/`:

| Agent | Host Source | Asylum Copy |
|-------|-----------|-------------|
| Claude | `~/.claude` | `~/.asylum/agents/claude/` |
| Gemini | `~/.gemini` | `~/.asylum/agents/gemini/` |
| Codex | `~/.codex` | `~/.asylum/agents/codex/` |

This is a **one-time copy**. After seeding, the asylum copy is independent — changes to your host config won't propagate to containers (and vice versa).

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

Agent installation requires their kit dependencies (Gemini and Codex need the `node` kit).
