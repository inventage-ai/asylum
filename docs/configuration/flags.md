# CLI Flags

Flags can be used with any subcommand and override all config file settings.

## General Flags

| Flag | Description |
|------|-------------|
| `-a`, `--agent <name>` | Agent to use: `claude`, `gemini`, `codex` (default: `claude`) |
| `-p <port>` | Forward a port (repeatable, e.g., `-p 3000 -p 8080:80`) |
| `-v <volume>` | Mount a volume (repeatable, e.g., `-v ~/data:/data:ro`) |
| `-e KEY=VALUE` | Set environment variable (repeatable, last wins) |
| `--java <version>` | Java version in container |
| `--kits <list>` | Comma-separated kits to enable (e.g., `--kits java,python,docker`) |
| `--agents <list>` | Comma-separated agents to install (e.g., `--agents claude,gemini`) |
| `--continue` | Forwarded to the agent — resume the previous session |
| `--resume` | Forwarded to the agent — resume the previous session |
| `-n`, `--new` | Deprecated no-op (starting a new session is the default) |
| `--rebuild` | Force rebuild the Docker image |
| `--skip-onboarding` | Skip project onboarding tasks for this run |
| `--cleanup` | Alias for [`cleanup`](../commands/cleanup.md) command |
| `--version` | Alias for [`version`](../commands/version.md) command |
| `-h`, `--help` | Show help |

## Agent Passthrough

Use `--` to pass flags directly to the agent:

```sh
asylum -- --verbose
asylum -a gemini -- --sandboxed
```

Everything after `--` is forwarded to the agent command unchanged.

## Examples

```sh
# Start Gemini with port forwarding and an env var
asylum -a gemini -p 3000 -e API_KEY=abc123

# Resume Claude's previous session
asylum --continue

# Force a fresh image build (new session is already the default)
asylum --rebuild

# Use only Java and Docker kits
asylum --kits java,docker
```
