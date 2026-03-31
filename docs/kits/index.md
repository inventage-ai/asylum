# Kits

Kits are modular bundles that group everything needed for a language or tool: installation, environment setup, caching, onboarding, and container configuration.

## Available Kits

| Kit | Description | Activation |
|-----|-------------|------------|
| [node](node.md) | Node.js LTS via fnm, global dev packages | Always on |
| [python](python.md) | Python 3 with uv, linters, formatters | Default |
| [java](java.md) | JDK 17/21/25 via mise, Maven, Gradle | Default |
| [docker](docker.md) | Docker-in-Docker with buildx and compose | Default |
| [github](github.md) | GitHub CLI (gh) | Default |
| [openspec](openspec.md) | OpenSpec CLI | Default |
| [shell](shell.md) | oh-my-zsh, tmux, direnv | Always on |
| [ports](ports.md) | Automatic port forwarding for web services | Always on |
| [ast-grep](ast-grep.md) | AST-based code search, lint, and rewrite (`sg`) | Opt-in |
| [agent-browser](agent-browser.md) | Browser automation via agent-browser | Opt-in |
| [cx](cx.md) | Semantic code navigation for AI agents | Opt-in |
| [apt](apt.md) | Extra apt packages in the project image | Opt-in |

## Activation Tiers

Kits have three activation levels:

| Tier | Behavior |
|------|----------|
| **Always on** | Active even if not mentioned in config. Cannot be disabled. |
| **Default** | Added to your config automatically when first detected. Active when present in config. |
| **Opt-in** | Only active if you explicitly enable it in your config. |

When Asylum detects new kits (e.g., after an update), it presents a single selection prompt with all new **Default** and **Opt-in** kits. Default kits are pre-selected; opt-in kits are unselected. Deselected kits are added as commented-out entries in your config.

## Enabling Kits

Add a kit to your `.asylum` config with an empty map to enable it with defaults, or with specific options:

```yaml
kits:
  docker: {}                  # enable with defaults
  java:
    versions: ["17", "21"]    # enable with options
```

## Disabling Kits

Default-tier kits can be disabled:

```yaml
kits:
  github:
    disabled: true
```

Always-on kits (node, shell, ports) cannot be disabled.

## Kit Resolution

When kits are configured explicitly (via config or `--kits` flag):

- **Full kit** (e.g., `java`) — activates the kit and all its sub-kits (maven, gradle)
- **Specific sub-kit** (e.g., `java/maven`) — activates the parent kit and only that sub-kit
- **Always-on kits** are added automatically regardless of config

## Sub-Kits

Some kits contain sub-kits for optional features:

| Kit | Sub-Kits |
|-----|----------|
| node | `node/npm`, `node/pnpm`, `node/yarn` |
| python | `python/uv` |
| java | `java/maven`, `java/gradle` |

When you enable a kit, all its sub-kits are included. To be selective, reference sub-kits directly with `--kits java/maven`.

## Dependencies

Kits can depend on other kits. For example, `openspec` depends on `node`. Missing dependencies are auto-activated at resolve time.
