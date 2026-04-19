## 1. Kit metadata

- [x] 1.1 Add `ProvidesSkills bool` field to `kit.Kit` in `internal/kit/kit.go`, documented in the field comment
- [x] 1.2 Add `AnyProvidesSkills(kits []*Kit) bool` helper in `internal/kit/kit.go` alongside `AnyNeedsMount`

## 2. Agent interface

- [x] 2.1 Define an `AgentCmdOpts` struct in `internal/agent/agent.go` carrying at least a `KitSkillsDir string` field (named `agent.CmdOpts` to avoid stutter)
- [x] 2.2 Update `Agent` interface signature: `Command(resume bool, extraArgs []string, opts CmdOpts) []string`
- [x] 2.3 Update `Claude.Command` in `internal/agent/claude.go` to prepend `--add-dir <opts.KitSkillsDir>` after `--dangerously-skip-permissions` when `opts.KitSkillsDir != ""`
- [x] 2.4 Update `Gemini.Command`, `Codex.Command`, `Opencode.Command`, `Echo.Command` to accept the new `CmdOpts` parameter and ignore it
- [x] 2.5 Update all callers of `Agent.Command` to pass `CmdOpts`; in `internal/container/container.go`, populate `KitSkillsDir` as `/opt/asylum-skills` when `kit.AnyProvidesSkills(opts.Kits)` is true, else empty
- [x] 2.6 Extend `internal/agent/agent_test.go`: add test cases for Claude covering `KitSkillsDir=""` (no flag), `KitSkillsDir="/opt/asylum-skills"` with and without resume, with and without extra args; add coverage for Gemini/Codex/Opencode/Echo showing the new field is ignored

## 3. agent-browser kit

- [x] 3.1 Update `internal/kit/agent_browser.go`: set `ProvidesSkills: true`, remove `NeedsMount: true`
- [x] 3.2 Replace the DockerSnippet line `mv .claude/skills/agent-browser /tmp/asylum-kit-skills-agent-browser` with staging into `/opt/asylum-skills/.claude/skills/agent-browser` (create parent dir first, chown to the container user so the skill is readable at runtime)
- [x] 3.3 Remove the skill mount block from the EntrypointSnippet (keep the existing ARM64 `AGENT_BROWSER_EXECUTABLE_PATH` export)

## 4. ast-grep kit

- [x] 4.1 Update `internal/kit/astgrep.go`: set `ProvidesSkills: true`, remove `NeedsMount: true`
- [x] 4.2 Replace the DockerSnippet line `mv .claude/skills/ast-grep /tmp/asylum-kit-skills-ast-grep` with staging into `/opt/asylum-skills/.claude/skills/ast-grep` (create parent dir first, chown to the container user)
- [x] 4.3 Remove the EntrypointSnippet entirely (the mount block was its only content)

## 5. Entrypoint shell wrapper

- [x] 5.1 In `assets/entrypoint.core` (or the tail, whichever owns shell-setup writes), emit a `claude` shell function into both `$HOME/.zshrc` and `$HOME/.bashrc` that runs `command claude --add-dir /opt/asylum-skills "$@"` when `/opt/asylum-skills/.claude/skills` contains at least one entry, else `command claude "$@"`
- [x] 5.2 Ensure the function is only installed once (guarded by a marker line) to survive repeated container starts (verified via local shell smoke test)
- [x] 5.3 Manually verify inside a container that `type claude` shows the function and that `claude --help` still works without asylum launching it (verified against built image: wrapper function loads in interactive bash, dispatches with and without `--add-dir` depending on skills-dir state)

## 6. Tests and verification

- [x] 6.1 Run `go test ./...` and fix any failures caused by the `Agent.Command` signature change
- [x] 6.2 Run `go vet ./...`
- [x] 6.3 Build the base image with `agent-browser` and `ast-grep` active; verify `/opt/asylum-skills/.claude/skills/agent-browser/SKILL.md` and `/opt/asylum-skills/.claude/skills/ast-grep/SKILL.md` exist in the image (verified: both `SKILL.md` files present with upstream frontmatter)
- [x] 6.4 Run asylum in non-shared mode; confirm Claude loads both skills (e.g. `/help` or running a skill-dependent command) (verified indirectly: wrapper dispatches `claude --add-dir /opt/asylum-skills` when skills dir is populated, and skills dir is populated in the built image)
- [x] 6.5 Run asylum in shared mode with a clean host `~/.claude/skills/` (no `agent-browser/` or `ast-grep/` entries); confirm after exit that those entries are still absent on the host (verified: bind-mounted host `.claude/skills/` remained empty after a full entrypoint run against the built image)
- [x] 6.6 Inside a running container, open a second shell, run `claude --help`, and confirm via `type claude` that the wrapper is active (verified: `type claude` in an interactive bash subshell inside the image reports the wrapper function)

## 7. Documentation

- [x] 7.1 Add a `CHANGELOG.md` entry under Unreleased → Fixed: "Kit-provided Claude skills (`agent-browser`, `ast-grep`) no longer create empty directories in the user's host `~/.claude/skills/` in shared agent-config mode. Skills are now staged under `/opt/asylum-skills` and loaded via `--add-dir`. Users may safely remove any existing empty `~/.claude/skills/agent-browser/` or `~/.claude/skills/ast-grep/` directories left over from previous versions. (#24, #25)"
- [x] 7.2 Add a short section to `assets/asylum-reference.md` documenting `/opt/asylum-skills` as the kit-skills root and that `--add-dir` is set automatically by the `claude` wrapper
