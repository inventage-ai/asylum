## Context

Asylum's kit system provides a modular way to install tools into sandbox containers. Each kit is a Go struct registered via `init()` in `internal/kit/`, containing Docker snippets, entrypoint snippets, config defaults, and metadata. Three new kits are needed for popular agent plugins: ast-grep (structural code search), browser (Chromium + Playwright), and cx (semantic code navigation).

All three follow established patterns — the main design questions are around installation method, tier assignment, and dependency management.

## Goals / Non-Goals

**Goals:**
- Add three self-contained kit files following existing conventions
- Each kit installs cleanly in the Debian-based container image
- Kits integrate with config sync, banner, rules, and caching systems

**Non-Goals:**
- No new kit infrastructure or framework changes
- No sub-kits for any of the three (keep them simple)
- No new KitConfig fields — reuse existing `packages` for cx languages

## Decisions

### 1. Installation methods

- **ast-grep**: Install via npm (`npm install -g @ast-grep/cli`). Requires the `node` kit as a dependency. npm is simpler than cargo (no Rust toolchain needed) and the npm package is officially maintained.
  - *Alternative*: Install via cargo. Rejected — would require Rust toolchain in the image, adding ~500MB for a single tool.
  - *Alternative*: Download prebuilt binary from GitHub releases. Viable but npm is simpler and already available via the node kit.

- **browser**: Install Chromium via Playwright's npm package (`npx playwright install --with-deps chromium`). This handles both the browser binary and required system libraries (fonts, graphics libs). Requires the `node` kit as a dependency.
  - *Alternative*: Install Chromium via apt. Rejected — Playwright's install command handles dependency management more reliably and ensures browser/driver version compatibility.

- **cx**: Install via the project's install script (`curl -sL https://raw.githubusercontent.com/ind-igo/cx/master/install.sh | sh`). No kit dependencies — the script downloads a standalone Rust binary. Configurable languages (tree-sitter grammars) are installed via `cx lang add` during the Docker build.
  - *Alternative*: Install via `cargo install cx-cli`. Rejected — same Rust toolchain concern as ast-grep.

### 2. Kit tiers

All three kits use **TierOptIn** (only active if the user explicitly enables them in config). These are specialized tools that not every project needs — they add significant image size (especially browser) and should only be installed when deliberately requested.

### 3. Cache directories

- **ast-grep**: None needed. No persistent cache beyond npm's own cache (handled by node kit).
- **browser**: Playwright cache at `/home/claude/.cache/ms-playwright`. This is where Chromium binaries are stored. Caching avoids re-downloading ~280MB on every image rebuild.
- **cx**: Index database (`.cx-index.db`) lives in the project directory — no container-level cache needed.

### 4. cx language configuration

cx uses tree-sitter grammars that must be explicitly installed via `cx lang add <language>`. Rather than adding a new KitConfig field, we reuse the existing `packages` field — the same field used by apt, node, and python kits for their respective installable items.

Config looks like:
```yaml
kits:
  cx:
    packages:        # tree-sitter language grammars
      - python
      - typescript
      - go
```

Implementation:
- Add `"cx": "cx-lang"` to `collectPackages` in `cmd/asylum/main.go`
- Add `"cx-lang"` to `knownPackageTypes` in `internal/image/image.go`
- Add a `cx-lang` block in `generateProjectDockerfile` that runs `cx lang add <lang>` for each configured language
- Languages are installed in the project image (not base), so different projects can have different language sets

### 5. File structure

One file per kit in `internal/kit/`:
- `astgrep.go` — ast-grep kit
- `browser.go` — browser kit
- `cx.go` — cx kit

## Risks / Trade-offs

- **Browser kit image size (~400MB)**: Chromium + system deps significantly increase image size. Mitigation: TierDefault means it's only installed when the user enables it; not in the base image by default.
- **cx install script**: Curling a script from GitHub is a supply-chain risk. Mitigation: This is a common pattern (same as how gh CLI is installed), and the script downloads from GitHub releases with known checksums.
- **Playwright version drift**: The installed Playwright version (via npm) must match the Chromium version. Mitigation: `npx playwright install` handles this automatically — it installs the compatible browser for the installed npm package version.
