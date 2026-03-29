## Why

Agents are increasingly using specialized tools beyond basic coding: structural code search (ast-grep), semantic code navigation (cx), and browser automation (Playwright). These tools require non-trivial installation (system dependencies, language runtimes, browser binaries) that users shouldn't have to figure out themselves. Adding kits for these popular plugins makes them one-line config additions.

## What Changes

- **New `ast-grep` kit**: Installs the `ast-grep` (`sg`) CLI via npm for AST-based code search, linting, and rewriting.
- **New `browser` kit**: Installs Chromium and Playwright inside the container so agents can automate browsers (navigate, screenshot, interact with web pages).
- **New `cx` kit**: Installs the `cx` CLI (Rust binary via install script) for semantic code navigation — file overviews, symbol search, definition lookup, and reference finding.

## Capabilities

### New Capabilities

- `ast-grep-kit`: Kit definition for ast-grep CLI installation and configuration.
- `browser-kit`: Kit definition for Chromium + Playwright installation, including system dependencies (libs, fonts) and cache directory management.
- `cx-kit`: Kit definition for cx CLI installation and configuration.

### Modified Capabilities

_(none)_

## Impact

- **Code**: Three new files in `internal/kit/` (one per kit), following existing kit patterns. Small additions to `cmd/asylum/main.go` and `internal/image/image.go` to support cx language packages in the project image build pipeline.
- **Images**: Each kit adds Docker layers. The browser kit is the heaviest (~400MB for Chromium + deps). All three are opt-in or default tier — not always-on — so they only affect images when enabled.
- **Dependencies**: ast-grep requires the `node` kit (installed via npm). cx and browser have no kit dependencies (cx installs a standalone binary; browser installs system packages directly).
- **Config**: Each kit adds a config snippet to `~/.asylum/config.yaml` when detected by kit sync.
