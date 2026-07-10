## Why

Per-kit `packages` (node `npm`, python `pip`, `apt`, `cx-lang`) and `shell.build` run-commands are always installed into the per-project image — even when they are declared in the global config (`~/.asylum/config.yaml`). Kits are already tier-split by origin (`resolveKitTiers` reads the global config separately to decide base-image vs project-image kits), but the `packages`/`build` fields are not: `collectPackages` flattens the fully-merged config and feeds everything to `EnsureProject`. As a result a globally-configured package (e.g. `@mermaid-js/mermaid-cli`) is rebuilt into every project image separately instead of once into the shared base image.

## What Changes

- Package installs gain provenance-based tiering, mirroring the existing kit tiering:
  - Packages/build declared in the **global** config install into the **base image** (built once, shared across projects).
  - Packages/build declared in **project** configs (`.asylum`, `.asylum.local`) install into the **project image** (current behavior, minus what moved to base).
- Scope covers all package types — `apt`, node `npm`, python `pip`, `cx-lang` — and `shell.build` run-commands.
- `EnsureBase` gains a global-packages input and emits the install commands as a base Dockerfile block placed **after all kit snippets** and **before the agent block**, so every providing kit (node/fnm, python/uv, cx) is already installed. Global packages participate in the base image hash, so changing one triggers a base rebuild (which cascades to project images).
- The apt/npm/pip/cx-lang/run install-command generation is factored so the base and project tiers share one implementation instead of duplicating it.
- The node kit's static base `DockerSnippet` (fixed default tools like typescript/eslint) is unchanged — this change only concerns the user-configurable `packages`/`build` fields.

## Capabilities

### New Capabilities
- `package-tiering`: Determines whether each configured package/build entry installs into the base image or the project image, based on which config layer declared it (global → base, project → project).

### Modified Capabilities
- `image-build`: `EnsureBase` accepts and installs global-config packages as a Dockerfile block after kit snippets and before agents, and includes them in the base image hash. `EnsureProject` receives and installs only project-layer packages. The install-command generation is shared across both tiers.

## Impact

- **Code**:
  - `cmd/asylum/main.go` — `resolveKitTiers` (or a sibling) additionally produces the global-packages map; `collectPackages`/`ensureImages` wiring splits packages into global vs project tiers.
  - `internal/image/image.go` — `EnsureBase` signature and Dockerfile assembly/hash gain a global-packages block; `EnsureProject` receives only project-layer packages; the apt/npm/pip/cx-lang/run writers become shared.
- **Behavior**: Editing a global package now rebuilds the base image (cascading to all project images) rather than only the current project image — accepted tradeoff for "install once, share everywhere".
- **Config/back-compat**: No config schema change. Existing project-level `packages`/`build` behavior is unchanged; only global-declared entries relocate from the project image to the base image.
