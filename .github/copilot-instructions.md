# Copilot instructions for inventage-ai/asylum

Purpose: quick, repository-specific guidance for Copilot sessions (build/test, high-level architecture, and conventions).

---

## Build, test, and lint commands (run from repo root)

- Build (single):
  - make build  # produces build/asylum
- Build cross-platform binaries:
  - make build-all
- Build locally in Docker (used by install-local.sh):
  - ./install-local.sh

- Run full unit tests:
  - make test  # runs `go test ./...`
- Run a single test (package-level):
  - go test ./path/to/package -run '^TestName$' -v
  - Example: go test ./internal/config -run '^TestParseVolume$' -v
- Integration tests (require Docker):
  - make test-integration  # runs `go test -tags integration -v -timeout 30m ./integration/`
- E2E tests (require Docker):
  - make test-e2e  # runs `go test -tags e2e -v -timeout 30m ./e2e/`

- Lightweight checks (used in CI):
  - go vet ./...
  - gofmt -w . (formatting)

---

## High-level architecture (big picture)

- Single Go binary (`cmd/asylum/main.go`) that shells out to Docker to provide an AI-agent sandbox.
- Core behavior organized under `internal/`:
  - `agent/` — agent adapters (Claude/Gemini/Codex/OpenCode/Echo)
  - `config/` — layered YAML config loading & migrations
  - `container/` & `docker/` — docker arg assembly and helper wrappers
  - `kit/`, `image/` — kit-driven Dockerfile snippets and two-tier image management
  - `firstrun/`, `onboarding/`, `tui/` — first-run wizard and onboarding UI
  - `ports/`, `ssh/`, `selfupdate/` — runtime helpers
- Assets (Dockerfile.core, entrypoint templates) embedded via go:embed in `assets/`.
- Kit system: modular feature bundles that inject Dockerfile/entrypoint/config snippets per project.
- Config layering: `~/.asylum/config.yaml` → `$project/.asylum` → `$project/.asylum.local` → CLI flags.

Read docs/ and openspec/ for change management and kit references.

---

## Key repository conventions and gotchas (must-know)

- Docker is required for most integration/e2e workflows and for the build scripts that use containerized Go.
- Tests with build tags:
  - Integration tests use `-tags integration` and are excluded from `go test ./...`.
  - E2E tests use `-tags e2e`.
- Entrypoint policy: the container entrypoint must NOT install packages. All package/tool installation belongs in Dockerfile (kit or image build). Avoid adding installs to entrypoint scripts.
- Two-tier images: base image (shared, kit-driven) + per-project image. Changing base invalidates project images; image rebuild logic is in `internal/image`.
- Deterministic container naming: `asylum-<sha256(project_dir)[:12]>` — be careful when writing tests or scripts that expect container names.
- Config migration: the project supports v1→v2 migration in `internal/config` — preserve user edits when updating config programmatically.
- Do not add Docker SDK: project shells out to `docker` CLI by design.
- OpenSpec workflow: changes should follow openspec artifacts (see `openspec/`) — use `/opsx:propose`, `/opsx:apply`, `/opsx:archive` where appropriate.

---

## AI assistant / agent files to consider

- CLAUDE.md exists with project-specific notes; prefer it for Claude/OpenCode-related behavior.
- No `.cursorrules`, `.windsurfrules`, or AGENTS.md were found. If adding agent-specific rules/docs, place them at repo root or under `.github/`.

---

## CI notes

- CI runs `go test` and `go vet` and builds cross-platform targets. Keep changes to build flags and CI scripts consistent with Makefile targets.

---

## Where to find more documentation

- README.md — quickstart and kits overview
- docs/ — full docs site source (MkDocs)
- openspec/ — change management specs and task history
- CHANGELOG.md — release notes and behavior changes

---

If an existing .github/copilot-instructions.md exists, merge the above into it rather than replacing wholesale.

Created: .github/copilot-instructions.md — contains build/test commands, architecture summary, and repository-specific conventions. 

Would you like additional coverage (examples for single-package test commands, more CI details, or explicit guidance for Code Assistant prompts)?
