# Packages

Asylum can install additional packages on top of the base image. Where a package lands depends on which config layer declares it: global packages go into the shared base image, project packages into a per-project image layer.

## Package Types

Packages are configured under each kit's config:

```yaml
kits:
  apt:
    packages:
      - libpq-dev
      - redis-tools
  node:
    packages:
      - "@anthropic-ai/claude-mcp-server-filesystem"
      - tsx
  python:
    packages:
      - pandas
      - numpy
  shell:
    build:
      - "curl -fsSL https://deno.land/install.sh | sh"
```

| Kit | Key | Installer | Runs as |
|-----|-----|-----------|---------|
| `apt` | `packages` | `apt-get install` | root |
| `node` | `packages` | `npm install -g` | claude |
| `python` | `packages` | `uv tool install` | claude |
| `shell` | `build` | Executed as-is | claude |

## Where Packages Install

Packages follow the config layer that declares them:

- **Global config** (`~/.asylum/config.yaml`) → installed in the shared **base image** (`asylum:latest`), built once and reused by every project.
- **Project config** (`.asylum`, `.asylum.local`) → installed in the **project image** (`asylum:proj-<hash>`), layered `FROM asylum:latest`.

A package whose provider kit is excluded (via the `--kits` flag or `disabled: true`) is skipped rather than installed without its toolchain.

## How It Works

When the global config declares packages, they are added to the base image after the kit layers, so a change to a global package rebuilds the base image (and, in turn, every project image).

When a project config declares packages, Asylum:

1. Generates a project Dockerfile (`FROM asylum:latest` + install commands)
2. Builds a project image tagged `asylum:proj-<hash>`
3. Caches the image — rebuilds only when the project packages change

If a project declares no packages of its own, its base image (`asylum:latest`) is used directly.

## Custom Build Commands

The `shell` kit's `build` key runs arbitrary commands during the project image build. Use this for tools that don't have a standard package manager:

```yaml
kits:
  shell:
    build:
      - "curl -fsSL https://deno.land/install.sh | sh"
      - "go install golang.org/x/tools/gopls@latest"
```

Each command is executed as your host user. Commands must not contain newlines.
