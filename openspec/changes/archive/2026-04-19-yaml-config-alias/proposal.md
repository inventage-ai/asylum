## Why

Project config files (`.asylum`, `.asylum.local`) have no extension, so editors don't apply YAML syntax highlighting, validation, or schema-based autocompletion. Renaming them outright would churn every existing user's repo. Accepting `.yaml`-suffixed aliases lets users opt in to better editor support without forcing migration. Resolves #15.

## What Changes

- Project config loader accepts either `.asylum` or `.asylum.yaml` for the project layer, and either `.asylum.local` or `.asylum.local.yaml` for the local layer.
- If both the canonical name and the alias exist for the same layer, asylum SHALL error out — silent merging or arbitrary preference would surprise the user.
- `.asylum` remains the canonical default in all examples, scaffolding, and primary documentation. The `.yaml` variants are documented as opt-in aliases.
- Documentation (`docs/configuration/index.md`, `docs/getting-started.md`, `assets/asylum-reference.md`, `README.md`) gains a brief note explaining the alias.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `config-loading`: project and local config layers SHALL accept either the canonical filename or its `.yaml`-suffixed alias, with both-present treated as an error.

## Impact

- `internal/config/config.go` — replace the hardcoded path slice with per-layer alias resolution; emit error when both names coexist.
- `internal/config/config_test.go` — coverage for alias loading, both-present error, and mixed canonical/alias across layers.
- `docs/configuration/index.md`, `docs/getting-started.md`, `assets/asylum-reference.md`, `README.md` — short note on the `.yaml` alias.
- No code change touches migration, scaffolding, or runtime behavior beyond which file is read.
