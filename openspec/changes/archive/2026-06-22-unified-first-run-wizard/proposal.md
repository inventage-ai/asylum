## Why

Today's first-run experience surfaces a slow image build before any user interaction, asks settings questions afterward, and ends with raw `ssh-keygen` output spilled into the terminal just before the agent launches. New users get no chance to pick which agent(s) or kits they want — asylum silently writes the defaults and bakes them into the image, so swapping to Gemini or enabling the Java kit means knowing about config files. The result reads as opaque on the first run and noisy at every stage.

## What Changes

- **Unified first-run wizard runs before any image build.** A new user is greeted, asked which agents they want, which top-level kits to activate, then the existing isolation and credentials questions — all in one `tui.Wizard` flow. Pressing enter through every step yields today's silent defaults (claude only, `TierDefault` kits, isolated config, no credentials), so the fast path is one keypress per step.
- **Multi-select agents.** Users can pick multiple agents to bake into the image; a follow-up single-select picks the default agent when more than one is chosen. The `echo` test stub is hidden from the picker but still selectable via `-a echo`.
- **Top-level kit picker.** Multi-select listing top-level kits with `TierDefault` pre-checked and `TierAvailable` unchecked. Sub-kits (`java/maven`, `python/pip`) are not exposed in the wizard — they remain config-file territory. `TierAlwaysOn` kits stay always-on and out of the picker.
- **Wizard ownership moves into the `firstrun` package.** `internal/firstrun/` becomes the home for the wizard build + result application; `main.go` calls `firstrun.Run(...)` once before `ensureImages`. The existing in-`main.go` `runOnboarding` collapses into the same package.
- **Context line before image builds.** When `EnsureBase` or `EnsureProject` is going to actually build (not a cache hit), asylum prints a single user-facing line ("Building sandbox image — this takes a few minutes the first time, subsequent runs reuse the cache") before the build log lines, so users on first-run don't stare at silence.
- **Silent SSH key generation.** `ssh-keygen`'s stdout/stderr is captured and dropped on success. The "Add this key to your Git hosting provider" sentence is removed. Asylum prints one line on key creation: `Generated SSH key at ~/.asylum/ssh/id_ed25519.pub — see asylum-reference.md for usage`.
- **SSH usage details moved to `asylum-reference.md`.** The existing SSH section in `assets/asylum-reference.md` is extended with the "how to add this key to a Git host" guidance, so the in-container agent can answer the question instead of asylum spamming on every fresh container start.
- **First-run detection switches signal.** From "does `~/.asylum/agents/` exist" to "does `~/.asylum/config.yaml` exist before `WriteDefaults` runs". The latter is the direct signal — it's exactly what `WriteDefaults` checks — and survives across the migration prompt rework that already shipped.
- **Default agent config isolation flips from `isolated` to `shared`.** The wizard's isolation step pre-selects `shared`, and the implicit default applied when no config value is set (interactive defaults / non-interactive runs) becomes `shared`. The "(recommended)" tag moves from `Isolated` to `Shared with host`. Rationale: most users expect their host login session, MCP servers, and plugins to "just work" inside the sandbox; isolated mode surprises them by hiding host state.

Not a breaking change for non-interactive flows: when stdin is not a TTY, the wizard is skipped entirely and today's silent defaults apply.

## Capabilities

### New Capabilities

None — this change reorganizes and extends existing capabilities rather than introducing new ones.

### Modified Capabilities

- `first-run-onboarding`: First-run signal changes from `~/.asylum/agents/` to `~/.asylum/config.yaml` absence; the package becomes the owner of the full first-run wizard rather than a future-task shell.
- `onboarding-wizard`: Wizard gains two new steps (agents, kits) on first run; runs *before* `ensureImages` rather than after. Existing isolation/credentials steps continue to fire whenever they're unconfigured.
- `ssh-kit`: First-run isolated/project key generation becomes silent — `ssh-keygen` output is captured, the public-key-plus-instructions print block is replaced with a single line pointing at `asylum-reference.md`.
- `image-build`: Adds a single user-facing context line before any actual base or project rebuild (suppressed on cache hits).
- `sandbox-rules`: SSH section in `asylum-reference.md` extended with the "how to use this generated key with a Git host" guidance that previously lived in `ssh-keygen` output.
- `config-isolation`: Default isolation level flips from `isolated` to `shared`. Affects the wizard pre-selection and the implicit fallback when no value is set.

## Impact

- Affected code:
  - `cmd/asylum/main.go` — remove inline `runOnboarding`, call `firstrun.Run` earlier, drop `firstrun.IsExistingInstall`/old `firstrun.Run` shell.
  - `internal/firstrun/` — new `wizard.go` (step builders), `writer.go` (config persistence helpers); existing `firstrun.go` becomes the real entry point; `migration.go` stays as-is.
  - `internal/kit/ssh.go` — capture `ssh-keygen` output; replace the print block with a one-liner emitted via the project's `log` package.
  - `internal/image/image.go` — emit a context line before `docker build` is actually invoked.
  - `assets/asylum-reference.md` — extend the SSH section with the prior "Add this key to your Git hosting provider" guidance.
- No external APIs change. CLI flag surface unchanged.
- Non-interactive mode (no TTY) preserves today's behavior: wizard skipped, defaults applied silently.
- `SyncNewKits` (the post-install "new kit available" prompt) is unaffected — it continues to fire for existing users on upgrade; first-run users go through the wizard and exit with config already up to date, so `SyncNewKits` is a no-op for them.
