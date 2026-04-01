# Commands

## Usage

```
asylum [flags]                Start default agent
asylum [flags] -- [args]      Start agent with extra args
asylum [flags] shell          Interactive zsh shell
asylum [flags] shell --admin  Admin shell with sudo notice
asylum [flags] run <cmd>      Run command in container
asylum config                 Configure kits, credentials, and isolation
asylum cleanup                Remove current project's container, volumes, and data
asylum cleanup --all          Remove all Asylum images, volumes, and cached data
asylum version [--short]      Show version
asylum self-update [version]  Update to latest (or specific) version
asylum self-update --dev      Update to latest dev build
asylum self-update --safe     Emergency update (always dev, no checks)
```

## Default (Agent Mode)

Running `asylum` with no subcommand starts the configured agent (default: Claude Code) in the current project directory. The agent runs in YOLO mode (auto-approve all actions).

Use `--` to pass extra flags to the agent:

```sh
asylum -- --verbose
asylum -a gemini -- --sandboxed
```

See [CLI Flags](../configuration/flags.md) for all available flags.

## Subcommands

| Command | Description |
|---------|-------------|
| [`shell`](shell.md) | Interactive zsh shell in the container |
| [`run`](run.md) | Run a command in the container |
| [`config`](config.md) | Configure kits, credentials, and isolation |
| [`cleanup`](cleanup.md) | Clean up current project (or `--all` for everything) |
| [`version`](version.md) | Show version information |
| [`self-update`](self-update.md) | Update Asylum to a new version |
