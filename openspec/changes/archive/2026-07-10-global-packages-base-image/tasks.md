## 1. Share the install-command generator

- [x] 1.1 Extract the apt/npm/pip/cx-lang/run install-command writers from `generateProjectDockerfile` in `internal/image/image.go` into a shared helper that renders a package block for a given packages map (apt as `USER root`, npm/pip/cx-lang/run as the container user).
- [x] 1.2 Move package-name validation (`validatePackageNames`, `knownPackageTypes`, run-command checks) so both the base and project paths validate through the shared helper.
- [x] 1.3 Update `generateProjectDockerfile` to call the shared helper for project packages; confirm output is byte-identical for existing project-only configs (existing tests still pass).

## 2. Install global packages in the base image

- [x] 2.1 Add a global-packages parameter to `EnsureBase` in `internal/image/image.go`.
- [x] 2.2 Render the global-package block via the shared helper and insert it in `assembleDockerfile` after all ordered kit snippets and before the agent block.
- [x] 2.3 Fold the rendered global-package block into `baseHash` so a global-package change flips `asylum.hash` and triggers a base rebuild.
- [x] 2.4 Ensure the block is emitted only when the global-packages map is non-empty (no stray `USER`/`RUN` lines otherwise).

## 3. Split package provenance in the CLI wiring

- [x] 3.1 Extend `resolveKitTiers` (or add a sibling) in `cmd/asylum/main.go` to produce a global-packages map from the global-only config already loaded via `config.LoadFile`, reusing `collectPackages`.
- [x] 3.2 Gate the global-packages map on `globalKits` membership: drop any provider kit's entries (npm→node, pip→python, cx-lang→cx, apt→apt, run→shell) when that kit is not in the effective base kit set, so excluded/disabled kits never emit an install command with no installer present.
- [x] 3.3 Compute the project-tier packages map as the merged config packages with the global entries removed by value (per package type), so global-declared packages no longer flow to the project image.
- [x] 3.4 Update `ensureImages` to pass the gated global-packages map to `EnsureBase` and the project-tier map to `EnsureProject`.

## 4. Tests

- [x] 4.1 Unit-test the shared generator: apt uses `USER root`, npm/pip/cx-lang/run use the container user, invalid names error.
- [x] 4.2 Test base image assembly places the global-package block after kit snippets and before the agent block, and that changing global packages changes `baseHash`.
- [x] 4.3 Test the provenance split: global-declared packages land in the base map only; project-declared packages land in the project map only; a package in both appears in base and does not break the build.
- [x] 4.4 Test that a config with only global packages (no project packages/snippets) yields `asylum:latest` from `EnsureProject`.
- [x] 4.5 Test provider-kit gating: with global `kits.node.packages` present, `--kits python` excludes the npm entry from the base map; a project-level `kits.node.disabled: true` does the same.

## 5. Verification

- [x] 5.1 Update `CHANGELOG.md` under **Unreleased** (Changed): global-config packages now install in the base image instead of the per-project image.
- [x] 5.2 Run `go test ./...` and `go vet ./...`; manually verify with a global `kits.node.packages` entry that the package lands in the base image and not the project image.
