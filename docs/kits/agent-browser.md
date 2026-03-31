# Agent-Browser Kit

Browser automation via [agent-browser](https://github.com/vercel-labs/agent-browser) — an AI-native browser automation CLI.

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **agent-browser** CLI for browser automation
- **Chrome** (Chrome for Testing, installed via `agent-browser install`)
- **Claude Code skill** — auto-mounted from the upstream project, teaching Claude the snapshot-ref interaction pattern

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
