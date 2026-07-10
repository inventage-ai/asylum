## Context

Asylum builds two image tiers: a base image (`asylum:latest`, shared across all projects, assembled from core + ordered kit snippets + agent block + tail) and a per-project image (`asylum:proj-<hash>`, layered `FROM asylum:latest`). Kits are already tier-split by origin: `resolveKitTiers` (cmd/asylum/main.go) loads the global config file separately via `config.LoadFile` and marks kits declared there as base-tier, everything else as project-tier.

The per-kit `packages` field (and `shell.build`) does not follow this split. `collectPackages` reads the fully-merged `config.Config` — where `merge:"concat"` has already flattened global + project package lists into one slice — and hands the whole map to `EnsureProject`. `EnsureBase` has no packages concept. So a package declared once in the global config is rebuilt into every project's image instead of once into the shared base.

The install-command generation lives in `image.generateProjectDockerfile`: it validates package names, then writes `apt-get` as `USER root` and npm/pip/cx-lang/run as the container user.

## Goals / Non-Goals

**Goals:**
- Global-config packages/build install into the base image (built once, shared).
- Project-config packages/build install into the project image (unchanged behavior for project-declared entries).
- Uniform across all package types: `apt`, node `npm`, python `pip`, `cx-lang`, `shell.build`.
- Reuse one install-command generator for both tiers.

**Non-Goals:**
- Changing the node kit's static base `DockerSnippet` (fixed default tools like typescript/eslint stay as-is).
- Per-kit interleaving of package installs within the base Dockerfile — a single block after all kits is sufficient.
- Dedup of a package that appears in both layers (re-install is a no-op; not required).
- Any config schema change.

## Decisions

### Provenance capture: reuse the `resolveKitTiers` pattern, intersected with effective base kits
`resolveKitTiers` already loads the global config file on its own and intersects the global-declared kits with the effective `allKits` (which reflects `--kits` flags and `disabled: true` toggles) to produce `globalKits`. Extend the same seam to also produce a global-packages map, but gate each provider kit's packages on membership in `globalKits`.

- **Global packages**: run `collectPackages` against the global-only config loaded by `config.LoadFile(~/.asylum/config.yaml)`, then drop the entries for any provider kit not present in `globalKits`. This gating is uniform across all package types — a global `npm`/`pip`/`cx-lang`/`apt` list or `shell.build` is included only if its provider kit (node/python/cx/apt/shell) survives into the base kit set.
- **Project packages**: run `collectPackages` against the merged config, then subtract the global slice per package type (order preserved; because `merge:"concat"` appends project entries after global ones, subtraction is a prefix strip, but a value-based "remove global entries" is used to stay robust).

**Why the intersection matters**: the base image installs a provider kit (node/fnm, python/uv, cx) only if that kit is in `globalKits`. `--kits python` or a project-level `kits.node.disabled: true` removes node from `allKits`, hence from `globalKits`, so the base image has no fnm/node. Reading global packages straight from the config file (without gating) would still emit `RUN npm install -g …` into the base block, which then fails to build because node is absent. Gating on `globalKits` keeps the package installs consistent with the kits actually baked into the base image. `apt`/`run` are gated the same way for uniformity, even though `apt-get` and the shell are always present.

Alternative considered: thread provenance through `config.Load` by returning a per-layer breakdown. Rejected as heavier — it changes a widely-called signature and adds a data model when the global config is already re-read cheaply in `resolveKitTiers`.

### Placement: one block, after kits, before agents
Global package installs are emitted as a single Dockerfile block inserted after all ordered kit snippets and before the agent block in `assembleDockerfile`. Because every providing kit precedes the block, npm (needs fnm/node), pip (needs uv), and cx-lang (needs cx) all resolve without per-kit ordering. This keeps the change to a single insertion point and preserves the existing rule that agent version bumps invalidate only agent layers.

The block is NOT a state-tracked ordering "source" (`kit:<name>`); it is a fixed-position block like the agent block, so it does not participate in `docker_source_order`.

### Hashing: fold the global block into `baseHash`
`baseHash` currently hashes core/tail assets, ordered kit snippets, the agent block, entrypoint snippets, banner lines, and host user identity. Add the rendered global-package block to this hash so a global package change flips `asylum.hash` and triggers a base rebuild. The existing `baseRebuilt` cascade then rebuilds project images.

### Shared generator
Extract the per-type install-command writers from `generateProjectDockerfile` into a shared helper that emits the `USER root` apt block and `USER <user>` npm/pip/cx-lang/run runs for a given packages map. `generateProjectDockerfile` calls it for project packages; the base assembly calls it for global packages. Package-name validation is shared too, so invalid global package names fail the base build early.

## Risks / Trade-offs

- **Base rebuild cascade** → Editing a global package rebuilds the base image and all project images. This is the intended tradeoff for "install once, share everywhere"; the first rebuild after a global-package edit is slower than today. Accepted.
- **Duplicate installs when a package is in both layers** → The package installs in base and may re-run in the project image. Harmless no-op for npm/apt/uv; not deduped to keep the split simple. Documented in the `package-tiering` spec.
- **Provider kit excluded by `--kits` or `disabled: true`** → A global package whose provider kit is not in `globalKits` would emit an install command into the base block with no installer present (e.g. `npm install -g` without node), failing the base build. Mitigated by gating the global-packages map on `globalKits` membership (see the provenance-capture decision): excluded kits' global packages are dropped from the base block entirely.
- **Subtraction correctness** → If project subtraction were computed as a naive prefix strip and merge order ever changed, it could misattribute. Mitigated by removing global entries by value rather than by index/length.

## Migration Plan

No data migration. On the next asylum invocation after upgrade, any global-declared packages move from the project image into the base image: the base image rebuilds once (picking up the global packages), project images rebuild via the cascade and no longer carry those packages. Rollback is reverting the binary — the config format is unchanged, so older versions read the same config and simply place all packages back in the project image.

## Open Questions

None — the four design decisions (scope, ordering, cascade, dedup) were confirmed with the user before proposing.
