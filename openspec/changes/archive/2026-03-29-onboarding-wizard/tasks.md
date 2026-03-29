## 1. Wizard TUI Component

- [x] 1.1 Add `WizardStep`, `StepKind`, `StepResult` types to `internal/tui/wizard.go`
- [x] 1.2 Implement the wizard bubbletea model: tab bar rendering, step delegation to selectModel/multiModel, step transitions on confirm, cancel handling
- [x] 1.3 Add `Wizard` function that creates the program, runs it, and returns `[]StepResult` (with non-interactive fallback returning defaults)
- [x] 1.4 Write tests for wizard step transitions, cancel behavior, and result collection

## 2. Unified Onboarding Function

- [x] 2.1 Create `runOnboarding()` function in `cmd/asylum/main.go` that collects pending steps (isolation if unconfigured, credentials if active kit has CredentialFunc but no credentials config)
- [x] 2.2 Build wizard steps from pending items and call `tui.Wizard`
- [x] 2.3 Apply wizard results: write isolation level and/or credential config to `~/.asylum/config.yaml`, update in-memory config
- [x] 2.4 Replace inline `promptConfigIsolation()` call with `runOnboarding()` in the `!docker.IsRunning` block

## 3. Remove Old Prompt Code

- [x] 3.1 Remove credential prompting from `internal/firstrun/firstrun.go` (keep first-run detection shell, remove `promptCredentials`, `CredentialCapableKits` usage)
- [x] 3.2 Remove `promptConfigIsolation()` function from `cmd/asylum/main.go`
- [x] 3.3 Update `internal/firstrun/firstrun_test.go` for the stripped-down firstrun

## 4. Cleanup

- [x] 4.1 Add changelog entry under Unreleased
