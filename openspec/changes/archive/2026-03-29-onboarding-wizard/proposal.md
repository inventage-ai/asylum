## Why

Onboarding prompts (config isolation, credentials) fire as separate, disconnected prompts with different detection logic in different locations. The credential prompt is gated by first-run detection (`~/.asylum/agents/` existence), so v1 migrators never see it. Grouping all pending onboarding questions into a single wizard flow with a step indicator fixes the migration bug and provides a cohesive experience that scales to future onboarding steps.

## What Changes

- Add a `Wizard` TUI component that wraps multiple Select/MultiSelect steps into a single flow with a tab bar showing completed (✓), current, and upcoming steps
- Replace the inline `promptConfigIsolation()` call in main.go and the `firstrun.Run()` credential prompt with a unified onboarding function that collects all pending steps and runs the wizard
- Change credential prompt detection from first-run gating (`~/.asylum/agents/`) to "not yet configured" gating (same pattern as config isolation)
- Remove credential-specific logic from `internal/firstrun/firstrun.go`

## Capabilities

### New Capabilities
- `onboarding-wizard`: Wizard TUI component and unified onboarding flow that collects pending configuration steps and presents them as a multi-step wizard with tab navigation

### Modified Capabilities
- `tui-prompts`: Add Wizard function to the TUI package alongside existing Select and MultiSelect
- `config-isolation`: Isolation prompt moves from inline code into the wizard flow (behavior unchanged, presentation changes)
- `first-run-onboarding`: Credential prompt moves from first-run-gated to "not configured" detection, removing dependency on `~/.asylum/agents/` existence

## Impact

- `internal/tui/wizard.go` — new file
- `cmd/asylum/main.go` — replace inline isolation prompt + firstrun credential call with unified onboarding wizard
- `internal/firstrun/firstrun.go` — remove credential prompting
- `internal/firstrun/firstrun_test.go` — update tests
- No new dependencies (uses existing bubbletea/lipgloss)
