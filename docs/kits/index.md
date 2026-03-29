# Kits

Kits are modular bundles that group everything needed for a language or tool: installation, environment setup, caching, onboarding, and container configuration.

## Available Kits

| Kit | Description | Default |
|-----|-------------|---------|
| [node](node.md) | Node.js LTS via fnm, global dev packages | Off |
| [python](python.md) | Python 3 with uv, linters, formatters | Off |
| [java](java.md) | JDK 17/21/25 via mise, Maven, Gradle | Off |
| [docker](docker.md) | Docker-in-Docker with buildx and compose | Off |
| [github](github.md) | GitHub CLI (gh) | **On** |
| [openspec](openspec.md) | OpenSpec CLI | **On** |
| [shell](shell.md) | oh-my-zsh, tmux, direnv | **On** |
| [apt](apt.md) | Extra apt packages in the project image | Off |

## Enabling Kits

Add a kit to your `.asylum` config with an empty map to enable it with defaults, or with specific options:

```yaml
kits:
  docker: {}                  # enable with defaults
  java:
    versions: ["17", "21"]    # enable with options
```

## Default-On Kits

Kits marked **On** in the table above are active unless explicitly disabled:

```yaml
kits:
  github:
    disabled: true     # disable a default-on kit
```

## Kit Resolution

When kits are configured explicitly (via config or `--kits` flag):

- **Full kit** (e.g., `java`) — activates the kit and all its sub-kits (maven, gradle)
- **Specific sub-kit** (e.g., `java/maven`) — activates the parent kit and only that sub-kit
- **Default-on kits** are added automatically unless disabled

## Sub-Kits

Some kits contain sub-kits for optional features:

| Kit | Sub-Kits |
|-----|----------|
| node | `node/npm`, `node/pnpm`, `node/yarn` |
| python | `python/uv` |
| java | `java/maven`, `java/gradle` |

When you enable a kit, all its sub-kits are included. To be selective, reference sub-kits directly with `--kits java/maven`.

## Dependencies

Kits can depend on other kits. For example, `openspec` depends on `node`. If a dependency is missing, Asylum emits a warning but continues.
