## Context

Today two kits (`agent-browser`, `ast-grep`) deliver Claude Code skills by:

1. At image-build time, invoking `npx skills add ... --copy` into a temporary working directory and moving the result to `/tmp/asylum-kit-skills-<name>`.
2. At container-start time, running a `mkdir -p "$HOME/.claude/skills/<name>"` followed by `sudo mount --bind /tmp/asylum-kit-skills-<name> "$HOME/.claude/skills/<name>"`.

`$HOME/.claude` inside the container is a bind mount. In **isolated/project** agent-config modes it maps to `~/.asylum/agents/claude/`, which is asylum-owned and invisible to the user — so pollution there is harmless. In **shared** mode, it maps to the user's real `~/.claude/` on the host. Every `mkdir -p` in shared mode creates a directory on the user's host filesystem that persists after the bind mount is torn down on container exit.

PR #25 guards the `mkdir`+`mount` pair with `[ ! -e "$dest" ] && [ ! -L "$dest" ]`. This avoids the `mkdir` error when the host already has a skill (commonly a symlink into `~/.agents/`), but it introduces a second bug: on the *next* container start in non-shared mode the empty directory created by the *previous* run still exists, the guard short-circuits, and the bind mount never happens — so the skill silently disappears. It also still creates host pollution on the first shared-mode run when the skill isn't already installed on host.

Claude Code supports `--add-dir <path>` and automatically loads skills from `<path>/.claude/skills/*` (per the Claude Code docs, "Skills from additional directories"). This lets us deliver kit-provided skills from a path asylum fully owns, without touching `$HOME/.claude/skills/` in any mode.

## Goals / Non-Goals

**Goals:**

- Deliver kit-provided Claude skills without `mkdir` or `mount --bind` against `$HOME/.claude/skills/`.
- No creation of files or directories on the host's real `~/.claude/` in shared mode, for the two skill-providing kits.
- Eliminate the second-container-start bug in PR #25's approach by removing the stateful `mkdir` artifact entirely.
- Keep the user-facing behavior identical: both skills remain discoverable by Claude with no extra user action.
- Remove `NeedsMount: true` from the two affected kits so they no longer require `SYS_ADMIN` capability on their account.

**Non-Goals:**

- Fixing shared-mode pollution from `cx` and `rtk` kits. Those mount non-skill artifacts (rules files, hooks, `RTK.md`, `settings.json` edits) and require a different fix — tracked in issue #29.
- Removing the `~/.agents` host bind mount added in PR #25. It remains useful for users whose host-installed skills symlink into `~/.agents/`, independent of this mechanism.
- Extending a general "kits contribute agent CLI args" hook system. This change adds a targeted single-contribution path for one flag (`--add-dir`); a general hook would be built when a second use case appears.
- Supporting skill delivery via this mechanism for non-Claude agents. Skills and `--add-dir` are Claude-specific.
- Cleaning up empty skill directories left in users' `~/.claude/skills/` by previous asylum versions. Documented in CHANGELOG; users clean up manually.

## Decisions

### Decision: Stage skills under a single shared container root

**Choice:** All skill-providing kits stage their skill directory under `/opt/asylum-skills/.claude/skills/<skill-name>/` at image-build time. The agent launcher passes a single `--add-dir /opt/asylum-skills` regardless of how many skill-providing kits are active.

**Rationale:** One root means one `--add-dir` flag regardless of kit count. Avoids any need for per-kit agent-arg contributions, deduplication logic, or a new Kit hook for contributing CLI args. The directory layout encodes the convention; no runtime assembly needed.

**Alternatives considered:**
- Per-kit `--add-dir /opt/asylum-skills/<name>`: would require collecting contributions from multiple kits and merging them into the agent command. More plumbing for zero user-visible benefit.
- Leave files under `/tmp/asylum-kit-skills-<name>` and pass multiple `--add-dir` flags: same downside as above and stretches the `/tmp` convention into a permanent home for installed artifacts.

### Decision: `ProvidesSkills bool` on Kit, not a function hook

**Choice:** Add a plain `ProvidesSkills bool` field to `kit.Kit`. Set to `true` on `agent-browser` and `ast-grep`.

**Rationale:** The signal is static (known at kit-registration time, not a function of config or runtime state). A bool matches existing static kit metadata like `NeedsMount`, `Hidden`, `Tools`. A function hook would be overfit for this one flag.

**Alternatives considered:**
- `AgentArgsFunc func(ContainerOpts) []string`: general hook for kits to contribute arbitrary agent args. Premature — we have one use case. Can be introduced later when a second kit needs to contribute a different flag.
- Inferring skill provision from the kit's DockerSnippet staging path: fragile, implicit. A declarative bool is clearer.

