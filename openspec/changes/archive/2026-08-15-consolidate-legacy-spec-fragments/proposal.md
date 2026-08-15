## Why

Five capability specs are single-requirement fragments left over from the profiles→kits rename: each is a `MODIFIED` delta that was merged into a directory of its own instead of into the capability it actually modified. They duplicate requirements that live elsewhere, and two of them are now wrong:

- `profile-container-setup` labels `isolated` as the default agent config isolation mode. The default is `shared` — `config-isolation`'s own "Implicit isolation fallback" requirement says so, and `ResolveConfigDir` implements it (`internal/agent/agent.go:63`).
- `profile-system` specifies a `DefaultOn` field on the `Kit` struct. That field no longer exists; it was replaced by `Tier`. The same requirement lists 10 struct fields where the real struct has 28.

A reader looking up how isolation defaults or how a kit is declared finds two specs disagreeing, with no way to tell which is current. The formatting pass in `normalize-spec-format` deliberately left this alone because it changes requirement text.

A third defect surfaced while scoping this change: `first-run-onboarding` states that a non-interactive first run applies "isolated config". Nothing writes an isolation value on that path — it is only written by the interactive wizard (`internal/firstrun/wizard.go:215`) or `asylum config` (`cmd/asylum/config.go:145`) — so the `shared` fallback applies. It is the same wrong claim as `profile-container-setup`, in a different spec.

## What Changes

Each fragment is either folded into the capability that owns its subject and then deleted, or deleted outright as a duplicate. Nothing unique to a fragment is lost:

- **`agent-install` → `config-isolation`, then deleted.** Its requirement covers how `agents.<agent>.config` parses, including the empty value that triggers the first-run prompt. That is the same subject as `config-isolation`.
- **`profile-config-integration` → `kit-defaults`, then deleted.** Mostly duplicated by "Kit disabling", but it uniquely specifies that a project layer may set `disabled: false` to re-enable a kit disabled globally. That scenario moves; the rest is dropped as duplicate.
- **`profile-image-build` → `host-user-alignment`, then deleted.** The `USER_HOME` build argument is already required by "Container home directory matches host". Its unique content — the base image hash including the home directory path, so a different host home triggers a rebuild — moves as a new requirement.
- **`profile-container-setup` deleted, nothing moved.** Its three scenarios duplicate `config-isolation`'s "Config isolation levels" mount-path scenarios exactly. Its only distinguishing content is the incorrect `(default)` label.
- **`profile-system` deleted, nothing moved.** `Deps` validation is specified by `kit-dependencies`, activation by `kit-defaults` "Kit activation tier". What remains is a stale field enumeration naming a field that no longer exists.
- **`first-run-onboarding` corrected** so its non-interactive scenario describes the `shared` fallback rather than "isolated config".

Deliberately out of scope: renaming `profile-entrypoint` (the last legacy-named capability, but one with three real requirements), and writing a spec for the agent install system, which has never had one.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none in the OpenSpec delta sense — this change edits main specs in place rather than proposing new behavior, so `.openspec.yaml` sets `skip_specs: true`. No system behavior changes; the specs are corrected to match what the code already does.)

## Impact

- Deleted: `openspec/specs/agent-install/`, `profile-config-integration/`, `profile-container-setup/`, `profile-image-build/`, `profile-system/`. Capability count drops from 79 to 74.
- Edited: `openspec/specs/config-isolation/spec.md`, `kit-defaults/spec.md`, `host-user-alignment/spec.md`, `first-run-onboarding/spec.md`.
- No Go source, no assets, no docs site, no CHANGELOG entry. No behavior change — the code is already correct; the specs were not.
- After this change the agent install system has no spec of its own. It had none before either: the `agent-install` capability described a config field, not the install system.
