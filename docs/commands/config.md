# config

Interactive wizard for configuring kits, credentials, and isolation.

## Usage

```
asylum config
```

## Description

Opens an interactive TUI wizard with three tabs for managing your global Asylum configuration (`~/.asylum/config.yaml`). Requires a terminal — cannot be run in non-interactive environments.

The wizard modifies `~/.asylum/config.yaml` in place, preserving comments and formatting. Changes take effect on the next container start.

## Tabs

### Kits

Toggle kits on or off. Shows all available kits except always-on and hidden ones. Kits that are currently active are pre-selected.

- **Enabling** a kit adds it to your config (or removes `disabled: true` if it was previously disabled).
- **Disabling** a kit sets `disabled: true` on its config entry.

See [Kits](../kits/index.md) for details on each kit.

### Credentials

Control which host credentials the sandbox can access (read-only). Each credential-capable kit is listed with its label. Kits with `credentials: auto` are pre-selected.

- **Enabling** sets `credentials: auto` — the kit mounts its host credentials into the container.
- **Disabling** sets `credentials: false`.

### Isolation

Choose how Claude's config (`~/.claude`) is managed inside the sandbox:

| Option | Mode | Behavior |
|--------|------|----------|
| **Shared with host** | `shared` | Use host `~/.claude` directly. Changes sync both ways. |
| **Isolated** (recommended) | `isolated` | Separate from host, shared across projects. |
| **Project-isolated** | `project` | Separate config per project. No state shared between projects. |

See [Config Isolation](../concepts/isolation.md) for a full explanation of each mode.

## Examples

```sh
# Open the config wizard
asylum config
```

## Notes

- Only modifies the global config (`~/.asylum/config.yaml`). Project-level config (`.asylum`, `.asylum.local`) must be edited manually.
- The wizard loads your current merged config (global + project layers) to show accurate current state, but writes only to the global file.
- You can also edit `~/.asylum/config.yaml` directly — the wizard is a convenience, not a requirement.
