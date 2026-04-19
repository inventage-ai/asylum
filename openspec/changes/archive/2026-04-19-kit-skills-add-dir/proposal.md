## Why

Kits that provide Claude Code skills (`agent-browser`, `ast-grep`) currently deliver them by `mkdir -p "$HOME/.claude/skills/<name>"` in the entrypoint and bind-mounting a staged directory over the top. In `shared` agent-config mode — where `$HOME/.claude` is a bind mount of the user's real `~/.claude` on the host — this pollutes the host filesystem with empty skill directories that persist after the container exits, silently scribbling into the user's personal Claude directory. PR #25 attempts a fix but (a) guards the mount with a directory-existence check that breaks on the second container start because the `mkdir` artifact persists and the guard then skips the mount, and (b) the approach fundamentally cannot avoid creating the host-side directory in shared mode. Claude Code supports `--add-dir <path>` with automatic discovery of `<path>/.claude/skills/*`, which removes the need to touch `$HOME/.claude/skills/` at all.

## What Changes

- Add a `ProvidesSkills bool` field to `kit.Kit` indicating the kit stages one or more Claude skills.
- Define a shared container path `/opt/asylum-skills/.claude/skills/<skill-name>/` where skill-providing kits stage their skills at image-build time.
- `agent-browser` and `ast-grep` kits stop using `/tmp/asylum-kit-skills-<name>` and stop emitting the entrypoint `mkdir`+`mount --bind` logic; their `DockerSnippet` now stages directly under the shared root and `NeedsMount: true` is removed from both.
- The Claude agent's launch command conditionally appends `--add-dir /opt/asylum-skills` when any active kit declares `ProvidesSkills: true`. The container layer passes that signal through to the agent command builder.
- The entrypoint installs a `claude` shell wrapper (or alias) so that interactive invocations in a secondary shell inside the container also pick up `--add-dir /opt/asylum-skills`.
- No change to the `~/.agents` host mount introduced by PR #25 — it remains valuable for users whose host-installed skills symlink into `~/.agents`, independent of this change.
- Out of scope: the `cx` and `rtk` kits, which mount non-skill artifacts (`~/.claude/rules/cx.md`, hooks, `RTK.md`, `settings.json` edits) and have their own shared-mode pollution problem tracked in issue #29.

## Capabilities

### New Capabilities
- `kit-skills-delivery`: the shared mechanism by which kits expose Claude Code skills to the container — a container-owned staging root and the agent-launch integration that makes Claude discover them.

### Modified Capabilities
- `browser-kit`: skill delivery switches from entrypoint bind-mount into `~/.claude/skills/agent-browser/` to build-time staging under `/opt/asylum-skills/.claude/skills/agent-browser/`, and the kit declares `ProvidesSkills: true` instead of `NeedsMount: true`.
- `ast-grep-kit`: same change pattern as `browser-kit`.
- `agent-interface`: Claude command generation conditionally appends `--add-dir /opt/asylum-skills` when any active kit provides skills.

## Impact

- **Code:** `internal/kit/kit.go` (new field), `internal/kit/agent_browser.go`, `internal/kit/astgrep.go` (DockerSnippet rewrite, snippet removals, field changes), `internal/agent/claude.go` and `internal/agent/agent.go` (command signature or options plumbing for the `--add-dir` decision), `internal/container/container.go` (feed kit list into agent command construction), `assets/entrypoint.core` (shell wrapper for interactive `claude`).
- **Users:** In shared config mode, no more empty directories created in `~/.claude/skills/`. Existing empty directories from prior asylum versions remain on host until the user cleans them up manually; a CHANGELOG note will cover this. Non-shared-mode users see no functional change.
- **Other agents (Gemini, Codex, OpenCode):** Unaffected. Skills are a Claude-only concept and the `--add-dir` flag is Claude-only.
- **SYS_ADMIN / cap-add --cap-add SYS_ADMIN:** Still required by `cx` and `rtk`. No change to container capabilities from this proposal.
- **Dependencies:** None added.
