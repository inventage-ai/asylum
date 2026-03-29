## Context

Kits already contribute four types of content — `DockerSnippet`, `EntrypointSnippet`, `BannerLines`, and `CacheDirs` — each with a corresponding assembly/aggregation function. The rules snippet follows this exact pattern: a new string field on `Kit`, assembled by a new function, consumed at container start.

The project directory is bind-mounted from the host at its real path. Docker allows layering a more-specific file mount on top of an existing bind mount, so we can mount a single file at `<project>/.claude/rules/asylum-sandbox.md` without it existing on the host or polluting the host filesystem.

Claude Code discovers `.claude/rules/*.md` files in the project directory and loads them as additional instructions alongside CLAUDE.md. Rules without path-scoping frontmatter are always loaded at session start.

## Goals / Non-Goals

**Goals:**
- Agents receive accurate, up-to-date context about their sandbox environment on every session
- Kits can describe their own contributions (tools, capabilities, constraints) in natural language
- The rules file reflects the actual active kits for this container, not a static superset
- The file is generated fresh on each container start so it stays in sync with config changes
- No host filesystem pollution — the file exists only inside the container

**Non-Goals:**
- Supporting agents other than Claude for the mount path (other agents can be added later with their own mount targets)
- User-editable sandbox rules (users have their own CLAUDE.md and `.claude/rules/` for that)
- Making the rules file content configurable via `.asylum` YAML

## Decisions

### 1. New `RulesSnippet` field on Kit

Follow the established pattern: add `RulesSnippet string` to the `Kit` struct and an `AssembleRulesSnippets` function. Each kit provides a markdown fragment describing what it installs and any relevant usage notes.

*Alternative: A structured type (map of tool names to descriptions).* Rejected — markdown is more flexible, matches the output format, and is consistent with how `DockerSnippet`/`EntrypointSnippet` work as opaque strings.

### 2. Generation at container start, not image build

The rules file is generated in `~/.asylum/projects/<container>/sandbox-rules.md` each time a container is started. This means the content always reflects the current config (kits can change between runs).

*Alternative: Bake into the Docker image.* Rejected — the image is shared across projects with different kit configurations, and the rules need to reflect the per-project resolved kits.

### 3. Mount as `.claude/rules/asylum-sandbox.md`

The file is mounted read-only at `<project>/.claude/rules/asylum-sandbox.md`. This is purely additive — it doesn't shadow the user's CLAUDE.md or any existing rules files. Docker supports file-level mounts layered on top of directory bind-mounts.

*Alternative: Mount at `<project>/.claude/CLAUDE.md`.* Rejected — would shadow the user's own file if it exists at that path.

### 4. Two-tier kit contributions: `Tools` and `RulesSnippet`

Kits contribute to the rules file in two ways:
- **`Tools []string`**: Simple command names (e.g., `"gh"`, `"mvn"`). Aggregated via `AggregateTools` into a comma-separated "Kit Tools" line with kit attribution: `gh (github), mvn (java/maven)`. For kits that just make a command available without needing explanation.
- **`RulesSnippet string`**: Prose markdown fragment for kits that need to explain behavior (e.g., Docker's privileged mode, Java's version switching). Assembled via `AssembleRulesSnippets` into an "Active Kits" section.

A kit may have both (e.g., gradle has `Tools: ["gradle"]` and a `RulesSnippet` explaining mise integration).

### 5. Core template + kit tools + kit snippets assembly

The generated file has four sections:
1. **Header**: Asylum version, changelog link, pointer to reference doc
2. **Core section** (Go format string `sandboxRulesTemplate`): sandbox identity, user/permissions, host connectivity, base tools from Dockerfile.core only (git, Docker CLI, curl, wget, jq, yq, ripgrep, fd, make, cmake, gcc, vim, nano, htop)
3. **Kit Tools section**: aggregated from `Tools` fields — one-line list of `tool (kit-name)` entries
4. **Active Kits section**: assembled from `RulesSnippet` fields, prose descriptions

### 6. Reference document (embedded, read on demand)

A comprehensive Asylum reference doc (`assets/asylum-reference.md`) is embedded in the binary via `go:embed`. It covers container lifecycle, layered config system, all kit options, volume mounting, self-update, and troubleshooting. Written alongside the rules file at `~/.asylum/projects/<container>/asylum-reference.md` and mounted at `<project>/.claude/asylum-reference.md` (not in `rules/`, so not auto-loaded). The rules file points to it.

*Alternative: Bake into the Docker image.* Rejected — the reference doc is embedded in the Go binary and changes with Asylum versions. Mounting avoids image rebuilds when the doc changes.

### 7. Pass resolved kits and version through RunOpts

`RunOpts` gains `Kits []*kit.Kit` and `Version string` fields. The caller in `main.go` passes `allKits` and `version`. The rules generation, reference doc write, and mounting all happen inside `RunArgs`.

*Alternative: Generate the files in `main.go` and pass paths.* This would work but spreads the responsibility. Keeping it in `container` keeps all mount logic together.

## Risks / Trade-offs

- **Docker file mount on top of bind mount**: Docker supports this (more-specific mounts win), but if the user has a real `.claude/rules/asylum-sandbox.md` in their project, it will be shadowed. The `asylum-` prefix makes collision very unlikely. → Mitigation: distinctive filename prefix.
- **Rules file size**: If many kits contribute verbose snippets, the file could bloat Claude's context. → Mitigation: keep snippets concise; the CLAUDE.md style guide already favors brevity.
- **Claude-only mount path**: Other agents (Gemini, Codex) don't have an equivalent discovery mechanism. → Acceptable for now; the generation layer is agent-agnostic, only the mount path is Claude-specific. Agent interface can gain a `RulesPath` method later if needed.
