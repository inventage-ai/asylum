# Configuration

Asylum uses a layered YAML config system. Settings are merged from multiple sources, with later sources overriding earlier ones.

## Config Files

| Priority | File | Purpose |
|----------|------|---------|
| 1 (lowest) | `~/.asylum/config.yaml` | Global defaults |
| 2 | `.asylum` | Project config (commit this) |
| 3 (highest) | `.asylum.local` | Local overrides (gitignore this) |

[CLI flags](flags.md) override all config files.

## Example `.asylum`

```yaml
agent: gemini

kits:
  java:
    versions: ["17"]
  node:
    onboarding: true
    packages:
      - "@anthropic-ai/claude-mcp-server-filesystem"
  docker: {}
  title:
    tab-title: "🤖 {project}"

ports:
  - "3000"
  - "8080:80"

volumes:
  - ~/shared-data:/data:ro

env:
  MY_API_KEY: "abc123"
  DEBUG: "true"
```

## Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Default agent: `claude`, `gemini`, `codex` (default: `claude`) |
| `release-channel` | string | Self-update channel: `stable` or `dev` |
| `kits` | map | Kit configurations (see [Kits](../kits/index.md)) |
| `agents` | map | Agent configurations |
| `ports` | list | Port forwarding rules |
| `volumes` | list | Additional volume mounts |
| `env` | map | Environment variables |

## Kits Configuration

Each kit is configured under the `kits` key. An empty map `{}` enables the kit with defaults:

```yaml
kits:
  docker: {}                    # enable with defaults
  java:
    versions: ["17", "21"]      # configure specific options
    default-version: "17"
  node:
    shadow-node-modules: false  # disable a feature
  github:
    disabled: true              # disable a default-on kit
```

See [Kits](../kits/index.md) for all available kits and their options.

## Ports

Forward ports from the container to the host:

```yaml
ports:
  - "3000"        # same port on host and container
  - "8080:80"     # host:container
```

## Volumes

Mount additional directories:

```yaml
volumes:
  - ~/data:/data:ro              # host:container:options
  - /tmp/shared:/tmp/shared      # absolute paths
```

Tilde (`~`) is expanded to the home directory. Valid mount options: `ro`, `rw`, `z`, `Z`, `shared`, `slave`, `private`, `cached`, `delegated`.

## Environment Variables

```yaml
env:
  API_KEY: "secret"
  DEBUG: "true"
```

Asylum also reads `.env` files from the project root (mounted as `--env-file`).

## Merge Rules

When multiple config files define the same field:

| Type | Behavior |
|------|----------|
| **Scalars** (agent, release-channel) | Last value wins |
| **Lists** (ports, volumes) | Concatenated across layers |
| **Maps** (env, kits, agents) | Merged per key, last value wins |
