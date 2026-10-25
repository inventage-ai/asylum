# Proposal

## Why

Projects often need a secret inside the container under a fixed name (e.g. a project-scoped GitHub PAT as `GH_TOKEN`), while on the host it must live under a different name so it does not override the host's own tools (e.g. the host `gh` login). `env:` only takes literal values, so the secret would have to be written into a config file; the only alternative today is a shell wrapper around `asylum` (GitHub issue #38).

## What Changes

- New top-level config key `env-from-host`: a map from container variable name to host variable name. At container creation, asylum reads each host variable and passes its value into the container under the mapped name.
- Unset or empty host variables are skipped. Asylum never writes the values to any file.
- `env-from-host` is honored only in user-maintained configs (`~/.asylum/config.yaml`, `.asylum.local`). In the repo-committed project layer (`.asylum` / `.asylum.yaml`) it is ignored with a warning, so a committed project config cannot map host environment variables into the container.
- A name set in `env` (including `-e` on the command line) takes precedence over the same name in `env-from-host`.

## Capabilities

### New Capabilities
- `env-from-host`: passing host environment variables into the container under a configurable name, restricted to user-maintained config layers.

### Modified Capabilities

## Impact

- `internal/config`: new `EnvFromHost` field, merge, and project-layer filtering in `Load`.
- `internal/container`: new RunArgs for the resolved variables.
- Docs: `docs/configuration/index.md`, `assets/asylum-reference.md`; `CHANGELOG.md`.
- Not addressed: masking of values in `--debug` output (#40) and visibility via `docker inspect`.
