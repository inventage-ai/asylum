# Asylum

Agent-agnostic Docker sandbox for AI coding agents (Claude Code, Gemini CLI, Codex). Single Go binary, cross-compiled for ARM and x86.

## Change Management

This project uses [OpenSpec](https://openspec.dev) for structured change management. Use the `/opsx:propose` skill to start a new change, `/opsx:apply` to implement, and `/opsx:archive` to archive completed changes. See `openspec/` for specs and change history.

## Architecture

- **Go** (latest stable) — single binary, no runtime dependencies beyond Docker
- Cross-compiled for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
- Shells out to Docker CLI via `os/exec` and `syscall.Exec` (process replacement)
- Layered YAML config: `~/.asylum/config.yaml` → `$project/.asylum` → `$project/.asylum.local` → CLI flags
- Embedded assets (Dockerfile, entrypoint.sh) via `go:embed`
- Manual CLI argument parsing with passthrough semantics (unknown flags forwarded to agents)
- One external dependency: `gopkg.in/yaml.v3`

### Project Structure

```
cmd/asylum/main.go          CLI entry point, argument parsing, dispatch
internal/
  agent/                    Agent interface + Claude/Gemini/Codex implementations
  config/                   Layered YAML config loading, merging, volume parsing
  container/                Docker run arg assembly, volume/env/port orchestration
  docker/                   Thin Docker CLI wrapper (build, inspect, prune)
  image/                    Two-tier image management with hash-based rebuild detection
  log/                      Colored terminal output (info/success/warn/error/build)
  ssh/                      SSH directory setup and key generation
assets/
  Dockerfile                Container image definition (embedded via go:embed)
  entrypoint.sh             Container startup script (embedded via go:embed)
  assets.go                 go:embed declarations
```

### Key Behaviors

- Agent config is seeded from host on first run (`~/.claude` → `~/.asylum/agents/claude/`), but resume is skipped for that first session since seeded data doesn't represent a container session.
- Base image rebuild invalidates all project images (the `baseRebuilt` flag cascades to `EnsureProject`).
- Container names are deterministic: `asylum-<sha256(project_dir)[:12]>`.
- Project directory is mounted at its real host path (not `/workspace`), preserving absolute paths.

## Code Style

### General

- **Less code is better.** Every line must earn its place. Avoid defensive boilerplate, speculative abstractions, and "just in case" code paths.
- Use modern Go: generics where they reduce duplication, errors as values, `slices`/`maps` packages.
- No unnecessary interfaces — don't create an interface until there are two implementations. A concrete type is fine.
- Keep functions short: one concern per function, early returns for error cases.
- Use `if err != nil { return err }` — don't wrap errors unless the wrapper adds information the caller doesn't already have.
- **Do not add fields, config options, or functionality without consulting the user.** If something seems needed but isn't explicitly requested, ask first.

### Comments

Code comments are used sparingly. Comprehensible and expressive code (consistent, logical naming) is preferred.

Comments are added when they contribute to much faster, better understanding in two cases:
- To explain **why** something was done, when it is not apparent from the context.
- To explain **what** is being done, if the code is necessarily difficult to understand.

If a log line explains what is happening, any comment above that line which essentially says the same thing is redundant and should not be added.

### Naming

- Package names: short, lowercase, no underscores. Avoid stutter (`config.Config` is fine, `config.ConfigConfig` is not).
- Functions/methods: verb-noun (`buildImage`, `loadConfig`). Getters drop the `Get` prefix (`Name()`, not `GetName()`).
- Variables: short-lived vars can be short (`f`, `err`, `cmd`). Longer-lived vars get descriptive names.
- Constants: `CamelCase`, not `SCREAMING_SNAKE`.

### Error Handling

- Return `error` from functions that can fail. Don't panic except for programmer errors.
- Wrap errors with `fmt.Errorf("context: %w", err)` only when the wrapper adds value.
- Log errors at the point of handling, not at the point of returning.
- Use the project's `log` package for user-facing output, not `fmt.Println` or the standard `log` package.

### Testing

- Use Go's built-in `testing` package. No test frameworks.
- Table-driven tests for functions with multiple input/output cases.
- Test files live next to the code they test (`config_test.go` next to `config.go`).
- Test the important logic: config merging, volume shorthand parsing, session detection, command generation, hash computation. Don't test trivial getters.
- Use `testdata/` directories for fixture files.

## Dependencies

Only `gopkg.in/yaml.v3` — everything else is standard library. ANSI colors are hand-rolled, CLI parsing is manual (to support passthrough semantics). Avoid adding dependencies unless they save significant effort.

## CI/CD

- **CI** (`.github/workflows/ci.yml`): Runs `go test` and `go vet` on every push/PR to main, then builds all four targets.
- **Release** (`.github/workflows/release.yml`): Triggered by version tags (`v*`). Builds binaries with version baked in and publishes them as GitHub release assets.
- **Install script** (`install.sh`): Detects OS/arch and downloads the correct binary from the latest GitHub release.

To release: `git tag v0.x.0 && git push origin v0.x.0`

## What NOT to Do

- Do not add Docker SDK. Shell out to the `docker` CLI — it's simpler and avoids a huge dependency tree.
- Do not create unnecessary abstractions, utility packages, or helper functions for one-off operations.
- Do not add config options, features, or agent support without consulting the user.
- Do not attempt to fix git corruption (broken packfiles, bad objects, etc.) yourself. Always prompt the user to resolve it.
