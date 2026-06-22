## Context

`asylum` runs on the host and launches the agent inside a container. The `openspec` kit installs the OpenSpec CLI into that container but does nothing else: no project init, no rules guidance. The in-container agent therefore improvises `openspec init` with guessed defaults.

OpenSpec's workflow selection (our preference: `custom` profile with `verify` instead of `sync`) is stored in a global config file, `~/.config/openspec/config.json`. This was verified empirically: with that file empty, `openspec init --tools claude` generates the `core` set (`propose, explore, apply, sync, archive`); with the file set to `{"profile":"custom","workflows":["propose","explore","apply","verify","archive"]}`, the *same* command instead generates `propose, explore, apply, verify, archive` and omits `sync` — matching this repo's `.claude/`. `openspec config set workflows "a,b"` rejects a comma string, and the only non-interactive profile preset is `core`, so seeding the JSON file is the clean reproducible lever.

## Goals / Non-Goals

**Goals:**
- One-command, non-interactive OpenSpec setup the agent can trigger: `asylum-openspec-init`.
- Deterministically produce our preferred workflow set (`verify`, not `sync`).
- Pick the correct OpenSpec `--tools` id for whichever agent is running.
- Make the setup discoverable (rules) and documented (kit docs).

**Non-Goals:**
- Changing OpenSpec's own behavior or contributing upstream.
- Per-project customization of the workflow set — the preference is global and identical across projects.
- Auto-running init on container start; setup stays user/agent triggered.

## Decisions

**Split static preference from dynamic agent.** The workflow preference is identical for every project, so it is baked into the image at build time (kit `DockerSnippet` writes `~/.config/openspec/config.json`). The agent is chosen per run against a shared image, so it is resolved at runtime via `ASYLUM_AGENT`. The script stays tiny because the profile/workflows are already in place — `openspec init` only needs `--tools`.
- *Alternative considered:* have the script write the global config at runtime. Rejected — redundant work on every run and easy to drift; baking it once in the image is cleaner and persists.

**`ASYLUM_AGENT` set centrally in `container.go`.** The active agent's name is universal identity, not kit- or agent-specific config. Setting it once from `opts.Agent.Name()` avoids repeating it across all six agent implementations. The script maps `copilot` → `github-copilot`; all other asylum agent names already match OpenSpec tool ids (`claude`, `codex`, `gemini`, `opencode`, `pi`).
- *Alternative considered:* add `ASYLUM_AGENT` to each agent's `EnvVars()`. Rejected — six identical lines for one universal fact.

**Idempotency via `openspec/` presence.** Fresh project → `openspec init --tools <id>`; existing → `openspec update --force`. This makes re-running safe and lets the agent fire the script without first checking state.

**Kit ships the script via `DockerSnippet`.** Consistent with how kits already install tooling; no new embedded-asset machinery. The script is short enough to live in the snippet heredoc.

**Kit becomes `TierDefault`.** The `openspec-kit` spec and `docs/kits/openspec.md` already describe it as default-on; the code's `TierOptIn` is the outlier. This change aligns code to spec.

## Risks / Trade-offs

- **OpenSpec global config path/format could change across CLI versions.** → The seeded JSON is minimal and matches observed 1.4.1 behavior; if it drifts, init falls back to a usable (core) set rather than failing. Pin/verify against the installed version in the init task.
- **Heredoc script in a Go string is easy to get subtly wrong (quoting).** → Keep the script minimal; an integration/e2e check confirms `asylum-openspec-init` produces the `verify` set.
- **Baked global config lives in the image's home; an agent that overlays/mounts `~/.config` could shadow it.** → Verified that asylum does not mount `~/.config/openspec`; revisit if a future kit does.

## Migration Plan

Additive. Existing containers pick up the script and seeded config on next image rebuild (the kit DockerSnippet hash changes, triggering a base-image rebuild). No config migration needed. Making the kit default-on only affects newly generated kit configs; existing configs are unchanged.

## Open Questions

- Confirm the exact seeded-config JSON against the OpenSpec version pinned by the kit (`@latest`) at implementation time — the first task is a quick spike to re-verify the field names (`profile`, `workflows`) still drive generation.
