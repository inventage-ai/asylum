## Context

The current first-run flow in `cmd/asylum/main.go` does this:

```
parseArgs → SyncNewKits → WriteDefaults → firstrun.Run (no-op) → config.Load
  → ensureImages (slow, silent build of base + project)
  → !IsRunning ⇒ runOnboarding (isolation + credentials wizard)
  → EnsureAgentConfig
  → RunArgs ⇒ ssh kit MountFunc ⇒ ensureSSHKey (raw ssh-keygen output + key + tip)
  → RunDetached → ExecArgs
```

Two distinct UX problems:

1. **Ordering**: image build runs before the wizard, so a new user waits in silence for minutes before being asked anything. The wizard only covers isolation and credentials — image-shaping inputs (agents, kits) are silently defaulted, so a user who wants a different agent has no in-flow chance to say so.
2. **Spam at launch**: `ssh-keygen` inherits stdout/stderr (randomart) and `ensureSSHKey` prints the public key plus a "Add this key to your Git host" sentence just before the agent session starts. Same content also lives in `asylum-reference.md`, which the in-container agent reads.

The `firstrun` package is currently an empty shell (`firstrun.go` is documented as "remains as a shell for any future first-run-only tasks"). It is the natural home for the expanded wizard.

The wizard infrastructure (`internal/tui/wizard.go`, `tui.WizardStep`, `StepSelect`/`StepMultiSelect`) already supports everything we need.

## Goals / Non-Goals

**Goals:**

- Move all first-run interaction *before* `ensureImages` so the user is never staring at silence on a fresh install.
- Add agents and top-level kits to the wizard so first-run choices actually shape the image being built.
- Eliminate `ssh-keygen` randomart and the redundant key-instructions block; replace with one line that points at the in-container reference.
- Print a single context line before any actual image build so non-first-run rebuilds are also less mysterious.
- Default-preserving: pressing enter through every step yields today's silent defaults.

**Non-Goals:**

- Sub-kit selection (e.g. `java/maven`, `python/pip`) in the wizard. Keep those config-file territory.
- A re-runnable `asylum config --wizard` command. The existing `asylum config` subcommand covers post-first-run edits; re-walking the full wizard later isn't worth a dedicated entry point yet.
- Changes to `SyncNewKits` semantics. It continues to fire for existing users on upgrade.
- Multi-agent default-picker ordering heuristics beyond "claude if selected, otherwise first picked".
- Hiding the `echo` agent from the registry — only from the picker UI. It stays selectable via `-a echo` for tests.
- Any change to non-interactive behavior. No TTY ⇒ wizard skipped, today's defaults applied.

## Decisions

### Detection signal: missing `~/.asylum/config.yaml`, not missing `~/.asylum/agents/`

The existing `firstrun.IsExistingInstall` probes `~/.asylum/agents/`. That directory is materialized by `EnsureAgentConfig`, which runs *after* `ensureImages` — so the signal works for the resume-migration prompt (which only needs to know "existed before this run"), but it's an indirect proxy.

For wizard gating we want "has the user ever seen our config write before?". The direct signal is the existence of `~/.asylum/config.yaml`. `WriteDefaults` already short-circuits when it exists. We capture the answer once at startup, *before* `WriteDefaults` runs:

```go
isFirstRun := !fileExists(filepath.Join(home, ".asylum", "config.yaml"))
```

**Why not keep the `agents/` probe?** It conflates "first run" with "agent config never seeded", which can drift if we ever pre-seed agent dirs for other reasons. The config.yaml signal is exactly what we mean.

`firstrun.IsExistingInstall` stays for the resume-migration prompt — its semantics (any agents dir at all) are intentionally different.

### Wizard ownership in `internal/firstrun/`

Today the wizard is built inline in `cmd/asylum/main.go:runOnboarding` (~130 lines). With agents + kits added, that function would balloon further. Moving it into the package gives us:

- A single seam (`firstrun.Run(ctx)`) called once from main.
- Per-step builders that can be unit-tested without `main.go` mocking.
- A natural place for the existing `migration.go` (resume-migration prompt) to coexist.

Proposed package layout:

```
internal/firstrun/
  firstrun.go    Run(ctx Context) — orchestrator, decides what to ask
  wizard.go      buildSteps(...) — emits []tui.WizardStep + appliers
  apply.go       persistence helpers (writes to ~/.asylum/config.yaml)
  migration.go   (existing) resume-migration prompt
```

