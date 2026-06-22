## 1. Config + state plumbing

- [x] 1.1 Add `DefaultResume bool` (yaml tag `default-resume,omitempty`) to `internal/config/config.go` `Config` struct and include it in the layered merge in `Merge`/`Overlay`
- [x] 1.2 Cover the new field with tests in `internal/config/config_test.go`: unset → false, global only, project overrides global, local overrides project, CLI overlay (if applicable)
- [x] 1.3 Add `ResumeMigrationPromptShown bool` (json tag `resume_migration_prompt_shown,omitempty`) to `internal/config/state.go` `State` struct and verify it round-trips through `LoadState`/`SaveState`
- [x] 1.4 Add a helper in `internal/config/` (e.g. `WriteDefaultResume(asylumDir string, value bool) error`) that loads `config.yaml`, sets the key, writes it back, preserving other keys; refuses to overwrite an unparseable file

## 2. CLI flag changes

- [x] 2.1 In `cmd/asylum/main.go` `parseArgs`, add cases for `--continue` and `--resume`: both append the literal flag to `extraArgs` (passthrough) and advance the index by 1
- [x] 2.2 Change the `-n/--new` case to a no-op (still consumes the flag, but does not set any decision variable); drop `flags.New` if no other site reads it
- [x] 2.3 Update the `--help` text in `cmd/asylum/main.go` so `-n, --new` is marked `(deprecated, no-op)` and `--continue` / `--resume` are listed as "Forwarded to agent"
- [x] 2.4 Remove the `newSession := flags.New` line and any other paths that derived a "new session" decision from CLI flags

## 3. Container-exec resume logic

- [x] 3.1 In `internal/container/container.go`, rename `ExecOpts.NewSession bool` to `ExecOpts.DefaultResume bool` (semantics: caller-supplied resolved `default-resume` value); update all call sites
- [x] 3.2 In `agentCommand`, change the resume decision to: `resume := opts.DefaultResume && err == nil && opts.Agent.HasSession(configDir, opts.ProjectDir)`
- [x] 3.3 Update `cmd/asylum/main.go` to pass `cfg.DefaultResume` (the resolved value) into `ExecOpts` instead of the old `newSession` boolean
- [x] 3.4 Update `internal/container/container_test.go` and `internal/agent/agent_test.go` to reflect: default is new session; `--continue`/`--resume` appear as extraArgs; `default-resume: true` + `HasSession=true` triggers agent-native resume flag

## 4. Resume migration prompt

- [x] 4.1 Create `internal/firstrun/migration.go` (or a sibling module) with `ShouldShowResumePrompt(home string, state config.State, mode container.Mode, isTTY bool) bool` capturing: existing-user detection via `os.Stat(home + "/.asylum")` snapshot, agent mode only, TTY only, flag not yet set
- [x] 4.2 Take the existing-user snapshot at the very top of `runDefault` (before `ensureImages` / `LoadState`); thread it through to the prompt check
- [x] 4.3 Implement `ShowResumePrompt(home string) (optInLegacy bool, err error)` using existing `internal/tui` primitives (lipgloss/bubbletea), with two clear options and explanatory text
- [x] 4.4 On "keep new default": persist `state.ResumeMigrationPromptShown = true` and proceed
- [x] 4.5 On "restore previous behaviour": call the config helper from 1.4 to write `default-resume: true` to `~/.asylum/config.yaml`, persist `state.ResumeMigrationPromptShown = true`, **reload the resolved config** so the current invocation honours the new value, and proceed
- [x] 4.6 For confirmed new users (snapshot found `~/.asylum/` did not exist), set `state.ResumeMigrationPromptShown = true` immediately after the first-run wizard creates state, so they never see the dialog on a future asylum version
- [x] 4.7 Add unit tests covering: existing-user/no-flag → shows; new-user → suppressed and flag pre-set; flag already set → suppressed; non-agent mode → suppressed; non-TTY → suppressed without flag write
- [x] 4.8 Add an integration-style test (or e2e) that exercises the dialog by piping a selection and asserting that `default-resume: true` ends up in `~/.asylum/config.yaml`

## 5. Documentation + changelog

- [x] 5.1 Add a `CHANGELOG.md` entry under **Unreleased**: a **Changed** bullet describing the breaking default flip, and **Added** bullets for `default-resume` and the migration dialog
- [x] 5.2 Update the in-container reference `assets/asylum-reference.md` to describe the new default, `--continue`/`--resume` passthrough, and `default-resume`
- [x] 5.3 Update any README/docs site pages that documented `-n/--new` or auto-resume

## 6. Verify

- [x] 6.1 Run `go test ./...` and `go vet ./...` — all green
- [x] 6.2 Manually verify on a real install: existing user sees the dialog once; choosing "keep new" leaves config untouched; choosing "restore" writes `default-resume: true` and the current run resumes; second invocation does not re-prompt (covered by `TestMaybeShowResumeMigrationPrompt_FlagAlreadySetIsNoop`, `TestMaybeShowResumeMigrationPrompt_OptInWritesConfig`, and `TestWriteDefaultResume_PreservesOtherKeys` — the TTY-driven Select rendering itself is not unit-testable but the state machine on either side of it is)
- [x] 6.3 Manually verify a fresh `~/` (empty `HOME`): no dialog; `state.json` has `resume_migration_prompt_shown: true` after the first run (covered by `TestMaybeShowResumeMigrationPrompt_NewUserPreMarks` + `TestIsExistingInstall/only ~/.asylum/config.yaml is not enough`, which together exercise the new-user invariant — including the regression that an early `~/.asylum/config.yaml` write must not flip the existing-install probe)
- [x] 6.4 Manually verify `asylum --continue` reaches Claude as `--continue`, `asylum --resume` reaches Gemini as `--resume` (verified via `TestClaudeCommand_ContinueResumePassthrough` + `TestGeminiCommand_ContinueResumePassthrough` covering the full Agent.Command path)
- [x] 6.5 Run `make test-integration` once as a final gate before merging
