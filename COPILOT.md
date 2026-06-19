GitHub Copilot CLI (copilot) — Asylum integration notes

Auth
- Copilot's token precedence: `COPILOT_GITHUB_TOKEN` → `GH_TOKEN` → `GITHUB_TOKEN`. The token must have the "Copilot Requests" permission.
- When the `github` kit is active, Asylum mounts the host's `gh` credentials into the container (writes `~/.config/gh/hosts.yml`). The copilot agent's launch wrapper reads `gh auth token` at startup and exports it as `GH_TOKEN` so copilot authenticates non-interactively — no extra config needed.
- Without the `github` kit (or if `gh` is unauthenticated on the host), copilot falls back to its own interactive device-flow login, which may not work in headless containers.

Config & LSP
- User config dir: `~/.copilot` (overridden via `COPILOT_HOME`, which the agent sets to the in-container path).
- Sessions live under `~/.copilot/session-state/<session-id>/` plus a SQLite session store. `HasSession` returns true iff that directory has at least one subdirectory.
- LSP config: user-level `~/.copilot/lsp-config.json` or repository-level `.github/lsp.json` (repo file is auto-mounted via the project mount).
- To enable LSP servers, install them inside the container (e.g., `npm i -g typescript-language-server`) or include them in a kit.

MCP servers
- MCP servers are configured in `~/.copilot/mcp-config.json` and managed via the `copilot mcp` subcommands. The GitHub MCP server is built-in; no extra setup needed for it.
- See `openspec/specs/copilot-mcp/spec.md` for Asylum's MCP plumbing contract.

Session resume
- Copilot's `session-state` directory is global to the config dir and copilot's `--resume` picker lists every recent session regardless of which project they came from. Auto-passing `--resume` purely on `session-state` content would expose unrelated projects' context.
- Asylum tracks an owned per-project marker under `<configDir>/asylum-projects/<encoded-path>/.has_session`, written after the first successful `asylum copilot` launch in a given project. `--resume` is only passed when that marker exists, so a fresh project never inherits another project's session list automatically.
- Cross-environment resume (e.g. host ↔ container ↔ codespace) is not supported by upstream copilot — sessions are local to where they were created.

Install
- Image install snippet: `curl -fsSL https://gh.io/copilot-install | bash` followed by `~/.local/bin/copilot --version`. Failure now aborts the image build (no `|| true` mask).

Notes for implementers
- Follow the existing agent patterns in `internal/agent/*.go` (register via RegisterInstall, implement Agent interface, use wrapZsh and term.ShellQuoteArgs).
- Avoid reworking Copilot auth; rely on the github kit's credential mount + the agent wrapper's GH_TOKEN export.
