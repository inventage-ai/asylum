# Agent-Browser Kit

Browser automation via [agent-browser](https://github.com/vercel-labs/agent-browser) — an AI-native browser automation CLI.

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **agent-browser** CLI for browser automation
- **Chrome** (Chrome for Testing, installed via `agent-browser install`)
- **Claude Code skill** — staged inside the container at `/opt/asylum-skills/.claude/skills/agent-browser/` and loaded via Claude's `--add-dir` flag, teaching Claude the snapshot-ref interaction pattern. Nothing is written to your host `~/.claude/`.

## Configuration

```yaml
kits:
  agent-browser: {}
```

The old `browser:` key still works as an alias.

## Dependencies

Depends on the [Node.js](node.md) kit (agent-browser is installed via npm).

## Usage

```sh
# Navigate to a page
agent-browser open https://example.com

# Get interactive elements with refs
agent-browser snapshot -i

# Interact using refs
agent-browser click @e1
agent-browser fill @e2 "search query"

# Re-snapshot after page changes
agent-browser snapshot -i
```

Run `agent-browser --help` for all commands.
