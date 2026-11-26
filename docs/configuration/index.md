# Configuration

Asylum uses a layered YAML config system. Settings are merged from multiple sources, with later sources overriding earlier ones.

## Config Files

| Priority | File | Purpose |
|----------|------|---------|
| 1 (lowest) | `~/.asylum/config.yaml` | Global defaults |
| 2 | `.asylum` | Project config (commit this) |
| 3 (highest) | `.asylum.local` | Personal overrides (never committed) |

[CLI flags](flags.md) override all config files. You can also use [`asylum config`](../commands/config.md) to interactively manage kits, credentials, and isolation.

!!! tip "YAML extension alias"
    Both project config files also accept a `.yaml` suffix — `.asylum.yaml` and `.asylum.local.yaml` — so editors recognize them as YAML for syntax highlighting. The canonical names without the extension remain the default; pick whichever you prefer per repo (but don't keep both copies of the same layer, asylum errors out to avoid silent ambiguity).

## Where Options Are Accepted

`.asylum` is usually committed, so anyone who can push to the repository decides what it contains. To keep a cloned repository from widening its own sandbox, each option has a scope that says which files accept it:

| Scope | Accepted in | Options |
|-------|-------------|---------|
| any | all three files | `version`, `agent`, `env`, `ports`, enabling kits not listed below, and kit options such as `packages`, `build`, `versions`, `default-version`, `onboarding`, `shadow-node-modules`, `tab-title`, `count` |
| personal | `~/.asylum/config.yaml`, `.asylum.local` | `volumes`, `env-from-host`, `default-resume`, `agents.<agent>.config`, `agents.<agent>.companions`, `agents.<agent>.seed`, kit options `credentials`, `isolation`, `allow-agent-terminal-title`, `schemes`, and enabling the `docker`, `dropshare`, `iterm`, `cx` and `rtk` kits |
| global | `~/.asylum/config.yaml` | `release-channel`, `version-check-interval` |

Personal options can give the container access outside the project directory or extra privileges. Global options configure asylum itself rather than one project. [CLI flags](flags.md) are not restricted.

Disabling a kit (`disabled: true`) is accepted in every file, because it only narrows the sandbox.

When a file sets an option it does not accept, asylum ignores that option and lists it with file, line, and the YAML to move to the right file. In a terminal it then asks whether to start without these options or abort. Your answer is remembered for the project until the ignored options change. Without a terminal, asylum only prints the warnings. Asylum never edits your config files to move options.

Keys at the top level of `.asylum` and `.asylum.local`, and directly under `kits`, `agents` and their entries, must be plain strings: YAML aliases, merge keys and tags (`*name`, `<<:`, `!!binary`) are refused there. Anchors and aliases are fine inside a value, for example a `packages` list.

### Keeping `.asylum.local` Personal

Asylum refuses to start when `.asylum.local` (or `.asylum.local.yaml`) is tracked by git, because a repository that ships one could set personal options. Untrack it with `git rm --cached .asylum.local`.

On each start in a git repository, asylum adds both filenames to the repository's `.git/info/exclude` unless git already ignores them, so they are not committed by accident. This file is local to your clone and is never shared.

!!! warning "Not a boundary against the agent yet"
    The project directory is writable from inside the container, so the agent can still edit `.asylum.local`. Scopes protect you from what a repository ships, not from the agent you run in it. See [Security Model](../concepts/security-model.md).

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
  title:
    tab-title: "🤖 {project}"

ports:
  - "3000"
  - "8080:80"

env:
  MY_API_KEY: "abc123"
  DEBUG: "true"
```

## Example `.asylum.local`

```yaml
kits:
  docker: {}

volumes:
  - ~/shared-data:/data:ro
```

## Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Default agent: `claude`, `gemini`, `codex` (default: `claude`) |
| `release-channel` | string | Self-update channel: `stable` or `dev` (global config only) |
| `version-check-interval` | string | How often to refresh cached agent versions in the background, as a Go duration (default: `24h`). Force a refresh any time with [`asylum update`](../commands/update.md) (global config only) |
| `kits` | map | Kit configurations (see [Kits](../kits/index.md)) |
| `agents` | map | Agent configurations (see [Config Isolation](../concepts/isolation.md)) |
| `ports` | list | Port forwarding rules |
| `volumes` | list | Additional volume mounts (not in `.asylum`) |
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

Mount additional directories. Volumes are accepted in `~/.asylum/config.yaml` and `.asylum.local` only, because a mount can expose anything on the host, such as credentials in `~/.m2/settings.xml`:

```yaml
# ~/.asylum/config.yaml or .asylum.local
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

Asylum hands environment variables to `docker run` through an env file on stdin rather than as `-e NAME=value` arguments, so their values do not show up in `ps` or in process logs. Variables an env file cannot hold (invalid UTF-8, a `NAME=value` line of 64 KiB or more, unusual names) are still passed as `-e`. `--debug` lists them as `-e NAME=***`. Values remain visible to anyone who can run `docker inspect` on the container.

### Passing Host Variables (`env-from-host`)

`env` only takes literal values. To pass a secret without writing it into a config file, map a container variable to a host variable:

```yaml
# ~/.asylum/config.yaml or .asylum.local
env-from-host:
  GH_TOKEN: ASYLUM_GH_TOKEN   # container name: host name
```

When the container is created, Asylum reads `ASYLUM_GH_TOKEN` from the environment `asylum` runs in (e.g. set per project by direnv) and sets `GH_TOKEN` in the container. The different host name keeps the host's own `gh` login untouched.

- Unset or empty host variables are skipped.
- A host value with newlines or invalid UTF-8, a `NAME=value` line of 64 KiB or more, or a name the env file cannot hold (e.g. with a space) aborts the start with an error naming the variable.
- Asylum never writes the value to a file. Like all environment variables, it is kept off the `docker run` command line and masked in `--debug` (see above). It is still visible in `docker inspect`.
- The value is fixed at container creation. A running container keeps the old value; the new one applies once the container is recreated (after the last session exits, or with `--rebuild`).
- A name also set in `env` or with `-e` keeps that literal value.
- `env-from-host` is only honored in `~/.asylum/config.yaml` and `.asylum.local` (or `.asylum.local.yaml`). In `.asylum` (or `.asylum.yaml`), which is usually committed, it is ignored and reported, so a committed `.asylum` cannot map host environment variables into the container (see [Where Options Are Accepted](#where-options-are-accepted)).

## Merge Rules

When multiple config files define the same field:

| Type | Behavior |
|------|----------|
| **Scalars** (agent, release-channel, version-check-interval) | Last value wins |
| **Lists** (ports, volumes) | Concatenated across layers |
| **Maps** (env, env-from-host, kits, agents) | Merged per key, last value wins |
