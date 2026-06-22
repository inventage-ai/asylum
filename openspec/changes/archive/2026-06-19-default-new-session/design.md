## Context

Today the agent-mode invocation in `cmd/asylum/main.go` derives `newSession := flags.New` and passes it through to `container.ExecArgs`. Inside `container.agentCommand`, the call `resume := err == nil && !opts.NewSession && opts.Agent.HasSession(configDir, opts.ProjectDir)` decides whether to inject the agent-native resume flag (`--continue` for Claude/Pi, `--resume` for Gemini/Copilot, `resume --last` for Codex).

Three properties of the current design constrain the rewrite:

1. Each agent has its own resume flag and `HasSession` heuristic. We do not want to duplicate that knowledge.
2. The session decision is taken *after* the container is up, inside `container.ExecArgs`. The dialog must run *before* the container is created (it gates the user-facing behaviour of this invocation).
3. State and config are read from `~/.asylum/`. A new user has no such directory; an existing user has at least `state.json` (or kit-tracking state) from any prior run.

This change is small in code surface but breaking in behaviour. The migration prompt is the part that absorbs the breakage.

## Goals / Non-Goals

**Goals:**
- Make "start a new session" the default for every `asylum` invocation.
- Make `--continue` / `--resume` the explicit, agent-native way to resume.
- Keep `-n/--new` parseable as a no-op so existing aliases and docs do not error.
- Inform existing users exactly once and give them a one-keystroke way back to the old behaviour.
- Skip the dialog entirely for new users, non-interactive runs, and non-agent subcommands.

**Non-Goals:**
- Changing each agent's resume semantics (each agent still owns its own `--continue` / `--resume` argv).
- Removing `HasSession` from the `Agent` interface — it stays for the opt-in `default-resume: true` path and for future use (e.g. resume picker UIs).
- Adding a migration that rewrites project-level config. The opt-back-in is written to the global layer (`~/.asylum/config.yaml`) only.
- A second dialog or reminder if the user dismisses the prompt without choosing.

## Decisions

### Decision 1: `-n/--new` becomes a no-op, not removed

We could remove the flag, but that breaks user scripts and aliases. Keeping it as a recognised no-op costs almost nothing (one case branch in the parser, a deprecation note in `--help`). After two or three releases we can re-evaluate removal.

**Alternative considered:** Hard-remove `-n`. Rejected — gratuitous breakage with no benefit, since the flag has no other behaviour to repurpose.

### Decision 2: `--continue` and `--resume` are pure passthrough

Asylum recognises them in the parser only so they are not rejected as unknown flags. The handler appends them to `extraArgs` without translation. The agent's existing `Command(resume bool, extraArgs []string, opts)` interface continues to work — `resume` is `false` by default, `extraArgs` carries `--continue`/`--resume` to the agent.

This sidesteps the agent-specific divergence (Codex uses `codex resume --last`, not a flag) — when a user explicitly types `--continue`, that maps cleanly onto Claude/Pi but is not what Codex expects. We accept that:
- For Claude/Pi/Copilot/Gemini, `--continue`/`--resume` work directly because each agent accepts them.
- For Codex, `--continue` will produce a Codex error. That is acceptable — Codex users know to use `codex resume` semantics, and the asylum-side fix is documented (use `default-resume: true` if you want auto-resume, or pass Codex-native resume args).

**Alternative considered:** Translate `--continue` per-agent. Rejected — duplicates agent-specific knowledge and re-creates the very abstraction we are removing.

### Decision 3: Existing-user detection probes `~/.asylum/agents/`

`firstrun.IsExistingInstall` checks for `<home>/.asylum/agents/` — the same signal `firstrun.Run` already uses to recognise repeat users. That directory is materialised lazily by `container.EnsureAgentConfig` when an agent's config is first set up; it is *not* touched by the early `WriteDefaults` that writes `~/.asylum/config.yaml` on every invocation. So the probe is robust to initialisation order: even if some code path above us creates `~/.asylum/` or `~/.asylum/config.yaml` first, the agents directory still only exists for users who have actually run an asylum session before.

