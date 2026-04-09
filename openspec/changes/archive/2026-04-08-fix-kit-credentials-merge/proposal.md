## Why

`mergeKitConfig` was introduced by the `deep-merge-kit-config` change but never updated to handle the `Credentials` field added shortly after by `kit-credentials`. As a result, credential configuration in overlay config files (`.asylum.local`, project `.asylum`) was silently discarded during merge, making it impossible to set or override credentials per-project.

## What Changes

- `mergeKitConfig` now carries the `Credentials` field from the overlay when non-nil (last-wins, matching the semantics of all other scalar fields)
- No config format changes, no new fields, no behavioral changes for users who don't rely on overlay configs for credentials

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `deep-merge-kit-config`: The field-level merge rules for `KitConfig` now include `Credentials` with last-wins semantics, closing the gap left when `kit-credentials` added the field after `mergeKitConfig` was written.

## Impact

- `internal/config/config.go` — one-line fix in `mergeKitConfig`
- No API changes, no config format changes, no migration needed
