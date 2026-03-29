## Why

Agents running inside the Asylum sandbox have no awareness of their environment — they don't know they're containerized, what tools are available, what constraints exist, or how to reach the host. This leads to agents wasting time discovering capabilities, making incorrect assumptions (e.g., trying to install tools that are already present), or missing capabilities entirely (e.g., not knowing `host.docker.internal` resolves to the host).

## What Changes

- Add a `RulesSnippet` field to the `Kit` struct, following the existing pattern of `DockerSnippet`, `EntrypointSnippet`, and `BannerLines`. Each kit can contribute a markdown snippet describing what it provides.
- Add an `AssembleRulesSnippets` aggregation function, consistent with the existing `AssembleDockerSnippets` / `AssembleBannerLines` pattern.
- Generate a sandbox rules file at `~/.asylum/projects/<container>/sandbox-rules.md` on each container start, assembled from a core template plus kit-contributed snippets.
- Mount that file read-only into the container at `<project>/.claude/rules/asylum-sandbox.md` so Claude Code picks it up automatically via its rules discovery mechanism.
- The core template covers: sandbox identity, user/permissions, host connectivity, base tooling (git, gh, glab, Docker CLI, jq, ripgrep, etc.), and general constraints. Kit snippets add tool-specific context (e.g., "Java 17/21/25 available via mise, default is 21").

## Capabilities

### New Capabilities
- `sandbox-rules`: Generation and mounting of a sandbox context rules file assembled from a core template and per-kit snippets.

### Modified Capabilities
<!-- No existing spec-level requirements are changing. The kit system gains a new optional field but existing kit behavior is unaffected. -->

## Impact

- `internal/kit/kit.go`: New `RulesSnippet` field on `Kit`, new `AssembleRulesSnippets` function.
- `internal/kit/*.go`: Each kit definition gains a `RulesSnippet` describing its tools.
- `internal/container/container.go`: New function to generate the rules file; `RunArgs` mounts it.
- `cmd/asylum/main.go`: Calls rules generation before container start, passes resolved kits to `RunOpts`.
- Currently Claude-specific (`.claude/rules/`), but the mechanism is agent-agnostic at the generation layer — only the mount path is agent-specific.