The first draft of this design probed `~/.asylum/` itself and required us to snapshot the stat result before any other code wrote into the directory. That ordering constraint was fragile and was flagged in adversarial review as already broken by the eager default-config write in `cmd/asylum/main.go`. Switching to the agents-dir signal removes the constraint entirely.

We do not use `state.json` alone as the signal — it has only existed since a recent release, and some long-time users may not have it yet. The directory test is robust across versions.

**Alternative considered:** Track an `installed_at` timestamp. Rejected — adds state we have to write, and we still need a directory test for the bootstrap case.

**Alternative considered:** Probe `~/.asylum/` directly. Rejected — eagerly created on every run, so the order-dependence is a latent bug.

### Decision 4: Dialog is shown only in agent mode with a TTY

`asylum cleanup`, `asylum version`, `asylum self-update`, `asylum config`, `asylum shell`, and `asylum run <cmd>` all skip the dialog. The dialog gates default session behaviour, which only matters in agent mode. Suppressing it elsewhere also prevents corrupting non-interactive workflows (CI, piped invocations).

We suppress without marking the flag, so the first interactive agent run still sees the dialog.

### Decision 5: Dialog writes to the global config layer

Opt-in to legacy behaviour writes `default-resume: true` to `~/.asylum/config.yaml` using a load → mutate → save round-trip. Other keys are preserved. We do not write per-project config; the user expressed a global preference.

If `~/.asylum/config.yaml` does not exist, it is created. If it exists but is unparseable, the dialog surfaces an error and asks the user to edit manually (we do not overwrite a broken file).

### Decision 6: State field name

Add `ResumeMigrationPromptShown bool` to `internal/config/state.go`'s `State` struct, JSON tag `"resume_migration_prompt_shown,omitempty"`. The `omitempty` keeps older `state.json` files clean (zero value is `false`).

### Decision 7: New users get the flag pre-set to "shown"

After running the first-run wizard for a brand-new user, we immediately write `ResumeMigrationPromptShown: true` to `state.json`. That way, when this same install later upgrades to a future version (hypothetical), the dialog stays suppressed for the original-new-user. The dialog is for *the transition from pre-change behaviour to post-change behaviour*, not for every fresh install going forward.

## Risks / Trade-offs

- **[Codex users surprised by `--continue` failure]** → Documented in CHANGELOG. Codex has no `--continue` flag; asylum cannot fix that. Mitigation: dialog text and help text mention that `--continue`/`--resume` are agent-native passthrough.
- **[Users with custom wrappers depending on auto-resume]** → Dialog plus `default-resume: true` config key gives them a one-line fix.
- **[Detection misfires on partially-populated `~/.asylum/`]** → The directory test is conservative: any existing dir counts as "existing user". Worst case is an existing user who never actually ran a session sees the dialog once. Acceptable.
- **[Race between dialog and first-run wizard]** → Order in `runDefault`: existing-user detection (stat) → first-run wizard if new → migration dialog if existing-and-not-yet-shown → rest of flow. The two flows are mutually exclusive by construction.
- **[Dialog shown twice if user Ctrl-C'd]** → If the user kills asylum after the dialog renders but before we persist the flag, they see it again. Acceptable — preferable to writing the flag before the user makes a choice.

## Migration Plan

1. Land code + specs + changelog in a single release.
2. CHANGELOG entry under **Changed** clearly tagged as breaking-default-flip, and under **Added** for the dialog and `default-resume` key.
3. No automated migration script. The dialog is the migration.
4. Rollback: revert the release. Users who opted into `default-resume: true` keep that config key; it is a no-op on a reverted asylum (unknown YAML keys are tolerated by current loaders).

## Open Questions

- Should the dialog also offer a "remind me later" option? **Tentative no** — it complicates state. Users can re-run `asylum` after they decide.
- Wording of the dialog: defer to implementation review.
