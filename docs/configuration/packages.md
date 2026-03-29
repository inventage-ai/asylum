# Packages

Asylum can install additional packages into a project-specific Docker image layer. When packages are configured, Asylum builds a project image on top of the base image.

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

## How It Works

When any packages are configured, Asylum:

1. Generates a project Dockerfile (`FROM asylum:latest` + install commands)
2. Builds a project image tagged `asylum:proj-<hash>`
3. Caches the image — rebuilds only when packages change

If no packages are configured, the base image (`asylum:latest`) is used directly.

## Custom Build Commands

The `shell` kit's `build` key runs arbitrary commands during the project image build. Use this for tools that don't have a standard package manager:

```yaml
kits:
  shell:
    build:
      - "curl -fsSL https://deno.land/install.sh | sh"
      - "go install golang.org/x/tools/gopls@latest"
```

Each command is executed as the `claude` user. Commands must not contain newlines.
