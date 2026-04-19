## Context

Project config is loaded in `internal/config/config.go` at `Load()`, which iterates a hardcoded slice of two paths (`.asylum`, `.asylum.local`) per project layer. Each path goes through `NeedsMigration` → `MigrateV1ToV2` → `LoadFile` → `Merge`. The migration writes back to whatever path it received, so it is filename-agnostic.

The global config (`~/.asylum/config.yaml`) already uses a `.yaml` extension. The asymmetry exists only at the project layer because dotfiles without extensions are conventional (`.gitignore`, `.dockerignore`, `.envrc`).

Issue #15 originally proposed a hard rename. The maintainer wants a softer change: allow the alias, keep the canonical name. This avoids the cost of a forced rename (every existing repo, every doc reference) while unlocking the editor experience for users who want it.

## Goals / Non-Goals

**Goals:**
- Project and local config layers each accept either the canonical filename or its `.yaml`-suffixed alias.
- Both-present condition produces a clear error rather than silently picking one.
- Behavior is identical regardless of which filename is used (migration, merging, error messages).
- `.asylum` remains the canonical name in scaffolding, examples, and primary documentation.

**Non-Goals:**
- Renaming `.asylum` to `.asylum.yaml` as the default.
- Switching documentation examples to `.yaml` extensions.
- Adding more aliases (e.g. `.yml`, `asylum.yaml`) — keep the surface minimal.
- Touching the global config path (already `.yaml`).
- Auto-migrating existing `.asylum` files to `.asylum.yaml`.

## Decisions

### Per-layer alias resolution

Replace the hardcoded slice with a small helper that, given a project dir and a `(canonical, alias)` pair, returns the path to load (or empty if neither exists, or an error if both exist).

```go
// resolveProjectConfigPath returns the path that should be loaded for a layer,
// preferring no preference: if both names exist for the same layer, it returns an error.
func resolveProjectConfigPath(projectDir, canonical, alias string) (string, error)
```

Call sites in `Load()` change from iterating `[".asylum", ".asylum.local"]` to iterating `[(".asylum", ".asylum.yaml"), (".asylum.local", ".asylum.local.yaml")]`. If `resolveProjectConfigPath` returns an empty string (neither file exists), skip that layer just like today's missing-file behavior.

**Why a helper, not inline code?** Two reasons. First, the both-present check is non-trivial enough that inlining it twice would be repetitive. Second, the helper is the natural unit to test — a unit test on the helper covers all branches without needing to set up the full `Load()` machinery.

### Both-present is an error

If a layer has both `.asylum` and `.asylum.yaml`, return an error like:

```
both .asylum and .asylum.yaml exist in <projectDir>; remove one
```

**Why error instead of pick-one?** Silent precedence (e.g. "canonical wins") would let users keep two diverging configs and not realize one was being ignored. The likely cause of both-present is a half-finished rename, and an error makes that obvious. Same justification covers `.asylum.local` / `.asylum.local.yaml`.

**Why not merge them?** Merging within a single layer breaks the layered-config mental model (global → project → local → CLI) — there's no spec-defined precedence between two files at the same layer.

### `.yaml` only — no `.yml`

The global config already uses `.yaml`, not `.yml`. Sticking to one spelling keeps the surface minimal. If a user requests `.yml` later, it's an additive change.

### Documentation placement

The alias note goes wherever project config is first introduced:
- `docs/configuration/index.md` — the canonical reference; mention as an admonition or short paragraph after the table of layers.
- `docs/getting-started.md` — one-liner where `.asylum` is first introduced.
- `assets/asylum-reference.md` — one-liner under the project config bullet.
- `README.md` — only if the existing reference to `.asylum` naturally accommodates a parenthetical; do not force it.

`CHANGELOG.md` gets an entry under Added.

## Risks / Trade-offs

- **[Two filenames coexisting in real repos]** → Both-present check + error message catches it loudly the first time the user runs asylum.
- **[Users assume the rename happened and rename their files]** → Acceptable. Asylum loads `.asylum.yaml` correctly, so the user's repo just works under the new name. We're not deprecating the canonical name; both stay valid indefinitely.
- **[Editor schema validation]** → Not in scope. Adding a `.yaml` extension makes editors *treat the file as YAML*, which is the issue's stated goal. Schema-based validation would require shipping a JSON schema and wiring it up — separate work, separate change if requested.
- **[Hidden cost: bumping spec language]** → The `config-loading` spec currently lists the exact filenames. The delta updates that requirement — minor wording change, no behavior change beyond the alias.
