## Why

Today asylum auto-resumes the previous agent session whenever one exists, and requires `-n/--new` to start fresh. In practice this is the inverse of what most users want: each `asylum` invocation is usually a new task, and a silent resume can leak prior context into unrelated work. Underlying agents (Claude, Codex, Gemini, Copilot, Pi) already expose their own `--continue` / `--resume` flows, so asylum's auto-resume is duplicative and harder to reason about.

We want the default to start a new session and let users opt into resume explicitly via the agent's own flags. Existing users rely on the current behaviour, so the switch must be announced and reversible.

## What Changes

- **BREAKING**: `asylum` (agent mode) SHALL start a new agent session by default. Auto-resume based on session markers is removed from the default path.
- `--continue` and `--resume` are forwarded verbatim to the underlying agent as passthrough args, so resume becomes the agent's responsibility.
- `-n/--new` is retained as a recognised flag but is a no-op (kept for backwards compatibility in user scripts/aliases). Help text marks it deprecated.
- A new config key (e.g. `default-resume: true`) lets users restore the old auto-resume behaviour.
- On the first asylum launch **after upgrading to the version that ships this change**, asylum SHALL detect that this is an existing installation and present a one-time TUI dialog explaining the behaviour change. The dialog offers: (a) keep new default, (b) opt back into auto-resume by writing `default-resume: true` to `~/.asylum/config.yaml`.
- New installations (no `~/.asylum/` state prior to this run) SHALL NOT see the dialog — they get the new default silently.
- The dialog is shown at most once; a marker in `state.json` records that the user has been notified.

## Capabilities

### New Capabilities
- `resume-migration-prompt`: one-time upgrade dialog that informs existing users of the new default-new-session behaviour and lets them opt back into auto-resume.

### Modified Capabilities
- `cli-dispatch`: `-n/--new` becomes a recognised no-op; `--continue` and `--resume` become recognised flags that are forwarded to the agent as passthrough args.
- `container-exec`: default agent invocation no longer resumes; resume happens only when the user passes `--continue`/`--resume` (forwarded to the agent) or has opted into `default-resume: true`.
- `config-loading`: new optional `default-resume` boolean in the layered YAML config.
- `kit-state-tracking`: `state.json` gains a flag recording that the resume-migration prompt has been shown.

## Impact

- Affected code:
  - `cmd/asylum/main.go` — flag parsing (`--new`, `--continue`, `--resume`), session decision logic, dialog dispatch.
  - `internal/container/container.go` — `ExecOpts.NewSession` semantics flip; resume gated on config + explicit flag rather than `HasSession`.
  - `internal/config/config.go` — new `DefaultResume bool` field, layered merge.
  - `internal/config/state.go` — new `ResumeMigrationPromptShown bool` (or similar) field.
  - `internal/agent/*.go` — no API changes; agents already accept resume-style passthrough args, but the call site no longer auto-derives `resume=true` from `HasSession`.
  - New package or module under `internal/firstrun/` (or sibling) for the migration dialog.
- Tests: `agent_test.go`, `container_test.go`, `config_test.go`, and any e2e tests asserting auto-resume must be updated.
- User-visible: CHANGELOG entry under **Changed** (breaking-default-flip) and **Added** (migration dialog, `default-resume` config). Help text updated.
- No Docker image or kit changes.