### Decision: Agent command gains active-kits context via options, not raw signatures

**Choice:** The `Agent.Command` interface grows to accept (or receive via an options struct) a signal indicating whether skill kits are active. Concrete shape: extend the signature to `Command(resume bool, extraArgs []string, opts AgentCmdOpts)` where `AgentCmdOpts` carries a `KitSkillsDir string` (empty when no skill-providing kit is active; set to `/opt/asylum-skills` when one or more are).

**Rationale:** Keeps the decision centralized in the agent implementation rather than smearing `--add-dir` across the container layer. Other agents (Gemini, Codex, OpenCode) ignore the field. Using an options struct is future-proof — more fields can be added without breaking the interface again.

**Alternatives considered:**
- Post-process the returned command in the container layer to inject `--add-dir`: separates the decision from the Claude-specific knowledge of the flag, which is worse encapsulation.
- New `ClaudeCommand` method specific to Claude: other agents wouldn't need to implement it, but the container layer would then have to type-switch on agent type to know whether to call it. Worse.

### Decision: Interactive `claude` inside the container uses a shell function/alias

**Choice:** The entrypoint installs a zsh function (`claude` wrapper) that invokes the real `claude` binary with `--add-dir /opt/asylum-skills` prepended when the shared-skills dir exists. Implemented in the same shell-setup step where PATH and other env are configured.

**Rationale:** The primary agent launch is done by asylum via `agent.Command`, but users also spawn `claude` directly from a secondary terminal inside the container. Without the wrapper, those invocations miss kit skills. A function is lightweight, introspectable (`which claude` shows it), and scoped to the container only.

**Alternatives considered:**
- Environment variable consumed by Claude (e.g. `CLAUDE_ADD_DIR`): no such variable exists in Claude Code.
- Symlinking `claude` to a wrapper script: more files, no benefit over a function.
- Asking users to remember the flag: unacceptable.

### Decision: Gate the wrapper and the launch flag on directory existence, not on active-kit lookup

**Choice:** Both the agent-launch code and the shell wrapper check whether `/opt/asylum-skills/.claude/skills` contains any entries before passing `--add-dir`. The container layer sets `opts.KitSkillsDir` based on the static active-kit list, but the shell wrapper runs in the container (no access to the kit list) and so uses a filesystem check.

**Rationale:** Avoids passing `--add-dir` to a nonexistent directory when a user customizes the image. Also keeps the shell wrapper self-contained — no generated content from the kit build.

**Alternatives considered:**
- Hard-coding the flag in the wrapper always: Claude may warn about missing directories. Cheap to avoid.

### Decision: Remove `NeedsMount: true` from `agent-browser` and `ast-grep`

**Choice:** With bind-mount gone, both kits become capability-free. `NeedsMount` drives SYS_ADMIN `--cap-add`; without mount operations it's not needed.

**Rationale:** Smaller capability surface per kit. `AnyNeedsMount` in the container layer continues to work unchanged — `cx` and `rtk` still set it, so containers that include them still get SYS_ADMIN.

## Risks / Trade-offs

- **[Risk]** Users running `claude` via an execution path that skips the wrapper (e.g. a `bash` shell instead of `zsh`, or a hardcoded absolute path `/usr/local/bin/claude`) miss kit skills. → **Mitigation:** Install the wrapper in both `.zshrc` and `.bashrc` setup, consistent with how other shell config is handled. Document the `/opt/asylum-skills` path in `assets/asylum-reference.md` so an advanced user can add the flag manually if needed.

- **[Risk]** A future Claude Code release changes or removes `--add-dir` semantics. → **Mitigation:** The mechanism is a thin layer on top of a documented, officially supported flag. If it changes, the failure mode is "skills not loaded" — no corruption, no pollution. The broken case is observable and fixable.

- **[Risk]** Previous asylum users already have empty `agent-browser/` and `ast-grep/` directories in their host `~/.claude/skills/`. Those are not cleaned up. → **Mitigation:** CHANGELOG note instructing users to remove them manually. Not worth building a cleanup routine.

- **[Risk]** A different kit (`cx`, `rtk`, or future ones) might try to hand off skills but not declare `ProvidesSkills: true`. → **Mitigation:** None required — those kits don't provide Claude skills, they mount other artifacts. The bool is additive and declarative; kits opt in.

- **[Trade-off]** The agent interface changes. All five agent implementations need to update their `Command` signature. This is a mechanical, one-time cost. The options struct leaves room for future additions without further interface churn.

- **[Trade-off]** `/opt/asylum-skills` is a hardcoded convention. If a kit ever wanted a different skills root (e.g. plugins grouped by kit), it would need to conform. Fine for the current design; revisit if the convention proves too rigid.
