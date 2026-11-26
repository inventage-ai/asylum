# Tasks

## 1. Scope declarations

- [x] 1.1 Add `kit.Scope` (`ScopeAny`, `ScopePersonal`, `ScopeGlobal`) and the `Scope`/`OptionScopes` fields on `kit.Kit`. Set `Scope: ScopePersonal` on `docker`, `dropshare`, `iterm`, `cx` and `rtk`. Verify with `go build ./...`.
- [x] 1.2 Add a `scope` tag to every field of `Config`, `AgentConfig` and `KitConfig`, following the tables in `specs/config-layer-scopes/spec.md`. Add a reflection test that fails on an untagged field (skipping `yaml:"-"`) or an unknown tag value. Verify the test passes, and that it fails when one tag is removed.

## 2. Filtering in Load

- [x] 2.1 Build the scope table from the tags and the kit declarations, with kit declarations taking precedence. Add the node filter that removes out-of-scope entries per layer and records a `Rejection` (file, line, path, YAML value, scope). Keep personal-kit nodes that contain exactly `disabled: true`. Verify with table-driven tests for the `config-layer-scopes` scenarios: any/personal/global per layer, a personal kit enabled and disabled in `.asylum`, ssh isolation, credentials, a mixed file.
- [x] 2.2 Switch `Load` to the order resolve → (local layer) git tracking check → migrate → parse node → filter → decode. Add `Config.Rejected` (`yaml:"-"`), concatenate it in `Merge`, and delete the hand-written `env-from-host` and `seed` guards. Verify with `TestLoad…` tests for the env-from-host, seed and browser-open `schemes` deltas, including "Project layer ignored" and "Layers accumulate" via `.asylum.local`.
- [x] 2.3 Implement the tracking check (`git ls-files -z` on both local filenames, only when the file exists; git errors count as untracked). Verify with tests in a temp git repo: committed, staged, subdirectory of a repo, untracked, not a repo.

## 3. Git exclude

- [x] 3.1 Add the exclude step (`check-ignore` on both names → `rev-parse --git-path info/exclude` → append the names not yet ignored, create `info/`, log one line). Call it once from the main run path. Verify with tests: not ignored, already ignored via `.gitignore`, only the alias missing, missing `info/`, project subdirectory, linked worktree writes to the shared repository, not a repo.

## 4. Startup prompt

- [x] 4.1 Add the rejection report in `cmd/asylum`. On a terminal, show a `tui.Select` prompt ("Start without these options" / "Abort", abort exits non-zero) after the first `Load` in the main run path, before the wizard. Without a terminal, log per-rejection warnings and continue. Verify by checking which subcommands reach this path, then running `asylum` by hand in a scratch project with `volumes` and `kits.docker` in `.asylum`, once with and once without a TTY.
- [x] 4.2 Remember the decision as a SHA-256 over sorted `file\0path\0value` in `~/.asylum/projects/<container>/rejected-config`. While it matches, log one line listing the ignored paths. Verify with unit tests: hash stability, unchanged set skips the prompt, an added entry prompts again.
- [x] 4.3 Log the same per-rejection warnings, without a prompt, in `self-update` and `asylum config`. Verify by running `asylum config` with `release-channel` in `.asylum.local` and checking the warning.

## 5. Repository, docs and changelog

- [x] 5.1 Move the `volumes` entry from this repository's `.asylum` to `.asylum.local`. Verify that `git status` shows `.asylum.local` as ignored and that the asylum container still mounts `~/.asylum/bin`.
- [x] 5.2 Update the docs:
  - Add a "Where options are accepted" section with the scope table to `docs/configuration/index.md`.
  - Move the `volumes` examples to `.asylum.local` in `docs/concepts/mounts.md` and `docs/configuration/index.md`.
  - Describe scopes, the git guard, and the #47 limitation in `docs/concepts/security-model.md`.
  - Note the personal scope on the `docker`, `dropshare`, `iterm`, `cx`, `rtk`, `ssh` (isolation) and `browser-open` (schemes) kit pages.

  Verify with `mkdocs build --strict`.
- [x] 5.3 Add CHANGELOG entries under **Unreleased**:
  - **Changed:** the options now restricted per file (breaking).
  - **Added:** the refusal of a tracked `.asylum.local` and the `info/exclude` entry.

  Verify by reading the section.

## 6. Verification

- [x] 6.1 Run `go vet ./...`, `go test ./...` and `openspec validate scope-config-options --strict`; verify all pass.
- [x] 6.2 End to end in a scratch git repo:
  - A committed `.asylum.local` refuses to start.
  - A clean repository gets the `info/exclude` entries.
  - `.asylum` with `kits.ssh.isolation: shared` and a volume shows the prompt once, does not mount `~/.ssh`, and does not prompt again on the next start.

  Verify by observation. (Done with the asylum binary up to the Docker check, in a scratch repo with a temporary HOME. That `~/.ssh` is not mounted was checked at the config level by `TestFilterScopes`/`TestLoadScopes`, not in a running container, because no image was built in the sandbox.)

## Workflow follow-up

- Coordinate with PR #50: whichever merges second gives host-mcp `Scope: ScopePersonal` and `OptionScopes{"servers": ScopeGlobal}` and removes its own project-layer handling.
- Backlog: remove the `SYS_ADMIN` need of cx and rtk, check whether rtk's `settings.json` edit reaches the host's `~/.claude` in shared mode, then make cx and rtk `any` again.
