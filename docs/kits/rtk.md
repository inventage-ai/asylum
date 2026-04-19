# rtk Kit

Token-reduction proxy via [RTK](https://github.com/rtk-ai/rtk) — intercepts shell commands and compresses output to reduce LLM token usage by 60-90%.

**Activation: Opt-in** — only active if explicitly enabled in your config.

## What's Included

- **rtk** — CLI proxy that rewrites command outputs to strip noise (comments, whitespace, boilerplate)
- **Claude Code PreToolUse hook** — registered in `~/.claude/settings.json` to run `rtk hook claude` on every Bash tool call, transparently intercepting commands

## Configuration

```yaml
kits:
  rtk: {}
```

## How It Works

RTK registers a PreToolUse hook that intercepts every Bash command the agent runs. Commands like `git status`, `ls`, `grep`, etc. are transparently rewritten through RTK, which strips noise and compresses the output before the agent sees it.

At container start the entrypoint adds `rtk hook claude` to the `PreToolUse` hooks in `~/.claude/settings.json` (replacing any older file-path rtk entry from previous asylum versions), and mounts `RTK.md` so Claude picks up the usage notes. In shared agent-config mode this does modify your host `~/.claude/settings.json`.

## Usage

```sh
# Check token savings
rtk gain

# See savings over time
rtk gain --graph

# Discover optimization opportunities
rtk discover
```

RTK supports 100+ commands across files, git, testing, build/lint, and containers. See the [RTK documentation](https://github.com/rtk-ai/rtk) for the full list.
