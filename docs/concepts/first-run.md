# First Run

The first time you invoke `asylum` on a host (no `~/.asylum/config.yaml` present), a short wizard runs before any image is built. It captures the choices that shape the sandbox image — which agents to install, which language/tooling kits to enable, how Claude's config should be mounted, and whether to expose host credentials to the container.

## Detection

The wizard fires when `~/.asylum/config.yaml` is absent at the very start of an invocation. On subsequent runs the wizard skips the image-shaping questions (agents, kits) and only prompts for individual values that are still unconfigured (e.g. an isolation mode the user never set).

In non-interactive environments (no TTY), the wizard is skipped entirely and today's defaults are written silently.

## Steps

The wizard runs as a single multi-step flow. Pressing enter through every step accepts the defaults and yields the same configuration the silent-default code path would have written.

| Step | Default | Notes |
|---|---|---|
| **Agents** (multi-select) | `claude` | All registered agents except the `echo` test stub. Picking more bakes them into the image. |
| **Default agent** (single-select) | `claude` (if picked) else first selected | Only shown when more than one agent is picked in the previous step. |
| **Kits** (multi-select) | `TierDefault` kits pre-checked | Top-level kits only. Always-on kits (`ssh`, `ports`, `shell`, etc.) stay active automatically and aren't shown. `TierAvailable` opt-in kits appear unchecked. Sub-kits like `java/maven` are managed in the config file, not the wizard. |
| **Config isolation** (single-select) | `Shared with host (recommended)` | Only shown for Claude when `agents.claude.config` is unset. Choices are `shared`, `isolated`, `project` — see [Config Isolation](isolation.md). |
| **Credentials** (multi-select) | Currently configured selections pre-checked | Only shown when an active kit exposes credentials (e.g. `github`, `java/maven`) and the parent kit has no `credentials:` value yet. |

After the wizard completes, asylum writes a complete `~/.asylum/config.yaml` reflecting the selections:

- Picked agents appear uncommented in the `agents:` map; unpicked agents appear as `# agent-name:` so you can enable them later by uncommenting.
- Picked kits appear active in the `kits:` map; unpicked top-level kits appear commented with their authored example body intact.
- `TierAlwaysOn` and `Hidden` kits keep their authored presentation verbatim — the wizard never touches them.

## What happens next

```
parse args
    │
    ▼
load defaults from kits + agents
    │
    ▼
first-run wizard          ← prompts here, before any docker work
    │
    ▼
write ~/.asylum/config.yaml from your choices
    │
    ▼
reload config (so the image sees your selections)
    │
    ▼
ensureImages              ← base + project image build
    │   "Building sandbox image — this takes a few minutes the first time…"
    ▼
start container, exec agent
```

The pre-build line is suppressed when both images are cache hits — it only shows when at least one will actually rebuild.

## SSH on first run

When the container starts for the first time, the SSH kit generates an ed25519 key pair at `~/.asylum/ssh/` (or per-project under `~/.asylum/projects/<container>/ssh/` when project mode is selected). `ssh-keygen`'s normal output and the randomart preview are suppressed; asylum prints one line pointing at `asylum-reference.md`, which documents how to view the public key and add it to a Git hosting provider.

## Editing the config later

The wizard is a one-shot — it doesn't re-run after `~/.asylum/config.yaml` exists. To enable agents or kits you skipped:

- Hand-edit `~/.asylum/config.yaml` (the unselected entries are present as commented-out lines you can uncomment).
- Or run `asylum config` for an interactive kit/credential editor that operates on the existing file.

When new kits ship in a future asylum release that aren't yet in your config, the kit-sync prompt (separate from this wizard) fires on the next interactive run and offers them.
