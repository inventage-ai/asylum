# SSH Kit

SSH key management for containers. Generates keys automatically and mounts them into `~/.ssh/`.

**Activation: Always On** — active in every container, no configuration needed.

## What It Does

On first container start, the SSH kit:

1. Generates an Ed25519 key pair (if one doesn't exist)
2. Prints the public key for you to add to GitHub/GitLab
3. Mounts the key pair into `~/.ssh/` (read-only)
4. Mounts the host's `~/.ssh/known_hosts` (read-write, if it exists)

## Configuration

```yaml
kits:
  ssh:
    isolation: isolated   # default
```

## Isolation Modes

!!! note "Personal config only"
    Set `isolation` in `~/.asylum/config.yaml` or `.asylum.local`. In a project's `.asylum` it is ignored, because `shared` mounts your host `~/.ssh` ([details](../configuration/index.md#where-options-are-accepted)).

| Mode | Key Storage | Behavior |
|------|------------|----------|
| `isolated` (default) | `~/.asylum/ssh/` | Shared across all projects, separate from host |
| `shared` | Host `~/.ssh/` | Mounts entire host SSH directory (read-write) |
| `project` | `~/.asylum/projects/<container>/ssh/` | Separate key pair per project |

### Isolated (default)

Keys are generated in `~/.asylum/ssh/` and shared across all projects. The host's `~/.ssh/known_hosts` is mounted read-write so new host keys added inside containers are available on the host.

### Shared

The host's entire `~/.ssh/` directory is mounted read-write. No key generation occurs. Use this if you want to use your existing host SSH keys directly.

### Project

Each project gets its own key pair in `~/.asylum/projects/<container>/ssh/`. Useful if you need different deploy keys per project. The host's `~/.ssh/known_hosts` is still mounted read-write.

## Replacing Keys

You can replace the generated key with your own by placing your key files in the appropriate directory:

- **Isolated**: `~/.asylum/ssh/`
- **Project**: `~/.asylum/projects/<container>/ssh/`
