## Context

SSH key management is currently split across three locations:
- `internal/ssh/` — standalone `Init()` that generates keys to `~/.asylum/ssh/` and merges known_hosts
- `internal/container/container.go` — hardcoded volume mount of `~/.asylum/ssh/` → `~/.ssh`
- `cmd/asylum/main.go` — `ssh-init` command dispatch

This is the only tooling area that isn't part of the kit system. Converting it to a kit unifies the model, removes a manual setup step, and adds isolation options mirroring the agent config isolation pattern.

## Goals / Non-Goals

**Goals:**
- SSH as a `TierAlwaysOn` kit with configurable isolation (isolated/shared/project)
- Key generation happens automatically in the `CredentialFunc` for isolated and project modes
- Mount the host's real `~/.ssh/known_hosts` directly (rw) in non-shared modes
- Remove the `ssh-init` CLI command and `internal/ssh/` package
- Remove hardcoded SSH mount from container assembly

**Non-Goals:**
- SSH agent forwarding
- Supporting multiple key types (ed25519 only)
- First-run wizard step for SSH isolation (use sensible default, configurable in YAML)

## Decisions

### Three isolation levels matching the agent pattern

| Level | Key storage | Mounting | known_hosts |
|-------|-------------|----------|-------------|
| `isolated` (default) | `~/.asylum/ssh/` | Key pair mounted individually (ro) | Host file mounted rw if exists |
| `shared` | Host `~/.ssh/` | Entire `~/.ssh/` mounted (rw) | Included in directory mount |
| `project` | `~/.asylum/projects/<container>/ssh/` | Key pair mounted individually (ro) | Host file mounted rw if exists |

Config YAML:
```yaml
kits:
  ssh:
    isolation: isolated  # or shared, project
```

The `isolation` value is read via a new `SSHIsolation()` config accessor, defaulting to `"isolated"` when absent.

### Use CredentialFunc for key generation and mounting

The kit's `CredentialFunc` receives the container name via `CredentialOpts` and:
1. Reads isolation mode from config (passed through opts or looked up)
2. For `shared`: returns a single `CredentialMount` with `HostPath` = `~/.ssh/`, `Destination` = `~/.ssh/`, writable
3. For `isolated`/`project`: ensures the key directory exists, generates key if missing, returns `CredentialMount` entries for individual files

This reuses the existing credential infrastructure.

### Pass container name through CredentialOpts

The `CredentialOpts` struct needs a `ContainerName` field so the SSH kit can resolve the per-project path in `project` mode. This is a minimal addition to an existing type.

### Bypass credential mode gating for always-on kits

The credential loop in `container.go` skips kits when `KitCredentialMode()` returns `""` (unconfigured). Always-on kits with no config entry would be skipped. Fix: treat empty mode as `"auto"` when the kit's tier is `TierAlwaysOn`.

### Print public key on first generation

When the `CredentialFunc` generates a new key (in isolated or project mode), it prints the public key and a hint to add it to Git hosting. This matches the current `ssh-init` UX.

## Risks / Trade-offs

- **Breaking change**: Users who run `asylum ssh-init` manually will get an error. Mitigated by the kit auto-generating keys on first container start. Existing keys in `~/.asylum/ssh/` continue to work.
- **Shared mode trust**: `shared` mounts the entire host `~/.ssh` read-write. This is opt-in and matches what users expect from "shared" — but modifications inside the container affect the host.
- **known_hosts write-back**: In isolated/project modes, container modifications to known_hosts affect the host. This is the desired behavior but is a change from the current copy-based approach.
