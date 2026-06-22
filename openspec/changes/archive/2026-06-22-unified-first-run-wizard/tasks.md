## 1. firstrun package: wizard ownership

- [x] 1.1 Add `firstrun.IsFirstRun(home)` helper that returns `!fileExists(filepath.Join(home, ".asylum", "config.yaml"))`. Keep `firstrun.IsExistingInstall` as-is for the resume-migration prompt.
- [x] 1.2 Capture the first-run flag in `cmd/asylum/main.go` *before* `config.WriteDefaults` is called (which currently writes the file when missing).
- [x] 1.3 Create `internal/firstrun/wizard.go` with `BuildSteps(ctx) (steps []tui.WizardStep, appliers []func(tui.StepResult), needsConfigReload bool)`. The context struct carries home, projectDir, registered agents, registered kits, current `*config.Config`, and the first-run flag.
- [x] 1.4 Create `internal/firstrun/config_writer.go` with `BuildConfig`/`WriteConfig` that emits a complete config file reflecting wizard choices — active agents/kits uncommented, deselected ones commented (with original formatting preserved when the user's choice matches the kit-authored default). Existing `config.SetAgentIsolation` / `config.SetKitCredentials` are reused unchanged for the isolation/credentials appliers.
- [x] 1.5 Update `internal/firstrun/firstrun.go` / new `wizard.go` to orchestrate: build steps, present via `tui.Wizard`, apply completed results, return an `Outcome` struct flagging whether the config was written (image-shaping change) or only isolation/credentials were touched.
- [x] 1.6 Move the inline onboarding wizard out of `cmd/asylum/main.go` and into the new `firstrun` package; delete the helper from `main.go`.

## 2. Wizard steps

- [x] 2.1 Welcome banner — `log.Info("Welcome to asylum — let's set up your sandbox.")` printed on first-run before the wizard. Not a `WizardStep`.
- [x] 2.2 Agent multi-select step builder: list `agent.AllInstallNames()` filtering `echo`, pre-check `claude`. Skipped when not first-run.
- [x] 2.3 Default-agent single-select step builder. Pre-selects `claude` if among picks, else first picked. (Step is always added on first-run for wizard simplicity; the applier collapses to a no-op when only one agent is picked.)
- [x] 2.4 Kit multi-select step builder: top-level kits (no `/` in `Name`) excluding `TierAlwaysOn` and `Hidden`, pre-check `TierDefault`. Skipped when not first-run.
- [x] 2.5 Isolation step builder: only for claude when `cfg.AgentIsolation("claude") == ""`. Pre-selects `Shared with host (recommended)`; `Isolated` and `Project-isolated` follow.
- [x] 2.5a Flipped the implicit isolation fallback. `agent.ResolveConfigDir` default branch now returns the native host config dir; `cmd/asylum/main.go` switch reordered (explicit `isolated`/`project` cases, `default` = shared/no-seed); `container.go` `.agents` mount fires for `"shared"` or `""`.
- [x] 2.6 Credentials step builder ported from the old `runOnboarding` — same gating (any active credential-capable kit whose parent kit has unconfigured credentials).
- [x] 2.7 Appliers: kit applier writes the full config via `firstrun.WriteConfig` (flips `Outcome.WroteConfig`); isolation/credentials appliers use the existing `config.Set*` helpers and update in-memory cfg.

## 3. main.go reordering

- [x] 3.1 `firstrun.Run(...)` now runs immediately after `config.Load` and before `ensureImages`. The post-image-build inline wizard call is gone.
- [x] 3.2 When `Outcome.WroteConfig` is true, `config.Load` is re-invoked so the merged layer (with chosen agents/kits) drives image generation.
- [x] 3.3 The follow-on `kit.Resolve`, `resolveKitTiers`, `agent.ResolveInstalls`, `cacheDirs`, `a, _ := agent.Get(...)` resolutions sit after the optional reload, so they all use the post-wizard config.
- [x] 3.4 `firstrun.IsExistingInstall(home)` remains as the existing-user probe for the resume-migration dialog only; the new first-run wizard uses `firstrun.IsFirstRun`.

## 4. Image build context line

- [x] 4.1 Added `image.announceBuild()` driven by a `sync.Once` in `internal/image/image.go`.
- [x] 4.2 Called from both `EnsureBase` and `EnsureProject` just before their per-image `log.Build("building ... image...")` line — only when the cache-hit short-circuit was already passed.
- [x] 4.3 Cache-hit paths return before the `announceBuild` call, so the line stays suppressed when both images are up to date.

## 5. Silent SSH key generation

- [x] 5.1 `ssh-keygen` stdout/stderr captured into a `bytes.Buffer`; on non-zero exit the buffer content is folded into the returned error.
- [x] 5.2 The "SSH public key:" print and "Add this key to your Git host" line are gone.
- [x] 5.3 Replaced with a single `log.Info("Generated SSH key at %s.pub — see asylum-reference.md for usage.", keyPath)`.
- [x] 5.4 `internal/kit/ssh_test.go::TestEnsureSSHKey_SilentOnSuccess` captures stdout, asserts the absence of `ssh-keygen`'s banner / randomart / public-key block, asserts the new one-liner pointing at `asylum-reference.md` is present, and confirms the keypair lands on disk. Skips when `ssh-keygen` is not on `$PATH` so the test stays portable.

## 6. asylum-reference.md SSH section

- [x] 6.1 Extended SSH section: host commands to display the key in both isolated and project mode, Git-provider UI paths, and a "you may replace with your own keys" note.
- [x] 6.2 Existing isolation table kept; new guidance is additive.

## 7. Config writer

- [x] 7.1 `firstrun.WriteConfig(path, Choices)` + `BuildConfig(Choices)` written in `internal/firstrun/config_writer.go`. Generates the full file from scratch — preserves authored kit formatting when the user's selection matches the kit's default, otherwise applies a `# `-prefix transform per line.
- [x] 7.2 Hidden + `TierAlwaysOn` kits are not surfaced in the picker (handled by the wizard) and are emitted verbatim by the writer.

## 8. Tests

- [x] 8.1 `internal/firstrun/firstrun_test.go` covers `IsFirstRun` in both states (config.yaml absent and present).
- [x] 8.2 `internal/firstrun/wizard_test.go` covers: `activeAgentIsClaude` truth table; `credentialKitsFrom` filtering; `pickedFromMulti` empty-fallback; `resolveDefaultAgent` single-pick / multi-pick / invalid-pick fallback paths; `defaultKitChoices` includes TierDefault and excludes non-selectable kits; Codex adversarial findings have explicit regression tests (`TestRegression_NonClaudeSelectionSkipsClaudeIsolation`, `TestRegression_DeselectedCredentialKitStaysCommented`); and `Run` is no-op-equivalent on non-TTY existing-user invocations.
- [x] 8.3 `internal/firstrun/config_writer_test.go` covers: output parses as valid YAML, top-level `agent:` reflects choice, agents block groups active before commented and filters `echo`, selected/deselected kits land in the right block, non-selectable kits pass through verbatim, kit-authored formatting is preserved when the user's choice matches the kit's default, and `WriteConfig` round-trips through the filesystem.
- [x] 8.4 covered by 5.4.
- [x] 8.5 `container_test.go::TestRunArgsSandboxRulesMount` updated to set explicit `isolated` isolation now that the default is `shared`. `agent_test.go::TestResolveConfigDir` expectation for the empty-isolation case updated to the host's native config dir.

## 9. Documentation

- [x] 9.1 New `docs/concepts/first-run.md` describes the trigger, the step-by-step flow with defaults table, the resulting config layout, and the SSH-key behavior. Linked from `docs/concepts/index.md`. `docs/concepts/isolation.md` updated for the `shared`-default flip and now links to the first-run page.
- [x] 9.2 `docs/configuration/flags.md` reviewed — no change needed. The wizard does not add CLI surface, and the existing flag table remains accurate.
- [x] 9.3 `CHANGELOG.md` under `## Unreleased` updated with four bullets covering wizard, isolation flip, silent ssh, and pre-build context line.

## 10. Cleanup

- [x] 10.1 `firstrun.IsExistingInstall` retained only for the resume-migration prompt path. The wizard uses `firstrun.IsFirstRun`.
- [x] 10.2 `SyncNewKits` unchanged — it continues to fire only for existing users on upgrade; first-run users go through the wizard and reach `SyncNewKits` with an already-written config so it is a no-op.
- [x] 10.3 `go build ./...`, `go vet ./...`, and `go test ./...` (588 passed in 15 packages) all clean.
