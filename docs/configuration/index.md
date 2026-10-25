# Configuration

Asylum uses a layered YAML config system. Settings are merged from multiple sources, with later sources overriding earlier ones.

## Config Files

| Priority | File | Purpose |
|----------|------|---------|
| 1 (lowest) | `~/.asylum/config.yaml` | Global defaults |
| 2 | `.asylum` | Project config (commit this) |
| 3 (highest) | `.asylum.local` | Local overrides (gitignore this) |

[CLI flags](flags.md) override all config files. You can also use [`asylum config`](../commands/config.md) to interactively manage kits, credentials, and isolation.

!!! tip "YAML extension alias"
    Both project config files also accept a `.yaml` suffix — `.asylum.yaml` and `.asylum.local.yaml` — so editors recognize them as YAML for syntax highlighting. The canonical names without the extension remain the default; pick whichever you prefer per repo (but don't keep both copies of the same layer, asylum errors out to avoid silent ambiguity).

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
| `version-check-interval` | string | How often to refresh cached agent versions in the background, as a Go duration (default: `24h`). Force a refresh any time with [`asylum update`](../commands/update.md) |
| `kits` | map | Kit configurations (see [Kits](../kits/index.md)) |
| `agents` | map | Agent configurations (see [Config Isolation](../concepts/isolation.md)) |
| `ports` | list | Port forwarding rules |
| `volumes` | list | Additional volume mounts |
| `env` | map | Environment variables |
| `env-from-host` | map | Host environment variables passed in under another name (not in `.asylum`) |

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

### Passing Host Variables (`env-from-host`)

`env` only takes literal values. To pass a secret without writing it into a config file, map a container variable to a host variable:

```yaml
# ~/.asylum/config.yaml or .asylum.local
env-from-host:
  GH_TOKEN: ASYLUM_GH_TOKEN   # container name: host name
```

When the container is created, Asylum reads `ASYLUM_GH_TOKEN` from the environment `asylum` runs in (e.g. set per project by direnv) and sets `GH_TOKEN` in the container. The different host name keeps the host's own `gh` login untouched.

- Unset or empty host variables are skipped.
- Asylum never writes the value to a file and keeps it off the `docker run` command line, so it does not show up in `ps` or the `--debug` output. It is still visible in `docker inspect`. Exception: names starting with `DOCKER_` are passed inline, because the docker CLI itself reads them. Other variables the docker CLI also reads (e.g. `HTTPS_PROXY` with a remote `DOCKER_HOST`) affect the CLI as well while it starts the container.
- The value is fixed at container creation. A running container keeps the old value; the new one applies once the container is recreated (after the last session exits, or with `--rebuild`).
- A name also set in `env` or with `-e` keeps that literal value.
- `env-from-host` is only honored in `~/.asylum/config.yaml` and `.asylum.local` (or `.asylum.local.yaml`). In `.asylum` (or `.asylum.yaml`), which is usually committed, it is ignored with a warning, so a committed `.asylum` cannot map host environment variables into the container. Keep `.asylum.local` out of version control, because mappings there are honored.

## Merge Rules

When multiple config files define the same field:

| Type | Behavior |
|------|----------|
| **Scalars** (agent, release-channel, version-check-interval) | Last value wins |
| **Lists** (ports, volumes) | Concatenated across layers |
| **Maps** (env, env-from-host, kits, agents) | Merged per key, last value wins |
