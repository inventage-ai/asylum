## Context

The image build has two tiers: a base image (`asylum:latest`) built by `EnsureBase`, and a project image (`asylum:proj-<hash>`) built by `EnsureProject`. Kits are split between these tiers by `resolveKitTiers` — kits in `~/.asylum/config.yaml` are global (base), all others are project-only.

The entrypoint script is assembled from `entrypoint.core` + kit `EntrypointSnippet`s + `entrypoint.tail` and baked into the base image. `EnsureProject` only uses `DockerSnippet`s, so project-level kit `EntrypointSnippet`s and `BannerLines` are never included anywhere.

This means a kit like cx, when enabled at the project level, gets its Docker commands run (e.g., `cx skill > /tmp/asylum-kit-rules/cx.md`) but its entrypoint logic (bind-mounting the rules file) is silently dropped.

## Goals / Non-Goals

**Goals:**
- Project kit `EntrypointSnippet`s run at container startup
- Project kit `BannerLines` appear in the welcome banner
- No changes to the base image build or entrypoint assembly
- Hash-based rebuild detection includes project entrypoint content

**Non-Goals:**
- Restructuring how kits are split between base and project tiers
- Making the base entrypoint aware of individual project kits

## Decisions

### 1. Project entrypoint as a separate script sourced by the base entrypoint

The project image already extends the base image via `FROM asylum:latest`. We generate a `/usr/local/bin/project-entrypoint.sh` script containing the project kit `EntrypointSnippet`s and `BannerLines`, COPY it into the project image, and have the base entrypoint source it if it exists.

**Alternative considered**: Rebuilding the full entrypoint in the project image. Rejected because it duplicates the core entrypoint logic and means project images must be rebuilt whenever the base entrypoint changes.

**Alternative considered**: Passing project entrypoint content as an environment variable. Rejected — too fragile and size-limited.

### 2. Source location: before the welcome banner in entrypoint.tail

The base entrypoint.tail sources `project-entrypoint.sh` (if it exists) immediately before the welcome banner block. This ensures project kit setup (bind-mounts, daemon starts, etc.) happens before the banner prints, and project `BannerLines` can be injected into the banner.

For banner lines, the project entrypoint exports a `PROJECT_BANNER` variable containing the assembled banner lines. The base entrypoint.tail checks for this variable and prints it in the banner block.

### 3. Project entrypoint assembly mirrors base pattern

`assembleProjectEntrypoint` in `image.go` concatenates `EntrypointSnippet`s from project kits, then appends a `PROJECT_BANNER` export with the assembled `BannerLines`. The result is a plain bash script with `set -e` at the top.

### 4. Hash includes project entrypoint

`generateProjectDockerfile` already drives the project image hash. Since the project entrypoint content is embedded in the Dockerfile (via COPY), it's implicitly part of the hash when the full Dockerfile string is hashed.

## Risks / Trade-offs

- **Ordering**: Project entrypoint runs after all base kit entrypoint snippets. This is correct since project kits may depend on base kit setup (e.g., PATH from fnm/mise).
- **`set -e` propagation**: The project entrypoint uses `set -e` independently. A failing snippet will abort the project entrypoint but the base entrypoint continues (since it sources with `|| true` — same resilience pattern as other optional features).