`Context` carries the inputs the wizard needs: home, projectDir, registered agents, registered kits, current `*config.Config`. The orchestrator decides per-step whether to include it based on first-run state + "is this value already set?".

### Wizard step composition

Six potential steps; each fires only when needed:

| # | Step           | Trigger                                                                 | Affects |
|---|----------------|-------------------------------------------------------------------------|---------|
| 1 | Welcome banner | First run only                                                          | none (info) |
| 2 | Agents         | First run AND `cfg.Agent == "" && cfg.Agents == nil`                    | image   |
| 3 | Default agent  | Step 2 produced >1 selection                                            | runtime |
| 4 | Kits           | First run AND no kits explicitly set in config                          | image   |
| 5 | Isolation      | Active agent is claude AND `agents.claude.config == ""`                 | runtime |
| 6 | Credentials    | Any active kit has `CredentialFunc` AND parent kit credentials unset    | runtime |

Steps 5 and 6 are the existing onboarding-wizard steps unchanged. The "first run AND X" gating in 2/4 means a user who deletes their `config.yaml` and re-runs (effectively a first run) will see them again — that's fine.

**Default agent step kept conditional, not always-shown.** If the user picks one agent, there's nothing to ask. Avoids an empty "default agent" prompt with one option.

### Two-phase wizard with a reload between phases

The wizard runs in two phases driven by separate `tui.Wizard` calls:

1. **Image-shaping phase (first-run only)** — agents, default-agent, kits. After the user confirms, the wizard writes a complete config from scratch via `WriteConfig`, then invokes the caller-supplied `Reload` callback to re-load the merged config and re-resolve the active kit set.
2. **Runtime phase (any run)** — isolation, credentials. Step inclusion is gated on the **post-reload** state. If phase 1 dropped Claude from the agent picker, `activeAgentIsClaude` returns false and the isolation step is skipped entirely — no stray `agents.claude.config:` write that would resurrect a deselected agent. The credential step builds from `kit.Resolve` applied to the just-written config, so a deselected credential-capable kit never appears in its options. The credential applier also gates its `SetKitCredentials("false")` write on `KitActive(parent)`, so even if a stale credential prompt somehow surfaced, declining it cannot inject an active kit entry under a commented kit.

This was necessary to fix two cross-phase bugs in the initial single-pass design:
- The runtime phase used the pre-wizard `cfg`, where `cfg.Agent == ""` always defaulted to Claude — so a user picking Gemini-only still got asked about Claude isolation, and a "yes" answer wrote an active `claude:` block.
- `kit.Resolve(nil, ...)` on a fresh install returns all kits, so the credentials step built its option list from kits the user hadn't yet been asked about. Declining a credential prompt then wrote `credentials: false` under that kit, creating an active entry where the kit had been commented.

Alternative considered: single-pass with smarter appliers that defer decisions until the wizard returns. Rejected because the user would still see prompts for steps that should not have fired (e.g. Claude isolation when only Gemini was picked).

### Steps run before `ensureImages`, not after

`firstrun.Run(ctx)` is called immediately after `config.Load` and *before* `ensureImages`. Wizard outputs that affect the image (agents, kits) are written to `~/.asylum/config.yaml`, then `config.Load` is re-invoked to re-resolve the merged config with the new layer baked in.

```
parseArgs → SyncNewKits → WriteDefaults → config.Load
  → firstrun.Run(ctx)  ← wizard here, may mutate config.yaml
  → config.Load (re-resolve if wizard wrote anything)
  → ensureImages
  → EnsureAgentConfig
  → silent ssh + 1-line summary
  → RunDetached → ExecArgs
```

The wizard's `apply` functions already mutate the in-memory `*config.Config` (see today's `runOnboarding`). For image-shaping changes (agents/kits) we need the disk write to be flushed before `ensureImages` reads it; an explicit re-`Load` after the wizard is the cleanest way to guarantee that.

**Alternative considered:** pass the wizard results directly into `ensureImages` without re-loading. Rejected — it bifurcates the config source-of-truth and we'd duplicate kit/agent resolution logic.

### Agent multi-select: list, default, hidden test stub

Wizard step 2 is `tui.StepMultiSelect`. `claude` is pre-checked. `echo` is filtered out of the option list (still resolvable via `-a echo`). Selection writes both `agent: <default>` and `agents:` map entries to config.yaml:

```yaml
agent: claude
agents:
  claude: {}
  gemini: {}
```

