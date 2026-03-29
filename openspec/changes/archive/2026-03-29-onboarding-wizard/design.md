## Context

Onboarding prompts currently exist in two places with different patterns:

1. **Config isolation** — inline in `main.go:206`, fires when `AgentIsolation() == ""`, uses `tui.Select`
2. **Kit credentials** — in `firstrun.Run()`, fires only when `~/.asylum/agents/` doesn't exist (first-run gate), uses `tui.MultiSelect`

The first-run gate means v1 migrators (who have `~/.asylum/agents/`) never see the credential prompt. Both prompts should use the same "not yet configured" detection pattern.

The existing TUI components (`tui.Select`, `tui.MultiSelect`) each run as independent bubbletea programs. A wizard needs to wrap multiple steps in a single program to maintain visual continuity.

## Goals / Non-Goals

**Goals:**
- Single wizard TUI that groups all pending onboarding steps with a tab/step indicator
- Both isolation and credentials use "not configured" detection (not first-run gating)
- Wizard scales to future onboarding steps without structural changes
- Completed steps show ✓ in tab bar, current step is highlighted
- Cancel (esc) at any point stops the wizard; completed steps are still applied

**Non-Goals:**
- Back-navigation between wizard steps
- Persistence of partial wizard state across asylum invocations
- Animated transitions between steps
- Wizard for any purpose beyond pre-container-start onboarding

## Decisions

### 1. Single bubbletea program, not sequential separate programs

The wizard runs as one `tea.Program` that manages an internal step index and delegates input to the current step's sub-model (selectModel or multiModel). The tab bar and step content render together in one `View()`.

**Why not sequential programs?** Separate programs would clear the screen between steps, losing the tab bar continuity. A single program keeps the visual context.

### 2. Wizard model composes existing select/multi models internally

The wizard doesn't import or reuse the exported `Select`/`MultiSelect` functions (those call `tea.NewProgram` themselves). Instead, it uses the same internal model structs (`selectModel`, `multiModel`) but manages them within the wizard's own `Update`/`View` cycle.

**Why?** The existing functions create and run their own programs. The wizard needs to be the single program that delegates. This means the internal model types need to be accessible within the `tui` package, which they already are (same package).

### 3. Wizard returns results for all steps, caller applies them

The wizard returns `[]StepResult` — one per step. The caller (main.go onboarding function) interprets the results and writes config. The wizard has no knowledge of isolation levels or credential config.

**Why?** Keeps the wizard generic. The onboarding function in main.go knows what each step means.

### 4. Credential detection moves to "not configured" pattern

Check `cfg.KitCredentialMode(parentKitName) == ""` for each active kit with a non-nil `CredentialFunc`. This fires for first-time users AND v1 migrators, matching the isolation prompt pattern.

**Why?** Same reason the isolation prompt uses this pattern — "not configured" is the right signal, not "first run."

### 5. Onboarding runs once per container lifecycle (inside `!docker.IsRunning` block)

The wizard fires in the same location as the current isolation prompt — inside the `if !docker.IsRunning(cname)` block. This means it only fires when starting a fresh container, not when attaching to a running one.

**Why?** Matches existing behavior. Onboarding config only matters at container start.

### 6. firstrun.Run() stripped of credential logic

`firstrun.Run()` no longer handles credentials. It retains the first-run detection (`~/.asylum/agents/` check) as a shell for any future first-run-only tasks, but its credential prompting and `SetKitCredentials` logic move to the main.go onboarding flow.

## Risks / Trade-offs

**[Wizard model is more complex than sequential prompts]** → Mitigation: the wizard delegates to the same model logic used by Select/MultiSelect, just composed differently. Testing can verify step transitions.

**[Single-step wizard feels over-engineered]** → Mitigation: if only one step is needed, the tab bar still renders but with just one tab — this is clean, not confusing. The visual overhead is one line.

**[Internal model coupling within tui package]** → Mitigation: selectModel and multiModel are already in the same package. No export boundary changes needed.
