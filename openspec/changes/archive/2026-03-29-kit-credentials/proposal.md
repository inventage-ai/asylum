## Why

The current credential handling mounts the entire `~/.m2/settings.xml` into the sandbox, exposing all Maven server credentials to the AI agent — not just the ones the project needs. This is a hardcoded first-run hack in `internal/firstrun/firstrun.go` with no kit awareness and no extensibility path for other ecosystems. Credentials are fundamentally ecosystem-specific (file formats, discovery, filtering), so kits should own them.

## What Changes

- Add `CredentialFunc` to the Kit struct — the first dynamic kit behavior. Each kit optionally provides a function that inspects the project and host, then returns scoped credential mounts.
- Implement Maven credential filtering on the `java/maven` sub-kit: parse the project's `pom.xml` for repository server IDs, look them up in the host's `~/.m2/settings.xml`, and generate a minimal settings.xml containing only matching `<server>` entries.
- Add a `credentials` field to KitConfig supporting three modes: `auto` (discover from project files), explicit list (kit-specific identifiers), or off (default).
- Replace the first-run credential Y/n prompt with a TUI multiselect listing all kits that provide credential support. Selected kits get `credentials: auto` written to `~/.asylum/config.yaml`.
- Integrate credential generation into container launch: call each active kit's CredentialFunc, write generated files to `~/.asylum/projects/<cname>/credentials/`, and bind-mount them read-only.

## Capabilities

### New Capabilities
- `kit-credentials`: Kit-owned credential system — CredentialFunc on Kit struct, CredentialMount type, container-launch integration, and Maven server ID filtering implementation.

### Modified Capabilities
- `first-run-onboarding`: Replace hardcoded credential file detection and Y/n prompt with TUI multiselect of credential-capable kits, writing `credentials: auto` to config.

## Impact

- `internal/kit/kit.go` — new types and field on Kit struct
- `internal/kit/java.go` — CredentialFunc on java/maven sub-kit (XML parsing of pom.xml and settings.xml)
- `internal/config/config.go` — new `Credentials` field on KitConfig with polymorphic parsing (string/bool/list)
- `internal/container/container.go` — credential generation and mount in appendVolumes
- `internal/firstrun/firstrun.go` — rewrite to TUI multiselect approach
- Standard library `encoding/xml` used for Maven XML parsing (no new dependencies)