When >1 agent is selected, step 3 (`StepSelect`) picks the default. When exactly 1 is selected, step 3 is skipped and the single pick becomes `agent:`.

**Default-default:** if the user accepts step 2 unchanged (claude only), step 3 is skipped, `agent: claude` is written. Identical to today's silent default.

### Kit multi-select: top-level only

Wizard step 4 lists kits where the registry entry has no `/` in `Name` (top-level only) and `Tier != TierAlwaysOn` (always-on kits are not user-toggleable here). `TierDefault` kits are pre-checked, `TierAvailable` unchecked. Selection writes uncommented entries to `kits:` for chosen items, commented entries for unchosen ones — the same comment-vs-active pattern used elsewhere in the config writer.

Sub-kits remain managed in config-file form. A user enabling `java` from the wizard still needs to opt into `java/maven` by editing config (or via a future kit-config UI — out of scope here).

### SSH key generation: silent, one-line notice

Two changes in `internal/kit/ssh.go`:

1. `ssh-keygen` is invoked with stdout/stderr captured into a buffer; on success the buffer is discarded. On non-zero exit the buffer is included in the returned error.
2. The trailing `fmt.Printf("\nSSH public key:\n%s\n", ...)` + "Add this key to your Git hosting provider..." block is replaced with one call to `log.Info` (or `log.Success`):
   ```
   Generated SSH key at ~/.asylum/ssh/id_ed25519.pub — see asylum-reference.md for how to add it to a Git host.
   ```

The "how to add it" content moves to the SSH section of `assets/asylum-reference.md`, which the in-container agent reads. A user who wants the key just runs `cat ~/.asylum/ssh/id_ed25519.pub` (or asks the agent).

**Why not move key generation entirely out of `MountFunc`?** Considered — it would let the key be generated during the wizard phase rather than during `RunArgs`. But the kit lifecycle currently calls `MountFunc` exactly once per asylum invocation, and the file-existence guard in `ensureSSHKey` already makes it idempotent. The win from extracting it is small and the change to the kit interface is large. Punt unless it becomes an actual problem.

### Image build context line

In `internal/image/image.go`, `EnsureBase` and `EnsureProject` currently emit `log.Build("building base image...")` / `"building project image..."` only when they're actually going to build. Add one more line emitted at the same point, but only the *first* time per invocation:

```
log.Info("Building sandbox image — this takes a few minutes the first time, subsequent runs reuse the cache.")
```

Suppressed when both calls are cache hits. Implemented via a once-flag passed down through `ensureImages` or by checking from `main.go` whether either `*Ensure*` reported a build.

**Why a single shared line?** Two separate "this takes a few minutes" messages (one per `Ensure*`) would be worse than today. One line up front covers both phases.

### Echo agent is hidden from picker only

Picker filtering happens in `wizard.go`, not in the agent registry. `agent.Registry()` still returns `echo`; `-a echo` and tests work unchanged.

## Risks / Trade-offs

- **Risk**: a user with a hand-edited but tiny `~/.asylum/config.yaml` (e.g. just `agent: gemini`) is treated as not-first-run, so they never see the wizard.
  - **Mitigation**: this matches today's behavior — they already configured something. The wizard's per-step gating ("is this value set?") handles partial config gracefully.
- **Risk**: re-running `config.Load` after the wizard double-parses YAML.
  - **Mitigation**: negligible cost on a tiny file. Trade-off accepted for source-of-truth simplicity.
- **Risk**: wizard adds latency to first run.
  - **Mitigation**: the wizard *replaces* the post-build wizard that already exists. Net new latency on first run is the agents + kits steps — ~2 keystrokes minimum. The slow-image-build silence is removed in exchange.
- **Risk**: silent `ssh-keygen` failure modes are harder to debug.
  - **Mitigation**: keep the captured output in the returned error so it surfaces via `log.Error`. The success path is the only silent one.
- **Risk**: existing tests in `internal/firstrun/migration_test.go` and `cmd/asylum/main_test.go` couple to `firstrun.IsExistingInstall` and the old `runOnboarding` layout.
  - **Mitigation**: this is a package-level refactor, tests need to follow. Listed in tasks.

## Migration Plan

No data migration. No on-disk format changes. Existing users with a populated `~/.asylum/config.yaml` see no wizard on next run — they're not first-run. The first-run wizard only triggers when `config.yaml` is absent.

Rollback: revert the change. No persistent state introduced.

## Open Questions

None blocking. A future enhancement could add `asylum config --wizard` to re-walk the flow on demand; not in scope here.
